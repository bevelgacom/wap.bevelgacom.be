package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"text/template"
	"time"

	"github.com/hectormalot/omgo"
	"github.com/labstack/echo/v4"
)

var weatherLocation = map[string]string{}
var weatherLocationLock = sync.RWMutex{}

type WeatherLocation struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Elevation   float64  `json:"elevation"`
	FeatureCode string   `json:"feature_code"`
	CountryCode string   `json:"country_code"`
	Admin1ID    int      `json:"admin1_id"`
	Admin3ID    int      `json:"admin3_id,omitempty"`
	Admin4ID    int      `json:"admin4_id,omitempty"`
	Timezone    string   `json:"timezone"`
	Population  int      `json:"population"`
	Postcodes   []string `json:"postcodes"`
	CountryID   int      `json:"country_id"`
	Country     string   `json:"country"`
	Admin1      string   `json:"admin1"`
	Admin3      string   `json:"admin3,omitempty"`
	Admin4      string   `json:"admin4,omitempty"`
	Admin2ID    int      `json:"admin2_id,omitempty"`
	Admin2      string   `json:"admin2,omitempty"`
}

type WeatherLocationResult struct {
	Results          []WeatherLocation `json:"results"`
	GenerationtimeMs float64           `json:"generationtime_ms"`
}

