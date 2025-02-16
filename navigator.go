package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bevelgacom/wap.wap.bevelgacom.be/pkg/dbnav"
	"github.com/labstack/echo/v4"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/scizorman/go-ndjson"
)

const dbTime = "2006-01-02T15:04:05-07:00"

var tz *time.Location

var nav *dbnav.Client

type Station struct {
	Id     string  `json:"id"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type Leg struct {
	From              string
	To                string
	DepartureTime     string
	DeparturePlatform string

	ArrivalTime     string
	ArrivalPlatform string

	Line string
}

type Connection struct {
	Id            int
	DepartureTime string
	ArrivalTime   string
	Changes       int

	From string
	To   string

	Legs []Leg
}

type queryPage struct {
	From  *Station
	To    *Station
	Modes string

	FromValue string
	ToValue   string

	Date string
	Time string

	FromList []Station
	ToList   []Station

	Connections []Connection
}

var Stations map[string]Station = make(map[string]Station)

func init() {
	apiURL := os.Getenv("NAVIGATOR_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:3000"
	}
	var err error
	nav, err = dbnav.NewClient(apiURL)
	if err != nil {
		log.Panicln(err)
	}

	f, err := os.Open("./hafas-stations.ndjson")
	if err != nil {
		log.Panicln(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		log.Panicln(err)
	}
	stationList := [][]any{}
	err = ndjson.Unmarshal(data, &stationList)
	if err != nil {
		log.Panicln(err)
	}

	for _, s := range stationList {
		id := s[0].(string)
		name := s[1].(string)
		weight := s[2].(float64)

		Stations[id] = Station{
			Id:     id,
			Name:   name,
			Weight: weight,
		}
	}

	tz, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Panicln(err)
	}
}

func seachStation(q string) []Station {
	result := []Station{}
	match := map[string]int{}
	for _, s := range Stations {
		if s.Weight < 0.5 {
			continue
		}
		if strings.EqualFold(q, s.Name) {
			return []Station{s}
		}
		if score := fuzzy.RankMatch(q, s.Name); score > 0 && score < 10 {
			result = append(result, s)
			match[s.Id] = score
		}
	}
	slices.SortFunc(result, func(a, b Station) int {
		return int(match[a.Id] - match[b.Id])
	})
	if len(result) > 10 {
		result = result[:10]
	}
	// sort on weight and match score
	slices.SortFunc(result, func(a, b Station) int {
		as := a.Weight - float64(match[a.Id]*30)
		bs := b.Weight - float64(match[b.Id]*30)
		return int(bs - as)
	})

	for _, s := range result {
		log.Println(s.Name, s.Weight, match[s.Id])
	}

	return result
}

func serveNavigator(c echo.Context) error {
	p := c.Request().URL.Path

	_, file := path.Split(p)
	if file == "" || file == "navigator" {
		file = "index.wml"
	}

	f, err := os.Open("./static/navigator/" + file)

	if err == os.ErrNotExist {
		return c.String(http.StatusNotFound, "")
	} else if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	mime := "text/vnd.wap.wml"
	if strings.HasSuffix(file, ".wbmp") {
		mime = "image/vnd.wap.wbmp"
	}

	return c.Stream(http.StatusOK, mime, f)
}

func serveNavigatorQuery(c echo.Context) error {
	advanced := c.QueryParam("q")
	var tmpl *template.Template
	if advanced == "true" {
		tmpl = template.Must(template.ParseFiles("./static/navigator/query-advanced.wml"))
	} else {
		tmpl = template.Must(template.ParseFiles("./static/navigator/query.wml"))
	}

	from := c.QueryParam("s") // these are the original HAFAS WAP query parameters
	to := c.QueryParam("z")

	dateStr := c.QueryParam("datum")
	timeStr := c.QueryParam("zeit")

	now := time.Now().In(tz)
	if dateStr == "" {
		// set to DDMMYY
		dateStr = now.Format("020106")
	}

	if timeStr == "" {
		// set to HHMM
		timeStr = now.Format("1504")
	}

	pageData := queryPage{
		FromValue: from,
		ToValue:   to,

		Date: dateStr,
		Time: timeStr,
	}

	if from != "" {
		if _, err := strconv.ParseInt(from, 10, 64); err != nil {
			// we got a search
			res := seachStation(from)
			if len(res) == 1 { // if we only have one result, we can skip the search page
				pageData.From = &res[0]
			} else if len(res) > 0 {
				pageData.FromList = res
			}
		} else {
			s, ok := Stations[from]
			if !ok {
				return c.String(http.StatusBadRequest, "Invalid from station")
			}
			pageData.From = &s
		}
	}

	if to != "" {
		if _, err := strconv.ParseInt(to, 10, 64); err != nil {
			// we got a search
			res := seachStation(to)
			if len(res) == 1 { // if we only have one result, we can skip the search page
				pageData.To = &res[0]
			} else if len(res) > 0 {
				pageData.ToList = res
			}
		} else {
			s, ok := Stations[to]
			if !ok {
				return c.String(http.StatusBadRequest, "Invalid to station")
			}
			pageData.To = &s
		}
	}

	if pageData.From != nil && pageData.To != nil {
		// parse date and time into one time.Time
		date, err := time.ParseInLocation("0201061504", pageData.Date+pageData.Time, tz)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid date or time")
		}

		resp, err := nav.GetJourneys(context.Background(), &dbnav.GetJourneysParams{
			From:      &pageData.From.Id,
			To:        &pageData.To.Id,
			Departure: &date,
		})
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
		data, err := dbnav.ParseGetJourneysResponse(resp)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
		if data.JSON2XX == nil {
			log.Println(string(data.Body))
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		for i, journey := range data.JSON2XX.Journeys {
			if i > 5 {
				break
			}
			if journey.Legs == nil {
				continue
			}
			legs := *journey.Legs

			conn := Connection{
				Id: i + 1,
			}

			plannedDeparture, err := time.Parse(dbTime, *legs[0].PlannedDeparture)
			if err != nil {
				log.Println(err)
				continue
			}
			plannedArrival, err := time.Parse(dbTime, *legs[len(legs)-1].PlannedArrival)
			if err != nil {
				log.Println(err)
				continue
			}

			conn.DepartureTime = plannedDeparture.In(tz).Format("15:04")
			conn.ArrivalTime = plannedArrival.In(tz).Format("15:04")

			if legs[0].DepartureDelay != nil && *legs[0].DepartureDelay != 0 {
				conn.DepartureTime += fmt.Sprintf(" +%d", int(*legs[0].DepartureDelay))
			}
			if legs[len(legs)-1].ArrivalDelay != nil && *legs[len(legs)-1].ArrivalDelay != 0 {
				conn.ArrivalTime += fmt.Sprintf(" +%d", int(*legs[len(legs)-1].ArrivalDelay))
			}

			for legNum, leg := range legs {
				if leg.Walking != nil && *leg.Walking {
					continue
				}
				var origin, dest dbnav.Station
				if leg.Origin != nil {
					origin, _ = leg.Origin.AsStation()
				}
				if leg.Destination != nil {
					dest, _ = leg.Destination.AsStation()
				}

				if legNum == 0 {
					conn.From = *origin.Name
				}
				if legNum == len(legs)-1 {
					conn.To = *dest.Name
				}

				newleg := Leg{
					From: *origin.Name,
					To:   *dest.Name,
				}

				if leg.DepartureDelay != nil && *leg.DepartureDelay != 0 {
					newleg.DepartureTime += fmt.Sprintf(" +%d", int(*leg.DepartureDelay))
				}

				if leg.ArrivalDelay != nil && *leg.ArrivalDelay != 0 {
					newleg.ArrivalTime += fmt.Sprintf(" +%d", int(*leg.ArrivalDelay))
				}

				if leg.DeparturePlatform != nil {
					newleg.DeparturePlatform = *leg.DeparturePlatform
				}

				if leg.ArrivalPlatform != nil {
					newleg.ArrivalPlatform = *leg.ArrivalPlatform
				}

				if leg.PlannedDeparture != nil {
					depTime, _ := time.Parse(dbTime, *leg.PlannedDeparture)
					newleg.DepartureTime = depTime.In(tz).Format("15:04")
				}

				if leg.PlannedArrival != nil {
					arrTime, _ := time.Parse(dbTime, *leg.PlannedArrival)
					newleg.ArrivalTime = arrTime.In(tz).Format("15:04")
				}

				if leg.Line != nil {
					newleg.Line = *leg.Line.Name
				}

				conn.Legs = append(conn.Legs, newleg)
			}

			conn.Changes = len(legs) - 1

			pageData.Connections = append(pageData.Connections, conn)
		}

		tmpl = template.Must(template.ParseFiles("./static/navigator/list.wml"))
		// we have everything we need for a results page
		c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
		c.Response().Header().Set("Cache-Control", "no-cache, must-revalidate")

		err = tmpl.Execute(c.Response().Writer, pageData)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
		return nil
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
	c.Response().Header().Set("Cache-Control", "no-cache, must-revalidate")

	return tmpl.Execute(c.Response().Writer, pageData)
}
