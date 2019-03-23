#!/bin/bash

PID="$( ps -ef | grep qr-short | grep -v grep | grep "Test-QR-Short" | awk '{print $2}' )"
if [ -z "$PID" ] ; then
	# echo "Nothing to do"
	:
else
	# echo "Kill $PID"
	kill $PID
fi

