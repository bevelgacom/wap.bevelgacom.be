package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	"github.com/bevelgacom/wap.wap.bevelgacom.be/pkg/image"
	"github.com/labstack/echo/v4"
)

func serveImage(c echo.Context) error {
	imageURL := c.QueryParam("url")
	if imageURL == "" {
		return c.String(http.StatusBadRequest, "no URL provided")
	}

	if strings.HasPrefix(imageURL, "cache:") {
		imageURL = GetLink(imageURL[6:])
		if imageURL == "" {
			return c.String(http.StatusBadRequest, "invalid cache link")
		}
	}

	// download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}
	defer resp.Body.Close()

	// check if content type is image
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "image") {
		log.Println(resp.Header.Get("Content-Type"))
		return c.String(http.StatusBadRequest, "")
	}

	buffer := bytes.NewBuffer([]byte{})
	buffer.ReadFrom(resp.Body)

	return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", image.ImageToWBMP(buffer.Bytes(), 80))
}
