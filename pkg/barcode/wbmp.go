package barcode

import (
	"bytes"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func CreateQR(input string) []byte {
	qrCode, _ := qr.Encode(input, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 90, 90)

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	png.Encode(bufer, qrCode)

	return ImageToWBMP(bufer.Bytes())
}

func CreateAztec(input string) []byte {
	aztecCode, _ := aztec.Encode([]byte(input), 0, 0)
	aztecCode, _ = barcode.Scale(aztecCode, 90, 90)

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	png.Encode(bufer, aztecCode)

	return ImageToWBMP(bufer.Bytes())
}

func CreateCode128(input string) []byte {
	bcode, err := code128.Encode(input)
	if err != nil {
		panic(err)
	}
	bcodeScaled, err := barcode.Scale(bcode, 101, 40)
	if err != nil {
		panic(err)
	}

	bufer := bytes.NewBuffer([]byte{})

	// encode the barcode as png
	err = png.Encode(bufer, bcodeScaled)
	if err != nil {
		panic(err)
	}

	return ImageToWBMP(bufer.Bytes())
}

func ImageToWBMP(input []byte) []byte {
	imagick.Initialize()
	defer imagick.Terminate()

	tmpdir, _ := os.MkdirTemp("", "imagick")
	defer os.RemoveAll(tmpdir)

	// write the image to a file
	err := os.WriteFile(tmpdir+"/image.png", input, 0644)
	if err != nil {
		panic(err)
	}

	_, err = imagick.ConvertImageCommand([]string{"convert", tmpdir + "/image.png", "-resize", "90", "-monochrome", tmpdir + "/output.bmp"})
	if err != nil {
		panic(err)
	}
	imagick.ConvertImageCommand([]string{"convert", tmpdir + "/output.bmp", "-resize", "90", tmpdir + "/output.wbmp"})

	// read the image from the file
	output, err := os.ReadFile(tmpdir + "/output.wbmp")
	if err != nil {
		panic(err)
	}

	return output
}
