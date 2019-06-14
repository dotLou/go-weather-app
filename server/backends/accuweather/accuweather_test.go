package accuweather

import (
	"errors"
	"go-weather-app/server/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/require"
)

func TestAccuweather_getLocationKey(t *testing.T) {
	tests := []struct {
		name          string
		logger        echo.Logger
		serverHandler func(http.ResponseWriter, *http.Request)
		city          string
		want          string
		expectedErr   error
	}{
		{
			name:   "error when backend returns non 200 status code",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want:        "",
			expectedErr: errors.New("Error communicating to backend"),
		},
		{
			name:   "error http.get throws error parsing url",
			city:   string(byte(0x7f)), // invalid city should cause error
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
			},
			want:        "",
			expectedErr: errors.New("net/url: invalid control character in URL"),
		},
		{
			name:   "error when backend returns bad response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want:        "",
			expectedErr: errors.New("Unable to determine location for provided city"),
		},
		{
			name:   "error when backend returns no key at 0",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"foo\":\"bar\"}]"))
			},
			want:        "",
			expectedErr: errors.New("Unable to determine location for provided city"),
		},
		{
			name:   "proper key response when backend returns proper response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Key\":\"1234\"}]"))
			},
			want:        "1234",
			expectedErr: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup fake backend
			ts := httptest.NewServer(http.HandlerFunc(tc.serverHandler))
			defer ts.Close()
			// override URL so we can use our test server above instead
			origCitySearchURIF := citySearchURIF
			citySearchURIF = ts.URL + "?q=%s&apiKey=%s"
			defer func() { citySearchURIF = origCitySearchURIF }()

			o := Accuweather{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got, err := o.getLocationKey(tc.city)

			if tc.expectedErr != nil {
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func TestAccuweather_get1DayForecast(t *testing.T) {
	tests := []struct {
		name          string
		logger        echo.Logger
		locationKey   string
		serverHandler func(http.ResponseWriter, *http.Request)
		want          location1DayForecastResp
		expectedErr   error
	}{
		{
			name:   "error when backend returns non 200 status code",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want:        location1DayForecastResp{},
			expectedErr: errors.New("Error communicating to backend"),
		},
		{
			name:        "error http.get throws error parsing url",
			locationKey: string(byte(0x7f)), // invalid locationKey should cause error
			logger:      echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
			},
			want:        location1DayForecastResp{},
			expectedErr: errors.New("net/url: invalid control character in URL"),
		},
		{
			name:   "error when backend returns bad response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want:        location1DayForecastResp{},
			expectedErr: errors.New("Unable to decode response from backend"),
		},
		{
			name:   "proper weather response when backend returns proper response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("{\"DailyForecasts\":[{\"Temperature\":{\"Minimum\":{\"Value\":15},\"Maximum\":{\"Value\":22}}}]}"))
			},
			want: location1DayForecastResp{
				DailyForecasts: DailyForecasts{
					{TemperatureDailyForecast{
						Maximum: Maximum{
							Value: 22,
						},
						Minimum: Minimum{
							Value: 15,
						},
					}},
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
			origLocation1DayForecastURIF := location1DayForecastURIF
			location1DayForecastURIF = ts.URL + "?q=%s&APPID=%s"
			defer func() { location1DayForecastURIF = origLocation1DayForecastURIF }()

			o := Accuweather{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got, err := o.get1DayForecast(tc.locationKey)

			if tc.expectedErr != nil {
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func TestAccuweather_getCurrentWeather(t *testing.T) {

	tests := []struct {
		name          string
		logger        echo.Logger
		locationKey   string
		serverHandler func(http.ResponseWriter, *http.Request)
		want          locationCurrentWeatherResp
		expectedErr   error
	}{
		{
			name:   "error when backend returns non 200 status code",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want:        locationCurrentWeatherResp{},
			expectedErr: errors.New("Error communicating to backend"),
		},
		{
			name:        "error http.get throws error parsing url",
			locationKey: string(byte(0x7f)), // invalid locationKey should cause error
			logger:      echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
			},
			want:        locationCurrentWeatherResp{},
			expectedErr: errors.New("net/url: invalid control character in URL"),
		},
		{
			name:   "error when backend returns bad response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			want:        locationCurrentWeatherResp{},
			expectedErr: errors.New("Unable to decode response from backend"),
		},
		{
			name:   "proper weather response when backend returns proper response",
			logger: echo.New().Logger,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Temperature\":{\"Metric\":{\"Value\":20}},\"WeatherText\":\"Sunny\"}]"))
			},
			want: locationCurrentWeatherResp{
				{
					WeatherText: "Sunny",
					TemperatureCurrentWeather: TemperatureCurrentWeather{
						Metric: Metric{
							Value: 20,
						},
					},
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
			origLocationCurrentWeatherURIF := locationCurrentWeatherURIF
			locationCurrentWeatherURIF = ts.URL + "?q=%s&apiKey=%s"
			defer func() { locationCurrentWeatherURIF = origLocationCurrentWeatherURIF }()

			o := Accuweather{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got, err := o.getCurrentWeather(tc.locationKey)

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

func TestAccuweather_GetWeather(t *testing.T) {
	tests := []struct {
		name             string
		logger           echo.Logger
		lkServerHandler  func(http.ResponseWriter, *http.Request)
		cwServerHandler  func(http.ResponseWriter, *http.Request)
		odfServerHandler func(http.ResponseWriter, *http.Request)
		city             string
		want             types.Weather
	}{
		{
			name:   "error when locationKey backend returns non 200 status code",
			logger: echo.New().Logger,
			lkServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: types.Weather{
				Error: "Error communicating to backend",
			},
		},
		{
			name:   "error when current weather backend returns non 200 status code",
			logger: echo.New().Logger,
			lkServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Key\":\"1234\"}]"))
			},
			cwServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: types.Weather{
				Error: "Error communicating to backend",
			},
		},
		{
			name:   "error when 1day forecast backend returns non 200 status code",
			logger: echo.New().Logger,
			lkServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Key\":\"1234\"}]"))
			},
			cwServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Temperature\":{\"Metric\":{\"Value\":20}},\"WeatherText\":\"Sunny\"}]"))
			},
			odfServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			want: types.Weather{
				Error: "Error communicating to backend",
			},
		},
		{
			name:   "proper weather response when all backends return proper response",
			logger: echo.New().Logger,
			lkServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Key\":\"1234\"}]"))
			},
			cwServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("[{\"Temperature\":{\"Metric\":{\"Value\":20}},\"WeatherText\":\"Sunny\"}]"))
			},
			odfServerHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
				w.Write([]byte("{\"DailyForecasts\":[{\"Temperature\":{\"Minimum\":{\"Value\":15},\"Maximum\":{\"Value\":22}}}]}"))
			},
			want: types.Weather{
				Source:          types.ACCUWEATHER,
				Temperature:     20,
				TemperatureMax:  22,
				TemperatureMin:  15,
				MainDescription: "Sunny",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup fake backends
			lkts := httptest.NewServer(http.HandlerFunc(tc.lkServerHandler))
			defer lkts.Close()
			// override URL so we can use our test server above instead
			origCitySearchURIF := citySearchURIF
			citySearchURIF = lkts.URL + "?q=%s&apiKey=%s"
			defer func() { citySearchURIF = origCitySearchURIF }()

			odfts := httptest.NewServer(http.HandlerFunc(tc.odfServerHandler))
			defer odfts.Close()
			// override URL so we can use our test server above instead
			origLocation1DayForecastURIF := location1DayForecastURIF
			location1DayForecastURIF = odfts.URL + "?q=%s&APPID=%s"
			defer func() { location1DayForecastURIF = origLocation1DayForecastURIF }()

			cwts := httptest.NewServer(http.HandlerFunc(tc.cwServerHandler))
			defer cwts.Close()
			// override URL so we can use our test server above instead
			origLocationCurrentWeatherURIF := locationCurrentWeatherURIF
			locationCurrentWeatherURIF = cwts.URL + "?q=%s&apiKey=%s"
			defer func() { locationCurrentWeatherURIF = origLocationCurrentWeatherURIF }()

			o := Accuweather{
				APIKey: "fookey",
				Logger: tc.logger,
			}
			got := o.GetWeather(tc.city)
			require.Equal(t, tc.want, got)
		})
	}
}
