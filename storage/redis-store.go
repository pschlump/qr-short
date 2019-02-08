package storage

// Copyright (C) Philip Schlump 2018-2019.

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/redis"
)

// RedisStore implements PersistentData storage with Redis as the database.
// This will support concurrent servers on multiple machines.
type RedisStore struct {
	Root        string
	RedisHost   string
	RedisPort   string
	RedisAuth   string
	RedisPrefix string // defauilt "qr:<ID>" and "qr!seq"
	Log         *os.File
	CountHits   bool
	redisConn   *redis.Client
}

// NewRedisStore creates a connection to Redis and initialized redis if necessary.
func NewRedisStore(rHost, rPort, rAuth, rPrefix string, countHits bool, log *os.File) (rv PersistentData, err error) {
	if rHost == "" {
		rHost = "127.0.0.1"
	}
	if rPort == "" {
		rPort = "6379"
	}
	if rPrefix == "" {
		rPrefix = "qr"
	}
	rs := &RedisStore{
		RedisHost:   rHost,
		RedisPort:   rPort,
		RedisAuth:   rAuth,
		RedisPrefix: rPrefix,
		Log:         log,
		CountHits:   countHits,
	}
	rc, err := ConnectToRedis(rs.RedisHost, rs.RedisPort, rs.RedisAuth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to redis: %s\n", err)
		os.Exit(1)
	}
	rs.redisConn = rc
	seq, err := rs.redisConn.Cmd("GET", rs.RedisPrefix+"!seq").Str()
	if err != nil || seq == "" {
		err := rs.redisConn.Cmd("SET", rs.RedisPrefix+"!seq", "1").Err
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to initialize redis '%s!seq': %s\n", rs.RedisPrefix, err)
			os.Exit(1)
		}
	}
	return rs, nil
}

// NextID returns the next higher integer that will be used to lookup the URL.  It is in base 36.
func (rs *RedisStore) NextID() string {
	nn, err := rs.redisConn.Cmd("INCR", rs.RedisPrefix+"!seq").Int()
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
		return ""
	}
	return strconv.FormatUint(uint64(nn), 36) // Base 36, take count of # of files add 1, this is the code.
}

// getID returns the current sequence value in integer format.
func (rs *RedisStore) getID() int64 {
	nn, err := rs.redisConn.Cmd("GET", rs.RedisPrefix+"!seq").Int64()
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
		return 0
	}
	rv := strconv.FormatUint(uint64(nn), 36) // Base 36, take count of # of files add 1, this is the code.
	if db6 {
		fmt.Printf("getID = %s/%d AT: %s\n", rv, nn, godebug.LF())
	}
	return nn
}

// setID takes the integer form of the seqeunce and sets it.
func (rs *RedisStore) setID(newID int64) error {
	if db6 {
		fmt.Printf("setID = %d AT: %s\n", newID, godebug.LF())
	}
	err := rs.redisConn.Cmd("SET", rs.RedisPrefix+"!seq", fmt.Sprintf("%d", newID)).Err
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
		return err
	}
	return nil
}

// Insert writes out the `urlStr` into the `~/data` direcotry under the file name in `code`
func (rs *RedisStore) Insert(urlStr string) (string, error) {
	code := rs.NextID()
	return rs.insertInternal(urlStr, code)
}

// insertInternal is the implemntation of an insert without the generation of a new ID.
func (rs *RedisStore) insertInternal(urlStr, code string) (string, error) {
	err := rs.redisConn.Cmd("SET", rs.RedisPrefix+":"+code, urlStr).Err
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: unable to update: %s\n", err)
		return code, err
	}
	if rs.CountHits {
		err := rs.redisConn.Cmd("SET", rs.RedisPrefix+"^"+code, "0").Err
		if err != nil {
			fmt.Fprintf(rs.Log, "Error: unable to update count: %s\n", err)
			return code, err
		}
	}
	return code, nil
}

