package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

// ─────────────────────────────── Config ──────────────────────────────────────

type Endpoint struct {
	Method string         `json:"method"`
	URL    string         `json:"url"`
	Body   map[string]any `json:"body,omitempty"`
}

type BotCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	ReportEveryS int `json:"report_every_s"`

	HTTP struct {
		Endpoints  []Endpoint `json:"endpoints"`
		BotCount   int        `json:"bot_count"`
		IntervalMs int        `json:"interval_ms"`
	} `json:"http"`

	DB struct {
		Driver     string `json:"driver"`
		DSN        string `json:"dsn"`
		BotCount   int    `json:"bot_count"`
		IntervalMs int    `json:"interval_ms"`
	} `json:"db"`

	UserBots struct {
		Enabled     bool            `json:"enabled"`
		BaseURL     string          `json:"base_url"`
		Credentials []BotCredential `json:"credentials"`

		// 10 bots: login → søg 2-3 gange med realistiske pauser → logout
		Normal struct {
			Count              int `json:"count"`
			SearchesPerSession int `json:"searches_per_session"`
			PauseMs            int `json:"pause_ms"`
			LoopDelayMs        int `json:"loop_delay_ms"`
		} `json:"normal"`

		// 5 bots: login → søg hurtigt og hyppigt → logout
		Heavy struct {
			Count              int `json:"count"`
			SearchesPerSession int `json:"searches_per_session"`
			PauseMs            int `json:"pause_ms"`
			LoopDelayMs        int `json:"loop_delay_ms"`
		} `json:"heavy"`

		// 5 bots: login → logout med det samme (stresser session-håndtering)
		Session struct {
			Count       int `json:"count"`
			LoopDelayMs int `json:"loop_delay_ms"`
		} `json:"session"`
	} `json:"user_bots"`

	LoadTest struct {
		Enabled    bool     `json:"enabled"`
		URLs       []string `json:"urls"`
		Concurrent int      `json:"concurrent"`
		TotalReqs  int      `json:"total_requests"`
	} `json:"load_test"`
}

func loadConfig(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("config: kan ikke åbne %s: %v", path, err)
	}
	defer f.Close()
	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("config: parse-fejl: %v", err)
	}
	return cfg
}

// ─────────────────────────────── Metrics ─────────────────────────────────────

type atomicCounter struct{ v int64 }

func (c *atomicCounter) Inc()        { atomic.AddInt64(&c.v, 1) }
func (c *atomicCounter) Swap() int64 { return atomic.SwapInt64(&c.v, 0) }

type latBuf struct {
	mu   sync.Mutex
	data []time.Duration
}

func (l *latBuf) Add(d time.Duration) {
	l.mu.Lock()
	l.data = append(l.data, d)
	l.mu.Unlock()
}

func (l *latBuf) Drain() []time.Duration {
	l.mu.Lock()
	out := make([]time.Duration, len(l.data))
	copy(out, l.data)
	l.data = l.data[:0]
	l.mu.Unlock()
	return out
}

type Metrics struct {
	// Endpoint bots
	httpOK, httpFail atomicCounter
	httpLat          latBuf

	// DB bot
	dbOK, dbFail atomicCounter
	dbLat        latBuf

	// User bots
	userLoginOK, userLoginFail     atomicCounter
	userSearchOK, userSearchFail   atomicCounter
	userSessionOK, userSessionFail atomicCounter

	// Load test
	loadOK, loadFail atomicCounter
	loadLat          latBuf
}

// ─────────────────────────────── Helpers ─────────────────────────────────────

func avgMs(ds []time.Duration) float64 {
	if len(ds) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range ds {
		sum += d
	}
	return float64(sum) / float64(len(ds)) / float64(time.Millisecond)
}

func uptimePct(ok, fail int64) float64 {
	if ok+fail == 0 {
		return 100.0
	}
	return float64(ok) / float64(ok+fail) * 100
}

func statusIcon(ok, fail int64) string {
	if ok+fail == 0 {
		return "⚪"
	}
	switch pct := uptimePct(ok, fail); {
	case pct >= 99:
		return "🟢"
	case pct >= 90:
		return "🟡"
	default:
		return "🔴"
	}
}

// ─────────────────────────────── HTTP Bot ────────────────────────────────────

