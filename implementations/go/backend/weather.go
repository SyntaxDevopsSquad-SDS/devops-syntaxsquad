package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
	weatherBaseURL   = "https://api.open-meteo.com/v1/forecast"
)

type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type openMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		Windspeed   float64 `json:"windspeed"`
		Weathercode int     `json:"weathercode"`
	} `json:"current_weather"`
	Hourly struct {
		Temperature2m []float64 `json:"temperature_2m"`
		Weathercode   []int     `json:"weathercode"`
	} `json:"hourly"`
}

type WeatherData struct {
	City        string
	Temperature float64
	Windspeed   float64
	Weathercode int
}

type cachedWeather struct {
	data      WeatherData
	fetchedAt time.Time
}

const weatherCacheTTL = 10 * time.Minute

var (
	weatherCache   = make(map[string]cachedWeather)
	weatherCacheMu sync.RWMutex
)

type WeatherPageData struct {
	BaseData
	Weather *WeatherData
	City    string
}

func fetchCoordinates(city string) (float64, float64, error) {
	reqURL := fmt.Sprintf("%s?name=%s&count=1", geocodingBaseURL, url.QueryEscape(city))
	resp, err := http.Get(reqURL) //nolint:gosec
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing geocoding response body: %v", err)
		}
	}()

	var result geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, fmt.Errorf("geocoding decode failed: %w", err)
	}

	if len(result.Results) == 0 {
		return 0, 0, fmt.Errorf("city not found: %s", city)
	}

	r := result.Results[0]
	return r.Latitude, r.Longitude, nil
}

func fetchWeather(lat, lon float64) (WeatherData, error) {
	reqURL := fmt.Sprintf("%s?latitude=%f&longitude=%f&current_weather=true&hourly=temperature_2m,weathercode",
		weatherBaseURL, lat, lon)
	resp, err := http.Get(reqURL) //nolint:gosec
	if err != nil {
		return WeatherData{}, fmt.Errorf("weather request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing weather response body: %v", err)
		}
	}()

	var result openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return WeatherData{}, fmt.Errorf("weather decode failed: %w", err)
	}

	return WeatherData{
		Temperature: result.CurrentWeather.Temperature,
		Windspeed:   result.CurrentWeather.Windspeed,
		Weathercode: result.CurrentWeather.Weathercode,
	}, nil
}

func getWeatherForCity(city string) (*WeatherData, error) {
	weatherCacheMu.RLock()
	cached, ok := weatherCache[city]
	weatherCacheMu.RUnlock()

	if ok && time.Since(cached.fetchedAt) < weatherCacheTTL {
		wd := cached.data
		return &wd, nil
	}

	lat, lon, err := fetchCoordinates(city)
	if err != nil {
		log.Printf("fetchCoordinates error for %q: %v", city, err)
		return nil, fmt.Errorf("City not found")
	}

	wd, err := fetchWeather(lat, lon)
	if err != nil {
		log.Printf("fetchWeather error for %q: %v", city, err)
		return nil, fmt.Errorf("Could not fetch weather data")
	}

	wd.City = city
	weatherCacheMu.Lock()
	weatherCache[city] = cachedWeather{data: wd, fetchedAt: time.Now()}
	weatherCacheMu.Unlock()

	return &wd, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")

	data := WeatherPageData{
		BaseData: BaseData{User: getSessionUser(r)},
		City:     city,
	}

	if city != "" {
		wd, err := getWeatherForCity(city)
		if err != nil {
			data.Error = err.Error()
		} else {
			data.Weather = wd
		}
	}

	tmpl, err := parseTemplates("layout.html", "weather.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
