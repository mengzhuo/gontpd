#!/bin/bash

set -ue

MAKEGOBIN=${MAKEGOBIN:-`which go`}
VERSION=`git describe --tags`

rm -rf *.deb
rm -rf .buildtmp
cp -r pkg .buildtmp
mkdir -p .buildtmp/usr/bin

$MAKEGOBIN build \
    -o .buildtmp/usr/bin/gontpd \
    -ldflags "-X main.Version=$VERSION" \
    cmd/gontpd/main.go

fpm -s dir -C '.buildtmp/' -t deb -n gontpd -v $VERSION --verbose --url https://gontpd.org\
    --post-install .post-install.sh
