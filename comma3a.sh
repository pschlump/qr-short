
wget -o out/,list3 -O out/,list4 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	"http://t432z.com/dec/40i"

cat out/,list3 out/,list4
