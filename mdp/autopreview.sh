#!/bin/bash

# previous hash
PHASH=$(md5sum $1)
while true; do
	NHASH=$(md5sum $1)
	if [ "${PHASH}" != "${NHASH}" ]; then
		./mdp $1
	fi
done
