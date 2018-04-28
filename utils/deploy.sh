#!/bin/bash

set -e

hugo
HASH=$(ipfs add -qr public | tail -n 1)
NS=$(ssh meico@meico.dance "ipfs id -f='<id>'")
TIMESTAMP=$(date +%s)
ipfs swarm connect /dnsaddr/meico.dance/ipfs/$NS
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$TIMESTAMP /ipfs/$HASH
    rm -rf ~/www
    mv ~/www-$TIMESTAMP ~/www
    ipfs pin add -r $HASH
    echo 'Publishing ...'
    ipfs name publish $HASH
EOF
