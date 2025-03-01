package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", serveHome)
	e.GET("/wap/*", serveWAP)
	e.GET("/dl/*", serveDL)

	e.GET("/navigator/*", serveNavigator)
	e.GET("/navigator/query", serveNavigatorQuery)

	e.GET("/nws/list", serveNewsList)
	e.GET("/nws/item", serveNewsItem)

	e.GET("/barcode/*", serveBarcode)
	e.GET("/barcode/barcode", serveBarcodePage)
	e.GET("/barcode/image.wbmp", serveBarcodeImage)
	e.GET("/png-convert.wbmp", serveImage)

	e.GET("/weather/location", serveWeatherLocation)
	e.GET("/weather/details", serveWeatherDetailView)
	e.GET("/weather/hourly", serveWeatherHourly)

	e.Start(":8080")
}

func serveHome(c echo.Context) error {
	f, err := os.Open("./static/home.wml")

	if err == os.ErrNotExist {
		c.String(http.StatusNotFound, "")
	} else if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	return c.Stream(http.StatusOK, "text/vnd.wap.wml", f)
}

func serveDL(c echo.Context) error {
	c.Path()
	req := c.Request()
	if req == nil {
		return c.String(http.StatusInternalServerError, "")
	}
	r, err := os.OpenRoot("./static/dl")
	if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	file := strings.TrimPrefix(req.URL.Path, "/dl/")
	f, err := r.Open(file)

	if err == os.ErrNotExist {
		return c.String(http.StatusNotFound, "")
	} else if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	// Go is incorrect when judging Java files
	// this overrides that
	contentType := "binary/octet-stream"
	if strings.HasSuffix(file, ".jar") {
		contentType = "application/java-archive"
	}

	return c.Stream(http.StatusOK, contentType, f)
}

func serveWAP(c echo.Context) error {
	c.Path()
	req := c.Request()
	if req == nil {
		return c.String(http.StatusInternalServerError, "")
	}
	file := strings.TrimPrefix(req.URL.Path, "/wap/")
	r, err := os.OpenRoot("./static/")
	if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	f, err := r.Open(file)

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
