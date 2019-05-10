#!/bin/bash

if [ "$QR_SHORT_AUTH_TOKEN" == "" ] ; then
	echo "Environment not set correctly, export QR_SHORT_AUTH_TOKEN" 
	exit 1
fi
if [ "$REDIS_HOST" == "" ] ; then
	echo "Environment not set correctly, export REDIS_HOST" 
	exit 1
fi

PID="$( ps -ef | grep qr-short | grep -v grep | awk '{print $2}' )"
if [ -z "$PID" ] ; then
	# echo "Nothing to do"
	:
else
	# echo "Kill $PID"
	kill $PID
fi

(
while true ; do
    date
	if [ -f ./set-env.sh ] ; then
		. ./set-env.sh
	fi
    ./qr-short >,log 2>&1
    date
    sleep 1
done
) 2>&1 > /tmp/qr-micro-service.2.out &
