#!/bin/bash

wget -o out/,b1 -O out/,b2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	--post-data='update={"data":[{"url":"http://www.2c-why.com/demo3","2"},{"url":"http://www.2c-why.com/demo3","5"}]}' \
	"http://t432z.com/bulkLoad"


#	"http://192.168.0.157:2004/bulkLoad"

#  	--header "Content-Type:application/json" \
