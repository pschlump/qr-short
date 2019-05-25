for i in  \
"5349,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5349" \
"5350,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5350" \
"5351,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5351" \
"5352,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5352" \
"5353,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5353" \
"5354,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5354" \
"5355,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5355" \
"5356,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5356" \
"5202,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5202" \
"5203,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5203" \
"5204,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5204" \
"5205,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5205" \
"5206,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5206" \
"5207,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5207" \
"5208,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5208" \
"5209,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5209" \
"5210,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5210" \
"5211,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5211" \
"5212,http:%2F%2Fwgb.beefchain.com%2Fproduct%2Fqr%2F5212" \
; do 
	id10=$(echo $i | sed -e 's/,.*//' )
	id36=$( to36/to36 "$id10" )
	url=$( echo $i | sed -e 's/.*,//' )
	wget -o out/,list1_$id10 -O out/,list2_$id10 \
		--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
		"http://t432z.com/upd/?id=${id36}&url=${url}"
done
