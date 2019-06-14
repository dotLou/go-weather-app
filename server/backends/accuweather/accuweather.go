package accuweather

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-weather-app/server/types"
	"net/http"

	"github.com/labstack/echo"
)

// Accuweather defines the configuration for an Accuweather backend
type Accuweather struct {
	APIKey string `json:"apiKey"`
	Logger echo.Logger
}

type locationCurrentWeatherResp []struct {
	WeatherText               string `json:"WeatherText"`
	TemperatureCurrentWeather `json:"Temperature"`
}
type TemperatureCurrentWeather struct {
	Metric `json:"Metric"`
}
type Metric struct {
	Value float32 `json:"Value"`
}

type location1DayForecastResp struct {
	DailyForecasts `json:"DailyForecasts"`
}
type DailyForecasts []struct {
	TemperatureDailyForecast `json:"Temperature"`
}
type TemperatureDailyForecast struct {
	Minimum `json:"Minimum"`
	Maximum `json:"Maximum"`
}
type Minimum struct {
	Value float32 `json:"Value"`
}
type Maximum struct {
	Value float32 `json:"Value"`
}
type locationKeyResp []struct {
	Key string `json:"Key"`
}

var citySearchURIF = "https://dataservice.accuweather.com/locations/v1/cities/search?q=%s&apikey=%s"
var locationCurrentWeatherURIF = "https://dataservice.accuweather.com/currentconditions/v1/%s?apikey=%s"
var location1DayForecastURIF = "https://dataservice.accuweather.com/forecasts/v1/daily/1day/%s?apikey=%s"

// GetWeather gets the whether for the specified city with via Accuweather
func (o Accuweather) GetWeather(city string) types.Weather {
	locationKey, err := o.getLocationKey(city)
	if err != nil {
		return types.Weather{
			Error: err.Error(),
		}
	}
	cwr, err := o.getCurrentWeather(locationKey)
	if err != nil || len(cwr) == 0 {
		return types.Weather{
			Error: err.Error(),
		}
	}

	odf, err := o.get1DayForecast(locationKey)
	if err != nil || len(odf.DailyForecasts) == 0 {
		return types.Weather{
			Error: err.Error(),
		}
	}

	return types.Weather{
		Source:          types.ACCUWEATHER,
		Temperature:     cwr[0].TemperatureCurrentWeather.Metric.Value,
		TemperatureMax:  odf.DailyForecasts[0].TemperatureDailyForecast.Maximum.Value,
		TemperatureMin:  odf.DailyForecasts[0].TemperatureDailyForecast.Minimum.Value,
		MainDescription: cwr[0].WeatherText,
	}
}

func (o Accuweather) getLocationKey(city string) (string, error) {
	citySearchURI := fmt.Sprintf(citySearchURIF, city, o.APIKey)
	resp, err := http.Get(citySearchURI)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		o.Logger.Error("accuweather encountered status code error:", resp.StatusCode)
		return "", errors.New("Error communicating to backend")
	}
	locResp := locationKeyResp{}
	err = json.NewDecoder(resp.Body).Decode(&locResp)
	if err != nil || locResp[0].Key == "" {
		return "", errors.New("Unable to determine location for provided city")
	}
	return locResp[0].Key, nil
}

func (o Accuweather) get1DayForecast(locationKey string) (location1DayForecastResp, error) {
	odf := location1DayForecastResp{}
	weatherURI := fmt.Sprintf(location1DayForecastURIF, locationKey, o.APIKey)
	resp, err := http.Get(weatherURI)
	if err != nil {
		return odf, err
	}
	if resp.StatusCode != 200 {
		o.Logger.Error("accuweather encountered status code error for 1dayforecast:", resp.StatusCode)
		return odf, errors.New("Error communicating to backend")
	}

	err = json.NewDecoder(resp.Body).Decode(&odf)
	if err != nil {
		o.Logger.Error("accuweather encountered error decoding response for 1dayforecast:", err)
		return odf, errors.New("Unable to decode response from backend")
	}
	return odf, nil
}

func (o Accuweather) getCurrentWeather(locationKey string) (locationCurrentWeatherResp, error) {
	cwr := locationCurrentWeatherResp{}
	weatherURI := fmt.Sprintf(locationCurrentWeatherURIF, locationKey, o.APIKey)
	resp, err := http.Get(weatherURI)
	if err != nil {
		return cwr, err
	}
	if resp.StatusCode != 200 {
		o.Logger.Error("accuweather encountered status code error for current weather:", resp.StatusCode)
		return cwr, errors.New("Error communicating to backend")
	}

	err = json.NewDecoder(resp.Body).Decode(&cwr)
	if err != nil {
		o.Logger.Error("accuweather encountered error decoding response for current weather:", err)
		return cwr, errors.New("Unable to decode response from backend")
	}
	return cwr, nil
}
