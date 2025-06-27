#! /bin/sh

set -e

cd /opt/florist/test

for t in *.test; do
    echo
    echo ===== $t =====
    #./$t -test.v
    ./$t
done