func httpBot(ctx context.Context, id int, cfg Config, m *Metrics) {
	if len(cfg.HTTP.Endpoints) == 0 {
		log.Printf("[HTTP-%d] ingen endpoints – afslutter", id)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(time.Duration(cfg.HTTP.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	ep := cfg.HTTP.Endpoints[id%len(cfg.HTTP.Endpoints)]

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			var resp *http.Response
			var err error

			if ep.Method == "POST" {
				body, _ := json.Marshal(ep.Body)
				resp, err = client.Post(ep.URL, "application/json", bytes.NewReader(body))
			} else {
				resp, err = client.Get(ep.URL)
			}

			lat := time.Since(start)

			if err != nil || resp == nil {
				m.httpFail.Inc()
				log.Printf("[HTTP-%d] ❌ %s – %v", id, ep.URL, err)
				continue
			}
			resp.Body.Close()
			m.httpLat.Add(lat)

			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				m.httpOK.Inc()
			} else {
				m.httpFail.Inc()
				log.Printf("[HTTP-%d] ⚠️  %s → HTTP %d", id, ep.URL, resp.StatusCode)
			}
		}
	}
}

// ─────────────────────────────── DB Bot ──────────────────────────────────────

func dbBot(ctx context.Context, id int, cfg Config, m *Metrics) {
	db, err := sql.Open(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		log.Printf("[DB-%d] ❌ open-fejl: %v", id, err)
		m.dbFail.Inc()
		return
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Second)

	ticker := time.NewTicker(time.Duration(cfg.DB.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			start := time.Now()
			if err := db.PingContext(ctx); err != nil {
				m.dbFail.Inc()
				log.Printf("[DB-%d] ❌ ping-fejl: %v", id, err)
			} else {
				m.dbLat.Add(time.Since(start))
				m.dbOK.Inc()
			}
		}
	}
}

// ─────────────────────────────── User Bots ───────────────────────────────────

// Rammer begge SQL-grene i searchHandler:
// ≤2 tegn → LIKE-fallback | >2 tegn → tsvector full-text search
var searchTerms = []string{
	"linux", "docker", "python", "devops", "kubernetes",
	"nginx", "golang", "ansible", "terraform", "ubuntu",
	"postgresql", "bash", "ssh", "firewall", "github",
	"container", "security", "monitoring", "prometheus", "grafana",
	"go", "ci", "cd", // ≤2 tegn – LIKE-fallback
}

type botKind int

const (
	kindNormal  botKind = iota // rolig bruger
	kindHeavy                  // hurtig søger
	kindSession                // login/logout-stresser
)

func newBotClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
		// Stop efter 3 redirects – fanger login-redirects som success
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func botLogin(ctx context.Context, client *http.Client, baseURL string, cred BotCredential, m *Metrics) bool {
	form := url.Values{
		"username":   {cred.Username},
		"password":   {cred.Password},
		"csrf_token": {"sim-bot"}, // ignoreres med CSRF_RELAXED=true
	}
	req, err := http.NewRequestWithContext(ctx, "POST",
		baseURL+"/api/login", strings.NewReader(form.Encode()))
	if err != nil {
		m.userLoginFail.Inc()
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil || resp == nil {
		m.userLoginFail.Inc()
		return false
	}
	resp.Body.Close()

	// Login returnerer redirect (302) til / ved success
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		m.userLoginOK.Inc()
		return true
	}
	m.userLoginFail.Inc()
	return false
}

func botSearch(ctx context.Context, client *http.Client, baseURL, term string, m *Metrics) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		baseURL+"/api/search?q="+url.QueryEscape(term)+"&language=en", nil)
	if err != nil {
		m.userSearchFail.Inc()
		return
	}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		m.userSearchFail.Inc()
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		m.userSearchOK.Inc()
	} else {
		m.userSearchFail.Inc()
	}
}

func botLogout(ctx context.Context, client *http.Client, baseURL string, m *Metrics) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/logout", nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		m.userSessionOK.Inc()
	} else {
		m.userSessionFail.Inc()
	}
}

