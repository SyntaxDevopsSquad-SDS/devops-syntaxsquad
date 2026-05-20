package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func setupWeatherTest() {
	setupRoutesTest()
	weatherCacheMu.Lock()
	weatherCache = make(map[string]cachedWeather)
	weatherCacheMu.Unlock()
}

func mockGeoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"results":[{"name":"Copenhagen","latitude":55.6761,"longitude":12.5683}]}`)
	}))
}

func mockWeatherServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"current_weather":{"temperature":15.5,"windspeed":10.2,"weathercode":1},"hourly":{"temperature_2m":[15.5],"weathercode":[1]}}`)
	}))
}

func TestWeatherHandler(t *testing.T) {
	t.Run("Returns 200 without city param", func(t *testing.T) {
		setupWeatherTest()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/weather", nil)

		weatherHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for weather page without city, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "weather") {
			t.Error("Expected weather page body to contain 'weather'")
		}
	})

	t.Run("Returns 200 with city=Copenhagen", func(t *testing.T) {
		setupWeatherTest()

		geoSrv := mockGeoServer(t)
		defer geoSrv.Close()
		wxSrv := mockWeatherServer(t)
		defer wxSrv.Close()

		geocodingBaseURL = geoSrv.URL
		weatherBaseURL = wxSrv.URL
		defer func() {
			geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
			weatherBaseURL = "https://api.open-meteo.com/v1/forecast"
		}()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/weather?city=Copenhagen", nil)

		weatherHandler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for weather page with city, got %d", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Copenhagen") {
			t.Errorf("Expected body to contain 'Copenhagen', got: %s", body)
		}
		if !strings.Contains(body, "15.5") {
			t.Errorf("Expected body to contain temperature '15.5', got: %s", body)
		}
	})

	t.Run("Cache returns cached data on second call", func(t *testing.T) {
		setupWeatherTest()

		callCount := 0
		geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"results":[{"name":"Copenhagen","latitude":55.6761,"longitude":12.5683}]}`)
		}))
		defer geoSrv.Close()
		wxSrv := mockWeatherServer(t)
		defer wxSrv.Close()

		geocodingBaseURL = geoSrv.URL
		weatherBaseURL = wxSrv.URL
		defer func() {
			geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
			weatherBaseURL = "https://api.open-meteo.com/v1/forecast"
		}()

		// First request – populates cache
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest(http.MethodGet, "/weather?city=Copenhagen", nil)
		weatherHandler(w1, r1)
		if w1.Code != http.StatusOK {
			t.Fatalf("First request: expected 200, got %d", w1.Code)
		}

		// Second request – should hit cache, not the geocoding server
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest(http.MethodGet, "/weather?city=Copenhagen", nil)
		weatherHandler(w2, r2)
		if w2.Code != http.StatusOK {
			t.Fatalf("Second request: expected 200, got %d", w2.Code)
		}

		if callCount != 1 {
			t.Errorf("Expected geocoding API to be called once (cache hit on second), got %d calls", callCount)
		}
	})

	t.Run("Cache invalidates after TTL", func(t *testing.T) {
		setupWeatherTest()

		// Pre-populate cache with a stale entry
		weatherCacheMu.Lock()
		weatherCache["Copenhagen"] = cachedWeather{
			data:      WeatherData{City: "Copenhagen", Temperature: 99.0},
			fetchedAt: time.Now().Add(-(weatherCacheTTL + time.Second)),
		}
		weatherCacheMu.Unlock()

		callCount := 0
		geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"results":[{"name":"Copenhagen","latitude":55.6761,"longitude":12.5683}]}`)
		}))
		defer geoSrv.Close()
		wxSrv := mockWeatherServer(t)
		defer wxSrv.Close()

		geocodingBaseURL = geoSrv.URL
		weatherBaseURL = wxSrv.URL
		defer func() {
			geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
			weatherBaseURL = "https://api.open-meteo.com/v1/forecast"
		}()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/weather?city=Copenhagen", nil)
		weatherHandler(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 after cache expiry, got %d", w.Code)
		}
		if callCount == 0 {
			t.Error("Expected geocoding API to be called after cache TTL expired, but it was not")
		}

		// Verify fresh data replaced the stale 99.0 temperature
		weatherCacheMu.RLock()
		updated := weatherCache["Copenhagen"]
		weatherCacheMu.RUnlock()
		if updated.data.Temperature == 99.0 {
			t.Error("Expected cache to be refreshed with new temperature, still has stale value 99.0")
		}
	})
}
