#!/bin/bash

set -e

echo 'Building ...'
hugo
echo 'Adding to IPFS ...'
HASH=$(ipfs add -qr public | tail -n 1)
ipfs swarm connect /dnsaddr/meico.dance/ipfs/QmY3p5zRK9obUETPowTMeU7ceDmyPyQGrVwguWffBb9rMM
ssh meico@meico.dance /bin/bash << EOF
    set -e
    echo 'Getting from IPFS ...'
    ipfs get -o ~/www /ipfs/$HASH
    echo 'Pinning ...'
    ipfs pin add -r $HASH
    echo 'Publishing ...'
    ipfs name publish $HASH
EOF
echo Deployed $HASH
