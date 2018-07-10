#!/bin/bash

set -ue

MAKEGOBIN=${MAKEGOBIN:-`which go`}
echo $MAKEGOBIN

rm -rf *.deb
rm -rf .buildtmp
cp -r pkg .buildtmp
mkdir -p .buildtmp/usr/bin

$MAKEGOBIN build -o .buildtmp/usr/bin/gontpd cmd/gontpd/main.go

fpm -s dir -C '.buildtmp/' -t deb -n gontpd -v `git describe --tags` --verbose --url https://gontpd.org
