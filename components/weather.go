package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var OPEN_WEATHER_API = Config.OpenWeatherAPI

type City struct {
	Name string
	Lat  float64
	Lng  float64
}

type Current struct {
	Sunrise     int           `json:"sunrise"`
	Sunset      int           `json:"sunset"`
	Temp        float64       `json:"temp"`
	FeelsLike   float64       `json:"feels_like"`
	WindSpeed   float64       `json:"wind_speed"`
	WeatherList []WeatherFlag `json:"weather"`
}

type WeatherFlag struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Hourly struct {
	Dt        int           `json:"dt"`
	Temp      float64       `json:"temp"`
	FeelsLike float64       `json:"feels_like"`
	Weather   []WeatherFlag `json:"weather"`
	Pop       float64       `json:"pop"`
}

type Weather struct {
	Timezone       string   `json:"timezone"`
	TimezoneOffset int      `json:"timezone_offset"`
	Current        Current  `json:"current"`
	Hourly         []Hourly `json:"hourly"`
}

const KELVIN = -272.15

func GetWeather(city *City) (*Weather, error) {
	client := FastHttpClient

	//req := new(fasthttp.Request)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)
	//url := "https://api.openweathermap.org/data/2.5/onecall?lat=30.2937&lon=120.1614&exclude=minutely,daily&appid=5d6ce190f514d7ee813a38d98e32665b"
	u := url.URL{
		Scheme: "https",
		Host:   "api.openweathermap.org",
		Path:   "/data/2.5/onecall",
	}
	q := u.Query()
	q.Set("lat", fmt.Sprintf("%f", city.Lat))
	q.Set("lon", fmt.Sprintf("%f", city.Lng))
	q.Set("exclude", "minutely,daily")
	q.Set("appid", OPEN_WEATHER_API)
	u.RawQuery = q.Encode()
	req.SetRequestURI(u.String())

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := client.Do(req, res)
	if err != nil {
		log.Println(err)
		return new(Weather), err
	}

	result := new(Weather)
	err = json.Unmarshal(res.Body(), result)
	if err != nil {
		log.Println(err)
		return new(Weather), err
	}

	//fmt.Println(result.Current.WeatherList)
	return result, nil
}

// var cityInfoCache = make(map[string]*City)

