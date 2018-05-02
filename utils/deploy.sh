#!/bin/bash

set -e

go run utils/build.go
HASH=$(ipfs add -qr www | tail -n 1)
NS=$(ssh meico@meico.dance "ipfs id -f='<id>'")
TIMESTAMP=$(date +%s)
ipfs swarm connect /dnsaddr/meico.dance/ipfs/$NS
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$TIMESTAMP /ipfs/$HASH
    rm -rf ~/www
    mv ~/www-$TIMESTAMP ~/www
    ipfs pin add -r $HASH
    ipfs name publish $HASH # This takes a while for some reason
EOF
