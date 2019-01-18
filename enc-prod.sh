#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://t432z.com/enc/?url=http://test.test.com"

