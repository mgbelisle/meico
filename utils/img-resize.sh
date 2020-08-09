#!/usr/bin/env bash
# https://guides.wp-bullet.com/batch-resize-images-using-linux-command-line-and-imagemagick/

# absolute path to image folder
FOLDER="$HOME/Desktop"

# max width
WIDTH=1024

# max height
HEIGHT=1024

# resize jpg only to either height or width, keeps proportions using imagemagick
find $FOLDER -iname '*.jp*g' -exec convert '{}' -verbose -resize $WIDTHx$HEIGHT\> '{}' \;
find $FOLDER -iname '*.png' -exec convert '{}' -verbose -resize $WIDTHx$HEIGHT\> '{}' \;