func lookUpLocation(name string) ([]WeatherLocation, error) {
	// do HTTP request to get location
	resp, err := http.Get(fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=10&language=en&format=json", url.QueryEscape(name)))
	if err != nil {
		return nil, err
	}

	// decode the response
	var result WeatherLocationResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}

type WeatherPageLocation struct {
	ID   string `json:"id"` // lat + long
	Name string `json:"name"`
}

type WeatherLocationPage struct {
	LocationList  []WeatherPageLocation
	LocationValue string
}

func serveWeatherLocation(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("./static/weather/location.wml"))

	page := WeatherLocationPage{
		LocationValue: c.QueryParam("loc"),
	}

	if page.LocationValue != "" {
		locations, err := lookUpLocation(page.LocationValue)
		if err == nil {
			locs := []WeatherPageLocation{}
			for _, loc := range locations {
				id := fmt.Sprintf("%f,%f", loc.Latitude, loc.Longitude)
				locs = append(locs, WeatherPageLocation{
					ID:   id,
					Name: fmt.Sprintf("%s, %s", loc.Name, loc.Country),
				})

				weatherLocationLock.Lock()
				weatherLocation[id] = loc.Name
				weatherLocationLock.Unlock()
			}
			page.LocationList = locs
		}
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
	return tmpl.Execute(c.Response().Writer, page)
}

type WeatherCondition struct {
	Icon          string
	Temperature   string
	WindSpeed     string
	WindDirection string

	// for forcecast
	Time          string
	Precipitation string
}

type WeatherDetailPage struct {
	LocationID string
	Location   string
	Now        WeatherCondition
}

func wwoToOpenweathermapIcon(wwo string) string {
	switch wwo {
	case "0":
		return "01d"
	case "1", "2", "3":
		return "02d"
	case "45", "48":
		return "50d"
	case "51", "53", "55":
		return "09d"
	case "56", "57":
		return "13d"
	case "61", "63", "65":
		return "10d"
	case "66", "67":
		return "13d"
	case "71", "73", "75":
		return "13d"
	case "77":
		return "13d"
	case "80", "81", "82":
		return "09d"
	case "85", "86":
		return "13d"
	case "95":
		return "11d"
	case "96", "99":
		return "11d"
	}
	return "01d"
}

func windDirection(dir float64) string {
	switch {
	case dir >= 337.5 || dir < 22.5:
		return "N"
	case dir < 67.5:
		return "NE"
	case dir < 112.5:
		return "E"
	case dir < 157.5:
		return "SE"
	case dir < 202.5:
		return "S"
	case dir < 247.5:
		return "SW"
	case dir < 292.5:
		return "W"
	default:
		return "NW"
	}
}

func serveWeatherDetailView(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("./static/weather/details.wml"))
	locStr := c.QueryParam("loc")
	if locStr == "" {
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	// parse the location
	var lat, long float64
	_, err := fmt.Sscanf(locStr, "%f,%f", &lat, &long)
	if err != nil {
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	wc, _ := omgo.NewClient()
	loc, err := omgo.NewLocation(lat, long)
	if err != nil {
		log.Println(err)
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	opts := omgo.Options{
		Timezone:      "Europe/Brussels",
		PastDays:      0,
		HourlyMetrics: []string{"cloudcover, relativehumidity_2m"},
		DailyMetrics:  []string{"temperature_2m_max"},
	}

	page := WeatherDetailPage{
		LocationID: locStr,
	}
	weatherLocationLock.RLock()
	locName, ok := weatherLocation[locStr]
	weatherLocationLock.RUnlock()
	if ok {
		page.Location = locName
	}

	cw, err := wc.CurrentWeather(context.Background(), loc, &opts)
	if err != nil {
		log.Println(err)
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	page.Now = WeatherCondition{
		Icon:          wwoToOpenweathermapIcon(fmt.Sprintf("%d", int(cw.WeatherCode))),
		Temperature:   fmt.Sprintf("%.1f°C", cw.Temperature),
		WindSpeed:     fmt.Sprintf("%.1f km/h", cw.WindSpeed),
		WindDirection: windDirection(cw.WindDirection),
	}

	/*wf, err := wc.Forecast(context.Background(), loc, &opts)
	if err != nil {
		log.Println(err)
		return c.Redirect(http.StatusFound, "/weather/location")
	}*/

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")

	return tmpl.Execute(c.Response().Writer, page)
}

type WeatherHourlyPage struct {
	LocationID string
	Location   string
	Offset     int
	Data       []WeatherCondition
}

func serveWeatherHourly(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("./static/weather/hourly.wml"))
	locStr := c.QueryParam("loc")
	if locStr == "" {
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	offset := 0
	if c.QueryParam("o") != "" {
		_, err := fmt.Sscanf(c.QueryParam("o"), "%d", &offset)
		if err != nil {
			offset = 0
		}
	}

	// parse the location
	var lat, long float64
	_, err := fmt.Sscanf(locStr, "%f,%f", &lat, &long)
	if err != nil {
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	wc, _ := omgo.NewClient()
	loc, err := omgo.NewLocation(lat, long)
	if err != nil {
		log.Println(err)
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	opts := omgo.Options{
		Timezone:      "Europe/Brussels",
		HourlyMetrics: []string{"temperature_2m", "precipitation_probability", "precipitation", "weather_code", "wind_speed_10m", "wind_direction_10m"},
	}

	page := WeatherHourlyPage{
		LocationID: locStr,
		Data:       []WeatherCondition{},
		Offset:     offset + 6,
	}
	weatherLocationLock.RLock()
	locName, ok := weatherLocation[locStr]
	weatherLocationLock.RUnlock()
	if ok {
		page.Location = locName
	}
	wf, err := wc.Forecast(context.Background(), loc, &opts)
	if err != nil {
		log.Println(err)
		return c.Redirect(http.StatusFound, "/weather/location")
	}

	startIndex := offset
	endIndex := offset + 6

	for _, t := range wf.HourlyTimes {
		if time.Now().After(t) {
			startIndex++
			endIndex++
		}
	}

	if endIndex > len(wf.HourlyTimes) {
		endIndex = len(wf.HourlyTimes)
		offset = 0
	}

	for i, t := range wf.HourlyTimes {
		if i < startIndex || i >= endIndex {
			continue
		}
		page.Data = append(page.Data, WeatherCondition{
			Time:          t.Format("Mon 15:04"),
			Precipitation: fmt.Sprintf("%.1f mm", wf.HourlyMetrics["precipitation"][i]),
			Temperature:   fmt.Sprintf("%.1f°C", wf.HourlyMetrics["temperature_2m"][i]),
			WindSpeed:     fmt.Sprintf("%.1f km/h", wf.HourlyMetrics["wind_speed_10m"][i]),
			WindDirection: windDirection(wf.HourlyMetrics["wind_direction_10m"][i]),
			Icon:          wwoToOpenweathermapIcon(fmt.Sprintf("%d", int(wf.HourlyMetrics["weather_code"][i]))),
		})
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")

	return tmpl.Execute(c.Response().Writer, page)
}
