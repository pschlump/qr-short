#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://t432z.com/upd/?url=http://test.test.com&id=5c"

#	"http://192.168.0.157:2004/bulkLoad"
#  	--header "Content-Type:application/json" \
