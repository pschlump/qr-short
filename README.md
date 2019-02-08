# Qr-short

This is a shorter specifically designed for use with QR codes.  It will work with
the iPhone camera and with a number of Andoid apps for scanning QR codes.

## Qr-Short: URL Shortener Service

This service encodes URL in base-36 and store them in filesystem.

It has 3 features: shorten, decode short URL, and redirect.

#### Example URL to Encode

```
	http://t432z.com/enc/?url=http://www.beefchain.com/qr-app?id=152211241231
```

Say the "output" is 2, then

### Example decode short URL

```
	http://t432z.com/dec/2
```

The output should be `http://www.beefchain.com/qr-app/?id=152211241231`


### Example redirect

```
	http://t432z.com/q/2
```

Will produce a 307 temporary redirect to the destination URL.

### Example QR that redirects to a site

If you take a picture of this QR with an iPhone or an Android app for QR codes
it should take you to http://www.2c-why.com/ . The text encode in the QR code is http://t432z.com/q/3 .

![QR Code for http://www.2c-why.com](https://github.com/American-Certified-Brands/tools/blob/master/qr-short/sample-qr/to-2c-why.com.png?raw=true "QR Code to demo proxy and go to http://www.2c-why.com/")


# API

http://t432z.com/index.html is an interactive app for setting/updating/listing the QR codes.


