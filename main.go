package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", serveHome)
	e.GET("/wap/*", serveWAP)
	e.GET("/dl/*", serveDL)
	e.GET("/nws/list", serveNewsList)
	e.GET("/nws/item", serveNewsItem)

	e.Start(":80")
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
	_, file := path.Split(req.URL.Path)
	f, err := os.Open("./static/dl/" + file)

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
	_, file := path.Split(req.URL.Path)
	f, err := os.Open("./static/" + file)

	if err == os.ErrNotExist {
		return c.String(http.StatusNotFound, "")
	} else if err != nil {
		log.Panicln(err)
		return c.String(http.StatusInternalServerError, "")
	}

	return c.Stream(http.StatusOK, "text/vnd.wap.wml", f)
}
