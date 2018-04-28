#!/bin/bash

set -e

echo 'Building ...'
hugo --quiet
echo 'Adding to IPFS ...'
HASH=$(ipfs add -qr public | tail -n 1)
ssh meico@meico.dance /bin/bash << EOF
    set -e
    echo 'Getting ...'
    ipfs get -o ~/www-$HASH /ipfs/$HASH
    echo 'Pinning ...'
    ipfs pin add -r $HASH
    echo 'Publishing ...'
    ipfs name publish $HASH
    echo 'Deploying ...'
    [ -e '~/www' ] && mv ~/www ~/www-pre-$HASH
    mv ~/www-$HASH ~/www
    [ -e "~/www-pre-$HASH" ] && rm -r ~/www-pre-$HASH
EOF
echo Deployed $HASH
