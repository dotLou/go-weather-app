package types

import "time"

// Weather defines the structure of a weather response
type Weather struct {
	Source              string  `json:"source"`
	Temperature         float32 `json:"temperature"`
	TemperatureMin      float32 `json:"temperature_min"`
	TemperatureMax      float32 `json:"temperature_max"`
	MainDescription     string  `json:"main_description,omitempty"`
	DetailedDescription string  `json:"detailed_description,omitempty"`
	Error               string  `json:"error,omitempty"` // this is used to give an error if the target backend returned an error
}

// WeatherBackend describes the interface for getting weather
type WeatherBackend interface {
	GetWeather(city string) Weather
}

// ACCUWEATHER defines the key for refering to the accuweather backend
const ACCUWEATHER = "accuweather"

// OPENWEATHERMAP defines the key for refering to the openweathermap backend
const OPENWEATHERMAP = "openweathermap"

// WeatherSchema is an example schema for what it might look like to store this data in a relational db
type WeatherSchema struct {
	ID                  int64
	TimeFetched         time.Time
	city                string
	Source              string
	Temperature         float32
	TemperatureMin      float32
	TemperatureMax      float32
	MainDescription     string
	DetailedDescription string
	Error               string
}
