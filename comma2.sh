#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://t432z.com/upd/?id=43k&url=http%3A%2F%2Fagroledge.com%3A9025%2Fapi%2Fv1%2Fqr%3Fbase10%3D5312"

wget -o out/,list3 -O out/,list4 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://t432z.com/dec/43k?abc=22"
