#!/bin/bash

set -e

echo 'Building ...'
hugo
echo 'Adding to IPFS ...'
HASH=$(ipfs add -qr public | tail -n 1)
ipfs swarm connect /ip6/2604:a880:2:d0::741:7001/tcp/4001/ipfs/QmY3p5zRK9obUETPowTMeU7ceDmyPyQGrVwguWffBb9rMM # TODO: Better address
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
