#!/bin/bash

#RUNNING=$( ps -ef | grep "/bin/bash.*run-chat2.sh" | grep -v grep  )
## echo "RUNNING=$RUNNING"
#if [ "$RUNNING" == "" ] ; then
#   :
#else
#   echo "RUNNING [$RUNNING]"
#   echo "chat2 is running now."
#   exit 0
#fi

while true ; do
    date
    # ./chat2
    ./qr-short >,log 2>&1
    date
    sleep 1
    # exit 0
done

