#!/usr/bin/env bash
# https://guides.wp-bullet.com/batch-resize-images-using-linux-command-line-and-imagemagick/

# absolute path to image folder
FOLDER="src/img"

# max height
WIDTH=1024

# max width
HEIGHT=1024

#resize jpg only to either height or width, keeps proportions using imagemagick
find ${FOLDER} -iname '*.jpg' -exec convert \{} -verbose -resize $WIDTHx$HEIGHT\> \{} \;
find ${FOLDER} -iname '*.jpeg' -exec convert \{} -verbose -resize $WIDTHx$HEIGHT\> \{} \;
find ${FOLDER} -iname '*.png' -exec convert \{} -verbose -resize $WIDTHx$HEIGHT\> \{} \;