func userBot(ctx context.Context, id int, kind botKind, cfg Config, m *Metrics) {
	ub := cfg.UserBots
	if len(ub.Credentials) == 0 {
		log.Printf("[UserBot-%d] ingen credentials – afslutter", id)
		return
	}

	cred := ub.Credentials[id%len(ub.Credentials)]
	kindName := map[botKind]string{kindNormal: "normal", kindHeavy: "heavy", kindSession: "session"}
	log.Printf("[UserBot-%d] starter som '%s' (%s)", id, cred.Username, kindName[kind])

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		client := newBotClient() // frisk cookie jar per session

		if !botLogin(ctx, client, ub.BaseURL, cred, m) {
			log.Printf("[UserBot-%d] ❌ login fejlede – venter 10s", id)
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			continue
		}

		switch kind {
		case kindNormal:
			for i := 0; i < ub.Normal.SearchesPerSession; i++ {
				botSearch(ctx, client, ub.BaseURL, searchTerms[rand.Intn(len(searchTerms))], m)
				select {
				case <-ctx.Done():
					botLogout(ctx, client, ub.BaseURL, m)
					return
				case <-time.After(time.Duration(ub.Normal.PauseMs) * time.Millisecond):
				}
			}

		case kindHeavy:
			for i := 0; i < ub.Heavy.SearchesPerSession; i++ {
				botSearch(ctx, client, ub.BaseURL, searchTerms[rand.Intn(len(searchTerms))], m)
				select {
				case <-ctx.Done():
					botLogout(ctx, client, ub.BaseURL, m)
					return
				case <-time.After(time.Duration(ub.Heavy.PauseMs) * time.Millisecond):
				}
			}

		case kindSession:
			// Ingen handling – bare login + logout
		}

		botLogout(ctx, client, ub.BaseURL, m)

		delayMs := map[botKind]int{
			kindNormal:  ub.Normal.LoopDelayMs,
			kindHeavy:   ub.Heavy.LoopDelayMs,
			kindSession: ub.Session.LoopDelayMs,
		}[kind]

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(delayMs) * time.Millisecond):
		}
	}
}

// ─────────────────────────────── Load Test ───────────────────────────────────

func runLoadTest(cfg Config, m *Metrics) {
	if !cfg.LoadTest.Enabled || len(cfg.LoadTest.URLs) == 0 {
		return
	}

	fmt.Printf("\n🚀 Load test: %d workers · %d requests · %d URL(s)\n",
		cfg.LoadTest.Concurrent, cfg.LoadTest.TotalReqs, len(cfg.LoadTest.URLs))
	for _, u := range cfg.LoadTest.URLs {
		fmt.Printf("   → %s\n", u)
	}

	sem := make(chan struct{}, cfg.LoadTest.Concurrent)
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 10 * time.Second}
	urlHits := make([]atomicCounter, len(cfg.LoadTest.URLs))
	start := time.Now()

	for i := 0; i < cfg.LoadTest.TotalReqs; i++ {
		idx := i % len(cfg.LoadTest.URLs)
		wg.Add(1)
		sem <- struct{}{}

		go func(urlIdx int) {
			defer wg.Done()
			defer func() { <-sem }()

			t := time.Now()
			resp, err := client.Get(cfg.LoadTest.URLs[urlIdx])
			lat := time.Since(t)

			if err != nil || resp == nil {
				m.loadFail.Inc()
				return
			}
			resp.Body.Close()
			m.loadLat.Add(lat)

			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				m.loadOK.Inc()
				urlHits[urlIdx].Inc()
			} else {
				m.loadFail.Inc()
			}
		}(idx)
	}

	wg.Wait()
	elapsed := time.Since(start)
	lats := m.loadLat.Drain()
	ok, fail := m.loadOK.Swap(), m.loadFail.Swap()

	fmt.Printf("──────────────────────────────────────────────\n")
	fmt.Printf("  ✅ Load test færdig på %.2fs\n", elapsed.Seconds())
	fmt.Printf("  OK:%-6d  Fejl:%-6d  RPS:%.0f  Avg lat:%.1f ms\n",
		ok, fail, float64(ok+fail)/elapsed.Seconds(), avgMs(lats))
	fmt.Printf("  Per URL:\n")
	for i, u := range cfg.LoadTest.URLs {
		fmt.Printf("    %-4d %s\n", urlHits[i].Swap(), u)
	}
	fmt.Printf("──────────────────────────────────────────────\n\n")
}

// ─────────────────────────────── Reporter ────────────────────────────────────

