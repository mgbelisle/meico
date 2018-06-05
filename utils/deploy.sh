#!/bin/bash

set -e

go run utils/build.go
HASH=$(ipfs add -qr www | tail -n 1)
ID=$(ssh meico@meico.dance "ipfs id -f='<id>'")
TIMESTAMP=$(date +%s)
ipfs swarm connect /dnsaddr/meico.dance/ipfs/$ID
ssh meico@meico.dance /bin/bash << EOF
    set -e
    ipfs get -o ~/www-$TIMESTAMP /ipfs/$HASH
    rm -rf ~/www
    mv ~/www-$TIMESTAMP ~/www
    ipfs pin add -r $HASH
    ipfs name publish $HASH # This takes a while for some reason
EOF

# Replicate in a few spots
# for ADDR in 138.68.18.245
# do
#     ID=$(ssh ipfs@$ADDR "ipfs id -f='<id>'")
#     ipfs swarm connect /ip4/$ADDR/tcp/4001/ipfs/$ID
#     ssh ipfs@$ADDR "ipfs pin add -r $HASH"
# done