// Update an existing key
func (rs *RedisStore) Update(urlStr, code string) (ID string, err error) {
	// TODO FIXME -- base 36 encode of ID?
	err = rs.redisConn.Cmd("SET", rs.RedisPrefix+":"+code, urlStr).Err
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: unable to write file: %s\n", err)
		return code, err
	}
	return code, nil
}

// Exists returns true if the ID exists in the database
func (rs *RedisStore) Exists(ID string) (rv bool) {
	tmp, err := rs.redisConn.Cmd("GET", rs.RedisPrefix+":"+ID).Str()
	if err != nil || tmp == "" {
		rv = false
	} else {
		rv = true
	}
	if db5 {
		fmt.Printf("%sExists(%s) = %v%s\n", MiscLib.ColorCyan, ID, rv, MiscLib.ColorReset)
	}
	return
}

// Fetch converts from a `code` into a `url` to be returned.
func (rs *RedisStore) Fetch(code string) (string, error) {
	urlBytes, err := rs.redisConn.Cmd("GET", rs.RedisPrefix+":"+code).Str()
	if rs.CountHits {
		_, err := rs.redisConn.Cmd("INCR", rs.RedisPrefix+"^"+code).Int()
		if err != nil {
			fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
		}
	}
	return string(urlBytes), err
}

