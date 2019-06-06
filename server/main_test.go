package main

import (
	"errors"
	"go-weather-app/server/backends/accuweather"
	"go-weather-app/server/backends/openweathermap"
	"go-weather-app/server/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func Test_configureBackends(t *testing.T) {
	tests := []struct {
		name                       string
		config                     *Config
		logger                     echo.Logger
		expectedConfiguredBackends map[string]types.WeatherBackend
		expectedDefaultBackends    []string
		expectedErr                error
	}{
		{
			name:                       "no configured backends returns error",
			config:                     &Config{},
			expectedConfiguredBackends: map[string]types.WeatherBackend{},
			expectedDefaultBackends:    []string{},
			expectedErr:                errors.New("No weather backends configured"),
		},
		{
			name: "configure accuweather",
			config: &Config{
				Backends: Backends{
					Accuweather: accuweather.Accuweather{
						APIKey: "foo",
					},
				},
			},
			expectedConfiguredBackends: map[string]types.WeatherBackend{
				types.ACCUWEATHER: accuweather.Accuweather{
					APIKey: "foo",
				},
			},
			expectedDefaultBackends: []string{types.ACCUWEATHER},
			expectedErr:             nil,
		},
		{
			name: "configure openweathermap",
			config: &Config{
				Backends: Backends{
					Openweathermap: openweathermap.Openweathermap{
						APIKey: "foo",
					},
				},
			},
			expectedConfiguredBackends: map[string]types.WeatherBackend{
				types.OPENWEATHERMAP: openweathermap.Openweathermap{
					APIKey: "foo",
				},
			},
			expectedDefaultBackends: []string{types.OPENWEATHERMAP},
			expectedErr:             nil,
		},
		{
			name: "configure openweathermap and accuweather",
			config: &Config{
				Backends: Backends{
					Accuweather: accuweather.Accuweather{
						APIKey: "bar",
					},
					Openweathermap: openweathermap.Openweathermap{
						APIKey: "foo",
					},
				},
			},
			expectedConfiguredBackends: map[string]types.WeatherBackend{
				types.ACCUWEATHER: accuweather.Accuweather{
					APIKey: "bar",
				},
				types.OPENWEATHERMAP: openweathermap.Openweathermap{
					APIKey: "foo",
				},
			},
			expectedDefaultBackends: []string{types.ACCUWEATHER, types.OPENWEATHERMAP},
			expectedErr:             nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := configureBackends(tc.config, tc.logger)
			require.Equal(t, tc.expectedErr, err)

			require.Equal(t, tc.expectedConfiguredBackends, ConfiguredBackends)
			require.Equal(t, tc.expectedDefaultBackends, DefaultBackends)
		})
	}
}

type mockWeatherBackend struct {
	returnWeather types.Weather
}

func (m mockWeatherBackend) GetWeather(city string) types.Weather {
	return m.returnWeather
}