func GetWeatherByName(cityName string) (*Weather, error) {
	// geocoding
	city := new(City)
	// get geocoding from sqlite3
	geocode, err := QueryGeoCodeByName(cityName)
	if err != nil {
		geocodingList, err := GetGeocodingByName(cityName)
		if err != nil {
			return nil, err
		}
		geocoding := geocodingList[0]
		g := &GeoCode{
			Name:    cityName,
			JaName:  geocoding.LocalNames.Ja,
			Lat:     geocoding.Lat,
			Lng:     geocoding.Lon,
			Country: geocoding.Country,
		}
		tx := SqliteConn.Begin()
		err = g.CreateGeoCode(tx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
		city.Lat = g.Lat
		city.Lng = g.Lng
		city.Name = g.Name
	} else {
		// if geocode.Updateat is outdate, update it
		city.Lat = geocode.Lat
		city.Lng = geocode.Lng
		city.Name = geocode.Name
	}

	// if v, ok := cityInfoCache[cityName]; !ok {
	// 	geocoding, err := GetGeocodingByName(cityName)
	// 	if err != nil {
	// 		return &Weather{}, err
	// 	}
	// 	city.Lat = geocoding[0].Lat
	// 	city.Lng = geocoding[0].Lon
	// 	city.Name = geocoding[0].Name
	// 	cityInfoCache[cityName] = city
	// } else {
	// 	city = v
	// }

	// getweatherbyLatLng
	weather, err := GetWeather(city)
	if err != nil {
		log.Println(err)
		return weather, err
	}
	return weather, err
}

func ShowCurrentWeather(cityName string) string {
	weather, err := GetWeatherByName(cityName)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s Weather: %s\nTemprature: %.2f°C\nFeels like: %.2f°C\nWind speed: %.2fm/s", cityName, weather.Current.WeatherList[0].Main, weather.Current.Temp+KELVIN, weather.Current.FeelsLike+KELVIN, weather.Current.WindSpeed)
}

type Geocoding struct {
	Name       string    `json:"name"`
	LocalNames NameLocal `json:"local_names"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	Country    string    `json:"country"`
}

type NameLocal struct {
	Ja string `json:"ja"`
	En string `json:"en"`
}

func GetGeocodingByName(cityName string) ([]*Geocoding, error) {
	client := FastHttpClient

	//req := new(fasthttp.Request)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(fasthttp.MethodGet)
	// https://api.openweathermap.org/geo/1.0/direct?q=hangzhou&limit=1&appid=5d6ce190f514d7ee813a38d98e32665b
	u := url.URL{
		Scheme: "https",
		Host:   "api.openweathermap.org",
		Path:   "/geo/1.0/direct",
	}
	q := u.Query()
	q.Set("q", cityName)
	q.Set("limit", fmt.Sprintf("%d", 1))
	q.Set("appid", OPEN_WEATHER_API)
	u.RawQuery = q.Encode()
	req.SetRequestURI(u.String())
	log.Println(u.String())

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := client.Do(req, res)
	if err != nil {
		log.Println(err)
		return []*Geocoding{}, err
	}

	//var result []*Geocoding
	result := make([]*Geocoding, 1)
	//var result = make([]*Geocoding, 0)
	err = json.Unmarshal(res.Body(), &result)
	if err != nil || len(result) == 0 {
		return result, fmt.Errorf("%s or len(result)==0", err)
	}
	return result, nil
}

func WeatherDaemon(b *tb.Bot, cityName string) {
	var myGroup = &tb.User{ID: -1001524256686}
	// get city info by cityName
	log.Println("weather bot daemon ...")
	// every new start, get a weather data and send to myGroup
	geocodings, err := GetGeocodingByName(cityName)
	if err != nil {
		b.Send(myGroup, err)
		return
	}
	geocoding := geocodings[0]
	queryCity := &City{
		Name: geocoding.Name,
		Lat:  geocoding.Lat,
		Lng:  geocoding.Lon,
	}
	data, err := GetWeather(queryCity)
	if err != nil {
		b.Send(myGroup, err)
		return
	}
	currentWeather := fmt.Sprintf("%s temprature:%.2f°C\nFeels like:%.2f°C\nWind speed: %.2fm/s\nWeather: %s\nTimezone: %s\nTime offset: %dh\n", cityName,
		data.Current.Temp+KELVIN, data.Current.FeelsLike+KELVIN, data.Current.WindSpeed, data.Current.WeatherList[0].Main, data.Timezone, data.TimezoneOffset/3600)
	b.Send(myGroup, currentWeather)

	const Possibility float64 = 0.8
	const HighPossibility float64 = 0.9
	for {
		now := time.Now().In(Loc)
		h := now.Hour()
		m := now.Minute()
		rainList := make([]Hourly, 0, 24)
		switch h {
		case 23, 0, 1, 2, 3, 4, 5, 6:
		default:
			// future 3 hours
			weather, err := GetWeather(queryCity)
			if err != nil {
				b.Send(myGroup, err.Error())
			} else if h == 18 && m == 0 || h == 8 && m == 0 {
				for i, v := range weather.Hourly {
					if i < 24 && v.Weather[0].Main == "Rain" && v.Pop > Possibility {
						rainList = append(rainList, v)
					}
				}
			} else {
				for i, v := range weather.Hourly {
					if i < 3 && v.Weather[0].Main == "Rain" && v.Pop > HighPossibility {
						rainList = append(rainList, v)
					}
				}
			}
		}
		if len(rainList) > 0 {
			b.Send(myGroup, processRainList(rainList))
			time.Sleep(time.Hour)
		} else {
			time.Sleep(time.Minute)
		}
	}
}

func processRainList(rainList []Hourly) string {
	result := ""
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range rainList {
		date := time.Unix(int64(v.Dt), 0)
		date = date.In(loc)
		dateFormat := date.Format("2006-01-02 15:04:05")
		dateFormat1 := date.Add(time.Hour).Format("2006-01-02 15:04:05")
		//dateFormatAfterOneHour := date.Add(time.Hour).Format("2006-01-02 15:04:05")
		result += fmt.Sprintf("At:%s - %s\nTemprature:%.2f°C\nFeels like: %.2f°C\nWeather: %s\nPossibility: %.2f\n========\n", dateFormat, dateFormat1, v.Temp+KELVIN, v.FeelsLike+KELVIN, v.Weather[0].Main, v.Pop)
	}
	return result
}
