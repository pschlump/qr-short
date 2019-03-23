
#!/bin/bash

mkdir -p ./out
LOCALIP=127.0.0.1

if [ "${QR_SHORT_AUTH_TOKEN}" == "" ] ; then
	echo "Must Set/Export QR_SHORT_AUTH_TOKEN"
	exit 1
fi

curl --header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" "http://${LOCALIP}:2004/enc/?url=http://www.2c-why.com/" >out/,t1
echo "" >>out/,t1
cat out/,t1

wget -o out/,t2.o -O out/,t2.oo "http://${LOCALIP}:2004/q/`cat out/,t1`" 

if grep 'A Lesical Analyzer Generator in Go' out/,t2.oo >/dev/null ; then
	exit 0
else 
	exit 1
fi

