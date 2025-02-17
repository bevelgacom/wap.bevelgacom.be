#!/bin/bash
magick "$1" -resize 100 -monochrome +dither output.bmp
magick output.bmp -resize 100 output.wbmp