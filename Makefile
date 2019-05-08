
# to run this test you will need to chagne this to your local system IP address.
# Philip's Class A block - dev8
# LOCALIP=10.240.10.18
# Philip's Class C block - dev system 1 
# LOCALIP=192.168.0.157
LOCALIP=127.0.0.1

# -v --cookie "USER_TOKEN=Yes"

all:
	../bin/cmp-local.sh

build_linux:
	../bin/cmp-prod.sh qr-short.linux

# ---------------------------------------------------------------------------------------
# Primary local test
# ---------------------------------------------------------------------------------------

test: setup test1 testSuccess

setup:
	@./kill-current.sh
	@rm -f qr-short
	@go build
	@./qr-short -note "Test-QR-Short" &

test1:
	./test1.sh

testSuccess:
	@echo PASS

test_data1:
	@mkdir -p ./out
	@curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:2004/enc/?url=http://www.2c-why.com/&data=1234" >out/,t1
	@echo "" >>out/,t1
	@curl "http://${LOCALIP}:2004/q/`cat out/,t1`" >out/,t2
	@grep 'Redirect' out/,t2


# ---------------------------------------------------------------------------------------
# Local test with forward proxy
# ---------------------------------------------------------------------------------------

loc_tst: setup lct_1 testSuccess

lct_1:
	@mkdir -p ./out
	@curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:9001/enc/?url=http://www.2c-why.com/" >out/,t4
	@echo "" >>out/,t4
	@curl "http://${LOCALIP}:9001/q/`cat out/,t4`" >out/,t5
	@grep 'Redirect' out/,t5




# ---------------------------------------------------------------------------------------
# Local Update test with forward proxy
# ---------------------------------------------------------------------------------------

loc_upd_tst: setup lct_upd1 testSuccess

lct_upd1:
	mkdir -p ./out
	echo "Set to initial value"
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:9001/enc?url=http://www.2c-why.com/" >out/,u4
	echo "" >>out/,u4
	curl "http://${LOCALIP}:9001/q/`cat out/,u4`" >out/,u5
	grep 'Redirect' out/,u5
	cat out/,u5 | sed -e 's/.*href="//' -e 's/".*//' >out/,url5
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:9001/upd?url=http://www.2c-why.com/demo32&id=`cat out/,u4`" >out/,u6
	echo "" >>out/,u6
	curl "http://${LOCALIP}:9001/q/`cat out/,u4`" >out/,u7
	grep 'Redirect' out/,u7
	cat out/,u7
	grep 'demo32' out/,u7



# ---------------------------------------------------------------------------------------
# Test with remote URL using http://t432z.com and forward proxy on server.
# ---------------------------------------------------------------------------------------

rmt_tst: rmt_1 testSuccess

rmt_1:
	@mkdir -p ./out
	@curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/enc/?url=http://www.2c-why.com/" >out/,r4
	@echo "" >>out/,r4
	@curl "http://t432z.com/q/`cat out/,r4`" >out/,r5
	@grep 'Redirect' out/,r5


# ---------------------------------------------------------------------------------------
# Test QR code update of redirect.
# ---------------------------------------------------------------------------------------

rmt_upd_tst: rmt_upd1 testSuccess

rmt_upd1:
	mkdir -p ./out
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/enc/?url=http://www.2c-why.com/" >out/,ru4
	echo "" >>out/,ru4
	curl "http://t432z.com/q/`cat out/,ru4`" >out/,ru5
	grep 'Redirect' out/,ru5
	cat out/,u5 | sed -e 's/.*href="//' -e 's/".*//' >out/,rurl5
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/upd?url=http://www.2c-why.com/demo32&id=`cat out/,ru4`" >out/,ru6
	echo "" >>out/,ru6
	curl "http://t432z.com/q/`cat out/,ru4`" >out/,ru7
	grep 'Redirect' out/,ru7
	cat out/,ru7
	grep 'demo32' out/,ru7

# ---------------------------------------------------------------------------------------
# Test QR remote update #3 to be "http://www.2c-why.com/demo32" using the 
# remote proxy.
# ---------------------------------------------------------------------------------------

rmt_upd3:
	mkdir -p ./out
	echo "3" >out/,ru4
	curl "http://t432z.com/q/`cat out/,ru4`" >out/,ru5
	grep 'Redirect' out/,ru5
	cat out/,u5 | sed -e 's/.*href="//' -e 's/".*//' >out/,rurl5
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/upd?url=http://www.2c-why.com/demo32&id=`cat out/,ru4`" >out/,ru6
	echo "" >>out/,ru6
	curl "http://t432z.com/q/`cat out/,ru4`" >out/,ru7
	grep 'Redirect' out/,ru7
	cat out/,ru7
	grep 'demo32' out/,ru7

# Setup for China
set59:
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/upd?url=http://www.2c-why.com/demo34?id=59&id=59" 


set62:
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/upd?url=https://beefchain.com/ranches/murraymere-ranch?id=62&id=62" 

set1q:
	curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://t432z.com/upd?url=https://beefchain.com/ranches/murraymere-ranch?id=1q&id=1q" 


# ---------------------------------------------------------------------------------------
# Test local usin "list" command.
# ---------------------------------------------------------------------------------------
loc_list1_tst: loc_list1 testSuccess

loc_list1:
	@curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:2004/list?beg=0&end=last" -o out/,l1
	@grep Id out/,l1 >/dev/null
	@grep URL out/,l1 >/dev/null
	@grep Count out/,l1 >/dev/null

#xx:
#	-wget -o out/,l1 -O out/,l2 --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:2004/list?beg=0&end=last" 

loc_list2:
	@curl --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:9001/list?beg=0&end=last" -o out/,l1

# ---------------------------------------------------------------------------------------
# ---------------------------------------------------------------------------------------
bulk_load1: testSuccess
	wget -o out/,b1 -O out/,b2 --post-file testdata/bl1.txt --header "X-Qr-Auth: `env | grep QR_SHORT_AUTH_TOKEN | sed -e 's/.*=//'`" "http://${LOCALIP}:2004/bulkLoad"


check_status:
	wget -o out/,s1 -O out/,s2 http://192.168.0.157:2004/status
	@cat out/,s2
