package barcode

import (
	"bytes"
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func CreateQR(input string, size int64) []byte {
	qrCode, _ := qr.Encode(input, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, int(size), int(size))

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	png.Encode(bufer, qrCode)

	return ImageToWBMP(bufer.Bytes(), size)
}

func CreateAztec(input string, size int64) []byte {
	aztecCode, _ := aztec.Encode([]byte(input), 0, 0)
	aztecCode, _ = barcode.Scale(aztecCode, int(size), int(size))

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	png.Encode(bufer, aztecCode)

	return ImageToWBMP(bufer.Bytes(), size)
}

func CreateCode128(input string, size int64) []byte {
	bcode, err := code128.Encode(input)
	if err != nil {
		panic(err)
	}
	bcodeScaled, err := barcode.Scale(bcode, 101, int(size))
	if err != nil {
		panic(err)
	}

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	err = png.Encode(bufer, bcodeScaled)
	if err != nil {
		panic(err)
	}

	return ImageToWBMP(bufer.Bytes(), size)
}

func ImageToWBMP(input []byte, size int64) []byte {
	imagick.Initialize()
	defer imagick.Terminate()

	tmpdir, _ := os.MkdirTemp("", "imagick")
	defer os.RemoveAll(tmpdir)

	// write the image to a file
	err := os.WriteFile(tmpdir+"/image.png", input, 0644)
	if err != nil {
		panic(err)
	}

	_, err = imagick.ConvertImageCommand([]string{"convert", tmpdir + "/image.png", "-resize", fmt.Sprintf("%d", size), "-monochrome", tmpdir + "/output.bmp"})
	if err != nil {
		panic(err)
	}
	imagick.ConvertImageCommand([]string{"convert", tmpdir + "/output.bmp", "-resize", fmt.Sprintf("%d", size), tmpdir + "/output.wbmp"})

	// read the image from the file
	output, err := os.ReadFile(tmpdir + "/output.wbmp")
	if err != nil {
		panic(err)
	}

	return output
}
