package image

import (
	"fmt"
	"os"

	"gopkg.in/gographics/imagick.v3/imagick"
)

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

	_, err = imagick.ConvertImageCommand([]string{"convert", tmpdir + "/image.png", "-resize", fmt.Sprintf("%d", size), "-dither", "FloydSteinberg", "-remap", "pattern:gray50", tmpdir + "/output.bmp"})
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

func ImageToJPEG(input []byte, size int64) []byte {
	imagick.Initialize()
	defer imagick.Terminate()

	tmpdir, _ := os.MkdirTemp("", "imagick")
	defer os.RemoveAll(tmpdir)

	// write the image to a file
	err := os.WriteFile(tmpdir+"/image.png", input, 0644)
	if err != nil {
		panic(err)
	}

	_, err = imagick.ConvertImageCommand([]string{"convert", tmpdir + "/image.png", "-resize", fmt.Sprintf("%d", size), "-quality", "15%", tmpdir + "/output.jpeg"})
	if err != nil {
		panic(err)
	}

	// read the image from the file
	output, err := os.ReadFile(tmpdir + "/output.jpeg")
	if err != nil {
		panic(err)
	}

	return output
}
