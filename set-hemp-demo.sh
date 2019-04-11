#!/bin/bash

#
# Setup 2 QR codes to re-direct to www.agroledge.com - for demo in Casper WY
#


if [ "${QR_SHORT_AUTH_TOKEN}" == "" ] ; then
	echo must set auth token
	exit 1
fi

# local
# export SER="http://192.168.0.157:2004"

# prod
export SER="http://t432z.com"

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"${SER}/upd/?url=http://www.agroledge.com/qr-setup.html&id=1z4"

wget -o out/,list1 -O out/,list2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"${SER}/upd/?url=http://www.agroledge.com/qr-final.html&id=1z5"
