#!/bin/bash

set -e

hugo
HASH=$(ipfs add -qr public | tail -n 1)
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$HASH /ipfs/$HASH
    ipfs pin add -r $HASH
    echo 'Publishing via IPNS. Not sure why this takes so long ...'
    ipfs name publish $HASH
    [ -e '~/www' ] && mv ~/www ~/www-pre-$HASH
    mv ~/www-$HASH ~/www
    [ -e "~/www-pre-$HASH" ] && rm -r ~/www-pre-$HASH
EOF
echo Deployed $HASH
