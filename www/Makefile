
all: api.html 


FR=./md

api.html: api.md
	markdown-cli --input=./api.md --output=api.html
	cat ${FR}/pre ${FR}/markdown.css ${FR}/post ${FR}/md.css ${FR}/hpre api.html ${FR}/hpost >/tmp/api.html
	mv /tmp/api.html ./api.html

