#!/bin/bash

set -e

hugo
HASH=$(ipfs add -qr public | tail -n 1)
ssh meico@206.189.70.13 ipfs pin add -r $HASH
ssh meico@206.189.70.13 ipfs name publish $HASH
