
Use a CSV file to create/setup QR Codes

3 Columns in the CSV file

1. The base10 version of the ID or a 'g' to indicate that the qr-micro-service should be used to allocate the next availabe QR code.

2. The base36 versio of the id or omitted.  You need to specify one of 'g', base10, or base36.

3. The Destination URL with templating to send this to.

Example:

```
3,,http://www.2c-why.com/demo3?id36={{.id36}}
,400,http://www.2c-why.com/demo3?id36={{.id36}}&id10={{.id10}}&tom=jerry
g,,http://www.2c-why.com/demo3?id36={{.id36}}&id10={{.id10}}&generated=true
```

On the first line set the QR with an id of 3 (base 10) to `http://www.2c-why.com/demo3?id36=3`.

ON the 2nd line set the QR with a base 36 id of 400 to `http://www.2c-why.com/demo3?id36=400&id10=5184&tom=jerry`.

On the 3rd line, use the generation capability to set  the URL with the next available QR code.

