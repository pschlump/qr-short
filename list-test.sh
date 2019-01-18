#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://192.168.0.157:2004/list?beg=1&end=9999"

#	"http://192.168.0.157:2004/bulkLoad"
#  	--header "Content-Type:application/json" \
