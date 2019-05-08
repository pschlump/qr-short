#!/bin/bash

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://127.0.0.1:2004/upd/?id=43j&url=http%3A%2F%2Fagroledge.com%3A9025%2Fapi%2Fv1%2Fqr%3Fbase10%3D5311"

wget -o out/,list3 -O out/,list4 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://127.0.0.1:2004/dec/43j?abc=22"