func Test_validateBackends(t *testing.T) {
	tests := []struct {
		name               string
		backends           []string
		ConfiguredBackends map[string]types.WeatherBackend
		expectedErr        error
	}{
		{
			name:               "no default backends, no provided backends",
			backends:           []string{},
			ConfiguredBackends: map[string]types.WeatherBackend{},
			expectedErr:        nil,
		},
		{
			name:               "no default backends, backend not found",
			backends:           []string{"foo"},
			ConfiguredBackends: map[string]types.WeatherBackend{},
			expectedErr:        errors.New("Backend specified is invalid or inactive: foo"),
		},
		{
			name:     "default backend set, provided backend found",
			backends: []string{"foo"},
			ConfiguredBackends: map[string]types.WeatherBackend{
				"foo": mockWeatherBackend{},
			},
			expectedErr: nil,
		},
		{
			name:     "multiple default backend set, all provided backends found",
			backends: []string{"foo", "bar"},
			ConfiguredBackends: map[string]types.WeatherBackend{
				"foo": mockWeatherBackend{},
				"bar": mockWeatherBackend{},
			},
			expectedErr: nil,
		},
		{
			name:     "multiple default backend set, NOT all provided backends found",
			backends: []string{"foo", "bar", "baz"},
			ConfiguredBackends: map[string]types.WeatherBackend{
				"foo": mockWeatherBackend{},
				"baz": mockWeatherBackend{},
			},
			expectedErr: errors.New("Backend specified is invalid or inactive: bar"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//override ConfiguredBackends for test
			origWeatherBackends := ConfiguredBackends
			ConfiguredBackends = tc.ConfiguredBackends
			defer func() { ConfiguredBackends = origWeatherBackends }()

			err := validateBackends(tc.backends)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func Test_getWeather(t *testing.T) {
	tests := []struct {
		name               string
		city               string
		backendParam       string
		ConfiguredBackends map[string]types.WeatherBackend
		DefaultBackends    []string
		expectedErr        error
		expectedHTTPStatus int
		expectedBody       string
	}{
		{
			name:               "no city specified",
			city:               "",
			backendParam:       "",
			ConfiguredBackends: map[string]types.WeatherBackend{},
			DefaultBackends:    []string{},
			expectedErr:        nil,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedBody:       "{\n  \"error\": \"No city specified. Please provide a city query parameter.\"\n}\n",
		},
		{
			name:         "default backends used",
			city:         "foo",
			backendParam: "",
			ConfiguredBackends: map[string]types.WeatherBackend{
				"fooBackend": mockWeatherBackend{
					returnWeather: types.Weather{
						Source:              "fooBackend",
						Temperature:         12,
						TemperatureMin:      2,
						TemperatureMax:      20,
						MainDescription:     "Sunny",
						DetailedDescription: "Mix of sun and clouds",
					},
				},
			},
			DefaultBackends:    []string{"fooBackend"},
			expectedErr:        nil,
			expectedHTTPStatus: http.StatusOK,
			expectedBody:       "{\n  \"city\": \"foo\",\n  \"data\": [\n    {\n      \"source\": \"fooBackend\",\n      \"temperature\": 12,\n      \"temperature_min\": 2,\n      \"temperature_max\": 20,\n      \"main_description\": \"Sunny\",\n      \"detailed_description\": \"Mix of sun and clouds\"\n    }\n  ]\n}\n",
		},
		{
			name:               "specified backend does not exist",
			city:               "foo",
			backendParam:       "?backend=fooBackend",
			ConfiguredBackends: map[string]types.WeatherBackend{},
			DefaultBackends:    []string{},
			expectedErr:        nil,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedBody:       "{\n  \"city\": \"foo\",\n  \"error\": \"Backend specified is invalid or inactive: fooBackend\"\n}\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//override ConfiguredBackends for test
			origWeatherBackends := ConfiguredBackends
			ConfiguredBackends = tc.ConfiguredBackends
			defer func() { ConfiguredBackends = origWeatherBackends }()

			//override DefaultBackends for test
			origDefaultBackends := DefaultBackends
			DefaultBackends = tc.DefaultBackends
			defer func() { DefaultBackends = origDefaultBackends }()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/weather/"+tc.city+tc.backendParam, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/weather/:city")
			if len(tc.city) > 0 {
				c.SetParamNames("city")
				c.SetParamValues(tc.city)
			}

			err := getWeather(c)
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedHTTPStatus, rec.Code)
			require.Equal(t, tc.expectedBody, rec.Body.String())
		})
	}
}

func Test_optionsWeather(t *testing.T) {
	t.Run("OPTIONS", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodOptions, "/weather/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/weather")

		err := optionsWeather(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "GET, OPTIONS", rec.Header()["Accept"][0])
	})

}

func Test_getBackends(t *testing.T) {
	tests := []struct {
		name               string
		configuredBackends map[string]types.WeatherBackend
		knownBackends      BackendResponse
		expectedBody       string
		expectedErr        error
	}{
		{
			name: "uses Configured backends when knownBackends is not set",
			configuredBackends: map[string]types.WeatherBackend{
				"foo": mockWeatherBackend{},
				"bar": mockWeatherBackend{},
			},
			knownBackends: BackendResponse{},
			expectedBody:  "{\n  \"backends\": [\n    \"foo\",\n    \"bar\"\n  ]\n}\n",
			expectedErr:   nil,
		},
		{
			name:               "uses knownBackends when knownBackends is already set",
			configuredBackends: map[string]types.WeatherBackend{},
			knownBackends: BackendResponse{
				Backends: []string{"bar", "foo"},
			},
			expectedBody: "{\n  \"backends\": [\n    \"bar\",\n    \"foo\"\n  ]\n}\n",
			expectedErr:  nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//override ConfiguredBackends for test
			origWeatherBackends := ConfiguredBackends
			ConfiguredBackends = tc.configuredBackends
			defer func() { ConfiguredBackends = origWeatherBackends }()

			//override knownBackends for test
			origKnownBackends := knownBackends
			knownBackends = tc.knownBackends
			defer func() { knownBackends = origKnownBackends }()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/backends", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/backends")

			err := getBackends(c)
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, http.StatusOK, rec.Code)
			require.Equal(t, tc.expectedBody, rec.Body.String())
		})
	}
}
