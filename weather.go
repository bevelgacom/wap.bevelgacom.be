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

type WeatherNow struct {
	Icon          string
	Temperature   string
	WindSpeed     string
	WindDirection string
}

type WeatherDetailPage struct {
	Location string
	Now      WeatherNow
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

	page := WeatherDetailPage{}
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

	page.Now = WeatherNow{
		Icon:          wwoToOpenweathermapIcon(fmt.Sprintf("%d", cw.WeatherCode)),
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
