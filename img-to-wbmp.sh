#!/bin/bash

NEW_NAME=$(echo "$1" | cut -f 1 -d '.')



magick "$1" -resize 100 -monochrome +dither "$NEW_NAME.bmp"
#magick "$1" -resize 100 -dither FloydSteinberg -remap pattern:gray50 $NEW_NAME.bmp

magick "$NEW_NAME.bmp" -resize 100 "$NEW_NAME.wbmp"

# For dynamic png/wbmp switching code
magick "$1" -resize 100 "$NEW_NAME.png"