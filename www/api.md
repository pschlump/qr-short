# API for qr-short 

`qr-short` takes a short URL and redirects it to a full URL.  An example:
`http://t432z.com/q/57` might redirect to `http://app.beefchain.com/sheepchain/display?id=57`.

The advantage of using a short URL is that it makes a better QR code with less text.

`qr-short` is running on `http://t432z.com`.

# Creating (Update) for a URL.

```
	http://t432z.com/upd?id=57&url=http://app.beefchain.com/sheepchain/display?id=57
```

## /upd

Create or update an existing redirect.

Authorization Cookie, `Qr-Auth` is required.

Parameters

| Name     |  Description                                                                       |
|----------|------------------------------------------------------------------------------------|
| id       | Base 36 ID.                                                                        |
| url      | URL to set this ID to.                                                             |

Other parameters are ignored.


## /q/ID

Redirect based on the ID in the URL to the specified location.

This is what would be encoded into the QR code.





## /enc

Create or update an existing redirect.

Authorization Cookie, `Qr-Auth` is required.

Parameters

| Name     |  Description                                                                       |
|----------|------------------------------------------------------------------------------------|
| url      | URL to set this ID to.                                                             |

An id is automatically generated and is 1 larger than the current largest used id.
This is returned.



## /dec

Decode an existing ID and show where it would be redirect to without doing a redirect.

Parameters

| Name     |  Description                                                                       |
|----------|------------------------------------------------------------------------------------|
| id       | Base 36 ID.                                                                        |

The URL that it would redirect to is returned.


## /list

List a range of IDs and the URLs that they would redirect to.

Authorization Cookie, `Qr-Auth` is required.

Parameters

| Name     |  Description                                                                       |
|----------|------------------------------------------------------------------------------------|
| beg      | Base 36 ID to start from.                                                          |
| end      | Base 36 ID to end with.                                                            |

Other parameters are ignored.

Output is JSON data with a set for this range.


## /status

Return the current status of the qr-short server and the build version and date.



## /bulkLoad

An example of using bulk load:

```
wget -o out/,b1 -O out/,b2 \
	--header "X-Qr-Auth: ${QR_SHORT_AUTH_TOKEN}" \
	--post-data='update={"data":[{"url":"http://www.2c-why.com/demo3","Id":"2"},{"url":"http://www.2c-why.com/demo3","Id":"5"}]}' \
	"http://t432z.com/bulkLoad"
```

Authorization Cookie, `X-Qr-Auth` header is required.  In this case it has been exported as the environment variable `QR_SHORT_AUTH_TOKEN`.
Bulk load will accept POST requests.  The data is in a JSON format.  You can have more than 1 value in the array of `url`, `Id` (note Cap. I).
If errors occurs they are listed by Id in the JSON response data.  The Id is in base 36.


# Planned Changes

1. Allow id_b10 a base ID to be used instead of expecting all IDs in base 36.
2. Provide responses in JSON format instead of text.
3. Add a `/del` to delete stuff.
4. Count the number of redirects and include this data in the list output.
5. Allow the reset of the count of redirects, `/reset-count`.
6. Limit the size of the `/list` output to 5000 rows.