func reporter(ctx context.Context, cfg Config, m *Metrics) {
	ticker := time.NewTicker(time.Duration(cfg.ReportEveryS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			hOK, hFail := m.httpOK.Swap(), m.httpFail.Swap()
			dOK, dFail := m.dbOK.Swap(), m.dbFail.Swap()
			lnOK, lnFail := m.userLoginOK.Swap(), m.userLoginFail.Swap()
			srOK, srFail := m.userSearchOK.Swap(), m.userSearchFail.Swap()
			ssOK, ssFail := m.userSessionOK.Swap(), m.userSessionFail.Swap()
			hLats := m.httpLat.Drain()
			dLats := m.dbLat.Drain()

			sep := "══════════════════════════════════════════════════════"
			fmt.Printf("\n╔%s╗\n", sep)
			fmt.Printf("║  📊 Rapport – %-39s║\n", t.Format("15:04:05"))
			fmt.Printf("╠%s╣\n", sep)
			fmt.Printf("║  %s HTTP    │ OK:%-5d Fejl:%-5d Uptime:%5.1f%% Lat:%5.1f ms ║\n",
				statusIcon(hOK, hFail), hOK, hFail, uptimePct(hOK, hFail), avgMs(hLats))
			fmt.Printf("╠%s╣\n", sep)
			fmt.Printf("║  %s DB      │ OK:%-5d Fejl:%-5d Uptime:%5.1f%% Lat:%5.1f ms ║\n",
				statusIcon(dOK, dFail), dOK, dFail, uptimePct(dOK, dFail), avgMs(dLats))
			fmt.Printf("╠%s╣\n", sep)
			fmt.Printf("║  👤 Login   │ OK:%-5d Fejl:%-5d Uptime:%5.1f%%              ║\n",
				lnOK, lnFail, uptimePct(lnOK, lnFail))
			fmt.Printf("║  🔍 Søgning │ OK:%-5d Fejl:%-5d Uptime:%5.1f%%              ║\n",
				srOK, srFail, uptimePct(srOK, srFail))
			fmt.Printf("║  🔓 Session │ OK:%-5d Fejl:%-5d Uptime:%5.1f%%              ║\n",
				ssOK, ssFail, uptimePct(ssOK, ssFail))
			fmt.Printf("╚%s╝\n", sep)
		}
	}
}

// ──────────────────────────────── Main ───────────────────────────────────────

func main() {
	cfgPath := "config.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg := loadConfig(cfgPath)
	m := &Metrics{}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// HTTP bots
	for i := 0; i < cfg.HTTP.BotCount; i++ {
		wg.Add(1)
		go func(id int) { defer wg.Done(); httpBot(ctx, id, cfg, m) }(i)
	}
	log.Printf("▶ %d HTTP-bot(s) startet", cfg.HTTP.BotCount)

	// DB bot
	if cfg.DB.DSN != "" && cfg.DB.BotCount > 0 {
		for i := 0; i < cfg.DB.BotCount; i++ {
			wg.Add(1)
			go func(id int) { defer wg.Done(); dbBot(ctx, id, cfg, m) }(i)
		}
		log.Printf("▶ %d DB-bot(s) startet", cfg.DB.BotCount)
	}

	// User bots
	if cfg.UserBots.Enabled && len(cfg.UserBots.Credentials) > 0 {
		ub := cfg.UserBots
		offset := 0

		for i := 0; i < ub.Normal.Count; i++ {
			wg.Add(1)
			go func(id int) { defer wg.Done(); userBot(ctx, id, kindNormal, cfg, m) }(offset + i)
		}
		offset += ub.Normal.Count

		for i := 0; i < ub.Heavy.Count; i++ {
			wg.Add(1)
			go func(id int) { defer wg.Done(); userBot(ctx, id, kindHeavy, cfg, m) }(offset + i)
		}
		offset += ub.Heavy.Count

		for i := 0; i < ub.Session.Count; i++ {
			wg.Add(1)
			go func(id int) { defer wg.Done(); userBot(ctx, id, kindSession, cfg, m) }(offset + i)
		}

		log.Printf("▶ %d user-bot(s): %d normal · %d heavy · %d session",
			ub.Normal.Count+ub.Heavy.Count+ub.Session.Count,
			ub.Normal.Count, ub.Heavy.Count, ub.Session.Count)
	}

	// Reporter
	wg.Add(1)
	go func() { defer wg.Done(); reporter(ctx, cfg, m) }()

	// Load test (one-shot)
	runLoadTest(cfg, m)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Lukker ned...")
	cancel()
	wg.Wait()
	fmt.Println("✅ Alle bots stoppet.")
}
