package openweathermap

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-weather-app/server/types"
	"net/http"

	"github.com/labstack/echo"
)

// Openweathermap defines the configuration for an openweathermap backend
type Openweathermap struct {
	APIKey string `json:"apiKey"`
	Logger echo.Logger
}

type cityWeatherResp struct {
	WeatherDetails `json:"weather"`
	MainDetails    `json:"main"`
}

type WeatherDetails []struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}
type MainDetails struct {
	Temp    float32 `json:"temp"`
	TempMin float32 `json:"temp_min"`
	TempMax float32 `json:"temp_max"`
}

var cityWeatherURIF = "https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&APPID=%s"

// GetWeather gets the whether for the specified city with via openweathermap
func (o Openweathermap) GetWeather(city string) types.Weather {
	cwr, err := o.getWeather(city)
	if err != nil {
		return types.Weather{
			Error: err.Error(),
		}
	}

	return types.Weather{
		Source:              types.OPENWEATHERMAP,
		Temperature:         cwr.MainDetails.Temp,
		TemperatureMax:      cwr.MainDetails.TempMax,
		TemperatureMin:      cwr.MainDetails.TempMin,
		MainDescription:     cwr.WeatherDetails[0].Main,
		DetailedDescription: cwr.WeatherDetails[0].Description,
	}
}

func (o Openweathermap) getWeather(city string) (*cityWeatherResp, error) {
	weatherURI := fmt.Sprintf(cityWeatherURIF, city, o.APIKey)
	resp, err := http.Get(weatherURI)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		o.Logger.Error("openweathermap encountered status code error:", resp.StatusCode)
		return nil, errors.New("Error communicating to backend")
	}
	cwr := &cityWeatherResp{}
	err = json.NewDecoder(resp.Body).Decode(cwr)
	if err != nil {
		o.Logger.Error("openweathermap encountered error decoding response:", err)
		return nil, errors.New("Unable to decode response from backend")
	}
	return cwr, nil
}
