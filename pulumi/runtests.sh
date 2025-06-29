#! /bin/sh

set -e

apt-get update > /dev/null
apt-get install gotestsum > /dev/null

cd /opt/florist/test

#export GOVERSION=1.24
for t in *.test; do
    echo
    echo ===== $t =====
    #./$t -test.v
    #./$t
    #
    # https://github.com/gotestyourself/gotestsum/blob/main/.project/docs/running-without-go.md
    # https://pkg.go.dev/cmd/test2json
    # -test.v is needed for test2json
    # -t: add timestamps
    # -p: set the package name
    gotestsum --raw-command -- ./test2json -t -p $t ./$t -test.v=test2json
done

