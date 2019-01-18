#!/bin/bash

cd /Users/corwin/go/src/github.com/American-Certified-Brands/tools/qr-short

# build/re-run script for development.

rm -f ./qr-short

if [ "$( hostname )" == "pschlump-dev2" ] ; then
	go build 2>&1 | color-cat -c red
else
	go build
fi

if [ -f qr-short ] ; then
	:
else
	echo "Failed to build qr-short."
	exit 1
fi

xx=$( ps -ef | grep QR-SHORT | grep -v grep | grep -v qr-short.sh | grep "$NN" | awk '{print $2}' )
if [ "X$xx" == "X" ] ; then	
	:
else
	kill $xx
fi

if [ -x ./qr-short ] ; then
	# echo Running It
	# ./qr-short --note "QR-SHORT" --debug storage.db5 &
	./qr-short --note "QR-SHORT" --debug db1 &
fi

