#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://192.168.0.157:2004/enc/?url=http://test.test.com"

