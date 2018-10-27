#!/bin/bash

set -e

go run utils/build/main.go --out www-deploy
HASH=$(ipfs add -qr www-deploy | tail -n 1)
ID=$(ssh meico@meico.dance "ipfs id -f='<id>'")
TIMESTAMP=$(date +%s)
ipfs swarm connect /dnsaddr/meico.dance/ipfs/$ID
ipfs swarm connect /ip6/2604:a880:2:d0::741:7001/tcp/4001/ipfs/$ID
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$TIMESTAMP /ipfs/$HASH
    rm -rf ~/www
    mv ~/www-$TIMESTAMP ~/www
    ipfs pin add -r $HASH
    ipfs name publish $HASH # This takes a while for some reason
EOF
