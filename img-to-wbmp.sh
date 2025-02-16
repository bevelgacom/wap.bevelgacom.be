#!/bin/bash
magick "$1" -resize 120 -monochrome +dither output.bmp
magick output.bmp -resize 120 output.wbmp