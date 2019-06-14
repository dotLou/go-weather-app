package openweathermap

import (
	"errors"
	"go-weather-app/server/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func TestOpenweathermap_GetWeather(t *testing.T) {

	tests := []struct {
		name          string
		logger        echo.Logger
		serverHandler func(http.ResponseWriter, *http.Request)
		want          types.Weather
	}{
		{
			name:   "error when backend returns non 200 status code",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: types.Weather{
				Error: "Error communicating to backend",
			},
		},
		{
			name:   "proper weather response when backend returns proper response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("{\"weather\":[{\"main\":\"Sunny\",\"description\":\"Mainly sunny\"}],\"main\":{\"temp\":20,\"temp_min\":15,\"temp_max\":22}}"))
			},
			want: types.Weather{
				Source:              types.OPENWEATHERMAP,
				Temperature:         20,
				TemperatureMax:      22,
				TemperatureMin:      15,
				MainDescription:     "Sunny",
				DetailedDescription: "Mainly sunny",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup fake backend
			ts := httptest.NewServer(http.HandlerFunc(tc.serverHandler))
			defer ts.Close()
			// override URL so we can use our test server above instead
			origCityWeatherURIF := cityWeatherURIF
			cityWeatherURIF = ts.URL + "?q=%s&APPID=%s"
			defer func() { cityWeatherURIF = origCityWeatherURIF }()

			o := Openweathermap{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got := o.GetWeather("foo")

			require.Equal(t, tc.want, got)

		})
	}
}

func TestOpenweathermap_getWeather(t *testing.T) {

	tests := []struct {
		name          string
		logger        echo.Logger
		city          string
		serverHandler func(http.ResponseWriter, *http.Request)
		want          *cityWeatherResp
		expectedErr   error
	}{
		{
			name:   "error when backend returns non 200 status code",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want:        nil,
			expectedErr: errors.New("Error communicating to backend"),
		},
		{
			name:   "error http.get throws error parsing url",
			city:   string(byte(0x7f)), // invalid city should cause error
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
			},
			want:        nil,
			expectedErr: errors.New("net/url: invalid control character in URL"),
		},
		{
			name:   "error when backend returns bad response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want:        nil,
			expectedErr: errors.New("Unable to decode response from backend"),
		},
		{
			name:   "proper weather response when backend returns proper response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("{\"weather\":[{\"main\":\"Sunny\",\"description\":\"Mainly sunny\"}],\"main\":{\"temp\":20,\"temp_min\":15,\"temp_max\":22}}"))
			},
			want: &cityWeatherResp{
				WeatherDetails: WeatherDetails{
					{
						Main:        "Sunny",
						Description: "Mainly sunny",
					},
				},
				MainDetails: MainDetails{
					Temp:    20,
					TempMin: 15,
					TempMax: 22,
				},
			},
			expectedErr: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup fake backend
			ts := httptest.NewServer(http.HandlerFunc(tc.serverHandler))
			defer ts.Close()
			// override URL so we can use our test server above instead
			origCityWeatherURIF := cityWeatherURIF
			cityWeatherURIF = ts.URL + "?q=%s&APPID=%s"
			defer func() { cityWeatherURIF = origCityWeatherURIF }()

			o := Openweathermap{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got, err := o.getWeather(tc.city)

			if tc.expectedErr != nil {
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			if tc.want == nil {
				require.Nil(t, got)
			} else {
				require.Equal(t, tc.want, got)
			}

		})
	}
}
