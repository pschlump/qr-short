#!/bin/bash

PID="$( ps -ef | grep qr-short | grep -v grep | awk '{print $2}' )"
if [ -z "$PID" ] ; then
	# echo "Nothing to do"
	:
else
	# echo "Kill $PID"
	kill $PID
fi

./new-run.sh &