// ConnectToRedis create the connection to the Redis in memory store and returns it.
func ConnectToRedis(redisHost, redisPort, redisAuth string) (redisConn *redis.Client, err error) {
	redisConn, err = redis.Dial("tcp", redisHost+":"+redisPort)
	if err != nil {
		fmt.Printf("qr-short: Failed to connect to redis-server.\n")
		return
	}

	if redisAuth != "" { // New Redis AUTH section
		t, err := redisConn.Cmd("AUTH", redisAuth).Str()
		if err != nil {
			fmt.Printf("%sqr-short: Failed to authorize to use redis. error:%s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			return nil, err
		}
		fmt.Printf("%sqr-short: Connected and Authorized to redis-server. %s%s\n", MiscLib.ColorGreen, t, MiscLib.ColorReset)
	} else {
		fmt.Printf("%sqr-short: Connected to redis-server.%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
	}
	return
}

// List returns a list of all the redirects between beg and end, where beg
// can be 0 to start at the beginning and end can be 'last' | '*' | 'latest'
// to to go the most recent item.
// The maximum number reutrned is 1000 at a time.  If you want more than
// 1000 returend then you will need to call multiple times.
func (rs *RedisStore) List(beg, end string) (dat []ListData, err error) {
	var maxID int
	var begI64, endI64 int64
	maxID, err = rs.redisConn.Cmd("GET", rs.RedisPrefix+"!seq").Int()
	if err != nil {
		fmt.Fprintf(rs.Log, "Fatal Error: %s, %s\n", err, godebug.LF())
		os.Exit(2)
		return
	}

	var begInt int
	begI64, err = strconv.ParseInt(beg, 10, 64)
	if err != nil {
		fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
		return
	}
	begInt = int(begI64)

	var endInt int
	// fmt.Printf("AT: %s\n", godebug.LF())
	if end == "last" || end == "*" || end == "latest" {
		// fmt.Printf("AT: %s\n", godebug.LF())
		endInt = maxID
	} else {
		// fmt.Printf("AT: %s\n", godebug.LF())
		endI64, err = strconv.ParseInt(end, 10, 64)
		if err != nil {
			// fmt.Printf("AT: %s\n", godebug.LF())
			fmt.Fprintf(rs.Log, "Error: %s, %s\n", err, godebug.LF())
			return
		}
		// fmt.Printf("AT: %s\n", godebug.LF())
		endInt = int(endI64)
	}
	// fmt.Printf("AT: %s\n", godebug.LF())
	err = nil

	var nUse int
	var dbURL string
	// limit # returned to 1000 at a time.
	dataRange := endInt - begInt + 1
	if dataRange <= 0 {
		fmt.Fprintf(rs.Log, "Error: dataRange is invalid, %d, %s\n", dataRange, godebug.LF())
		err = fmt.Errorf("Invalid range for data, %d from end(%d)-beg(%d)+1", dataRange, endInt, begInt)
		return
	}
	if dataRange > 1000 {
		dataRange = 1000
	}
	// fmt.Printf("AT: %s\n", godebug.LF())
	dat = make([]ListData, 0, dataRange)
	jj := 0
	for ii := begInt; ii < endInt && jj < 1000; ii++ {
		jj++

		if db3 {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		key := strconv.FormatUint(uint64(ii), 36) // Base 36, take count of # of files add 1, this is the code.
		if db4 {
			fmt.Printf("GET %s\n", fmt.Sprintf("%s:%s", rs.RedisPrefix, key))
		}
		dbURL, err = rs.redisConn.Cmd("GET", fmt.Sprintf("%s:%s", rs.RedisPrefix, key)).Str()
		if err != nil || dbURL == "" {
			if db4 {
				fmt.Printf("err: %s AT: %s\n", err, godebug.LF())
			}
			continue
		} else {
			if db4 {
				fmt.Printf("GET %s -- success\n", fmt.Sprintf("%s:%s", rs.RedisPrefix, key))
			}
		}

		if db3 {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		nUse = 0
		if rs.CountHits {
			key := strconv.FormatUint(uint64(ii), 36) // Base 36, take count of # of files add 1, this is the code.
			nUse, err = rs.redisConn.Cmd("GET", fmt.Sprintf("%s^%s", rs.RedisPrefix, key)).Int()
			if err != nil {
				// fmt.Printf("AT: %s\n", godebug.LF())
				fmt.Fprintf(rs.Log, "Ignored Error: %s, %s\n", err, godebug.LF())
				nUse, err = 0, nil
			}
		}

		if db4 {
			fmt.Printf("%sAppend data to list to return - AT: %s%s\n", MiscLib.ColorGreen, godebug.LF(), MiscLib.ColorReset)
		}
		dat = append(dat, ListData{
			ID:    fmt.Sprintf("%d", ii),
			URL:   dbURL,
			Count: nUse,
		})
	}

	err = nil
	if db4 {
		fmt.Printf("dat=%s AT: %s\n", godebug.SVarI(dat), godebug.LF())
	}
	return

}

// SetDebug turns on debug flags in this module.
func SetDebug(db map[string]bool) {
	if db["storage.db3"] {
		db3 = true
	}
	if db["storage.db4"] {
		db4 = true
	}
	if db["storage.db5"] {
		db5 = true
	}
	if db["storage.db6"] {
		db6 = true
	}
}

// UpdateInsert performs an update for existing IDs and an insert on new ids.
func (rs *RedisStore) UpdateInsert(URL string, ID string) (ur UpdateRespItem) {
	var err error
	var code string

	if db5 {
		fmt.Printf("%sUpdateInsert(%s,%s)%s\n", MiscLib.ColorCyan, URL, ID, MiscLib.ColorReset)
	}

	// get current max ID in decimal.
	cID := rs.getID()

	// convert from ID - in B36 to decimal, update NextID = max(cur,1+id) if necessary.
	IDint, err := strconv.ParseInt(ID, 36, 64) // Base 36, Parse the int into a number
	if err != nil {
		ur.ID = ID
		ur.Msg = fmt.Sprintf("fail:%s", err)
		return
	}
	if IDint >= cID {
		if db6 {
			fmt.Printf("IDint=%d AT: %s\n", IDint, godebug.LF())
		}
		rs.setID(IDint + 1)
	}

	if !rs.Exists(ID) {
		code, err = rs.insertInternal(URL, ID)
		ur.Msg = "success/insert"
	} else {
		code, err = rs.Update(URL, ID)
		ur.Msg = "success/update"
	}
	ur.ID = code

	if err != nil {
		ur.Msg = fmt.Sprintf("fail:%s", err)
	}

	return
}

var db3 = false
var db4 = false
var db5 = false
var db6 = false
