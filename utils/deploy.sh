#!/bin/bash

set -e

ID=$(ssh meico@meico.dance "ipfs id -f='<id>'")
ipfs swarm connect /ip6/2604:a880:2:d0::741:7001/tcp/4001/ipfs/$ID
ipfs swarm connect /dnsaddr/meico.dance/ipfs/$ID
go run utils/build/main.go --out www-deploy
HASH=$(ipfs add -qr www-deploy | tail -n 1)
TIMESTAMP=$(date +%s)
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$TIMESTAMP /ipfs/$HASH
    rm -rf ~/www
    mv ~/www-$TIMESTAMP ~/www
    ipfs pin add -r $HASH
    ipfs name publish $HASH # This takes a while for some reason
EOF
