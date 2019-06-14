package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"go-weather-app/server/backends/accuweather"
	"go-weather-app/server/backends/openweathermap"
	"go-weather-app/server/metrics"
	"go-weather-app/server/types"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	//metrics
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WeatherResponse defines a json response for multiple weather responses (i.e. from multiple backends)
type WeatherResponse struct {
	City  string          `json:"city"`
	Data  []types.Weather `json:"data"`
	Error string          `json:"error"` // this is used as a response whenever a bad request comes in
}

// BackendResponse defines a json response for the configured/known backends
type BackendResponse struct {
	Backends []string `json:"backends"`
}

// DefaultBackends defines the default backends to pull weather data from, when none are specified explicitly
var DefaultBackends = []string{}

// ConfiguredBackends is the map of known backend configuration interfaces
var ConfiguredBackends map[string]types.WeatherBackend

// requestMetricsMiddleware is used to gather metrics on every incoming request
func requestMetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			metrics.HTTPRequestsTotal.With(prometheus.Labels{"path": c.Path(), "method": c.Request().Method}).Inc()
			return next(c)
		}
	}
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost", "http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	v1Api := e.Group("/v1")
	v1Api.Use(requestMetricsMiddleware()) // only track metrics on requests under /v1 (i.e. don't track /metrics)

	err := configureServer(e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	v1Api.GET("/weather/:city", getWeather)
	v1Api.OPTIONS("/weather", optionsWeather)
	v1Api.GET("/backends", getBackends)

	// Start server
	go func() {
		if err := e.Start(":8080"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	/* Wait for interrupt signal to gracefully shutdown the server with
	a timeout of 10 seconds. */
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

// Config defines the server configurations
type Config struct {
	Backends Backends `json:"backends"`
}

// Backends defines the structure used to configure various weather backends for the server
type Backends struct {
	openweathermap.Openweathermap `json:"openweathermap"`
	accuweather.Accuweather       `json:"accuweather"`
}

func loadConfigFile(configFilePath string) (*Config, error) {
	configFile, err := os.Open(configFilePath)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func configureBackends(config *Config, logger echo.Logger) error {
	ConfiguredBackends = map[string]types.WeatherBackend{} // init the map

	if config.Backends.Accuweather.APIKey != "" {
		config.Backends.Accuweather.Logger = logger
		ConfiguredBackends[types.ACCUWEATHER] = config.Backends.Accuweather
	}
	if config.Backends.Openweathermap.APIKey != "" {
		config.Backends.Openweathermap.Logger = logger
		ConfiguredBackends[types.OPENWEATHERMAP] = config.Backends.Openweathermap
	}

	if len(ConfiguredBackends) == 0 {
		return errors.New("No weather backends configured")
	}

	DefaultBackends = []string{}
	// set our default backends to be all known backends, for cases where none is specified
	for backend := range ConfiguredBackends {
		DefaultBackends = append(DefaultBackends, backend)
	}
	return nil
}

func configureServer(logger echo.Logger) error {
	config, err := loadConfigFile("config.json")
	if err != nil {
		return err
	}

	err = configureBackends(config, logger)
	if err != nil {
		return err
	}

	return nil
}

func validateBackends(backends []string) error {
	for _, backend := range backends {
		if ConfiguredBackends[backend] == nil {
			return errors.New("Backend specified is invalid or inactive: " + backend)
		}
	}
	return nil
}

func getWeather(c echo.Context) error {
	response := &WeatherResponse{}

	response.City = strings.TrimSpace(c.Param("city"))
	if len(response.City) == 0 {
		response.Error = "No city specified. Please provide a city query parameter."
		return c.JSONPretty(http.StatusBadRequest, response, "  ")
	}

	backendParam := strings.TrimSpace(c.QueryParam("backend"))
	var targetBackends []string

	if len(backendParam) == 0 {
		targetBackends = DefaultBackends
	} else {
		targetBackends = strings.Split(backendParam, ",")
		err := validateBackends(targetBackends)
		if err != nil {
			response.Error = err.Error()
			return c.JSONPretty(http.StatusBadRequest, response, "  ")
		}
	}

	for _, backend := range targetBackends {
		weather := ConfiguredBackends[backend].GetWeather(response.City)
		response.Data = append(response.Data, weather)
	}

	return c.JSONPretty(http.StatusOK, response, "  ")
}

var knownBackends BackendResponse

func getBackends(c echo.Context) error {
	// only compute known backends the first time to save time
	if len(knownBackends.Backends) == 0 {
		for backend := range ConfiguredBackends {
			knownBackends.Backends = append(knownBackends.Backends, backend)
		}
	}
	return c.JSONPretty(http.StatusOK, knownBackends, "  ")
}

func optionsWeather(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderAccept, "GET, OPTIONS")
	return c.String(http.StatusOK, "")
}
