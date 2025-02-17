package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/bevelgacom/wap.wap.bevelgacom.be/pkg/barcode"
	"github.com/labstack/echo/v4"
)

type barcodeContent struct {
	Type    string
	Content string
	Size    string
}

func serveBarcode(c echo.Context) error {
	p := c.Request().URL.Path

	_, file := path.Split(p)
	if file == "" || file == "barcode" {
		file = "index.wml"
	}

	f, err := os.Open("./static/barcode/" + file)

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

func serveBarcodePage(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("./static/barcode/barcode.wml"))
	content := base64.StdEncoding.EncodeToString([]byte(c.QueryParam("c")))

	pageContent := barcodeContent{
		Type:    c.QueryParam("t"),
		Size:    c.QueryParam("s"),
		Content: content,
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")

	return tmpl.Execute(c.Response().Writer, pageContent)
}

func serveBarcodeImage(c echo.Context) error {
	content, err := base64.StdEncoding.DecodeString(c.QueryParam("c"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid content")
	}

	sizeStr := c.QueryParam("s")
	if sizeStr == "" {
		sizeStr = "60"
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid size")
	}

	if c.QueryParam("t") == "qr" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateQR(string(content), size))
	} else if c.QueryParam("t") == "aztec" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateAztec(string(content), size))
	} else if c.QueryParam("t") == "code128" {
		return c.Blob(http.StatusOK, "image/vnd.wap.wbmp", barcode.CreateCode128(string(content), size))
	}

	return nil
}
