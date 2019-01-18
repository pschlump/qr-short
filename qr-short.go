package main

// Copyright (C) Philip Schlump 2018-2019.

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/American-Certified-Brands/tools/qr-short/storage"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib"
	"github.com/pschlump/getHomeDir"
	"github.com/pschlump/godebug"
)

// ConfigType is the global configuration that is read in from cfg.json
type ConfigType struct {
	DataDir       string   // ~/data
	Port          string   // 2004
	StorageSystem string   // --store file (default), --store Redis
	RedisHost     string   // if Redis, then default 127.0.0.1
	RedisPort     string   // if Redis, then default 6379
	RedisAuth     string   // If auth is not used then leave empty.
	RedisPrefix   string   // default "qr"
	AuthToken     string   // authorize update/set of redirects
	CountHits     bool     // Count number of times referenced
	DebugFlags    []string // List of debug flags - set by default
	DataFileDest  string   // Where to store data when it is passed
}

var gCfg ConfigType
var gLog *os.File
var gDebug map[string]bool

func init() {
	gDebug = make(map[string]bool)
	gCfg = ConfigType{
		DataDir:       "~/data",
		Port:          "2004",
		StorageSystem: "file",
		RedisHost:     "127.0.0.1",
		RedisPort:     "6379",
		RedisPrefix:   "qr",
		AuthToken:     "ENV$QR_SHORT_AUTH_TOKEN",
		CountHits:     false,
		DebugFlags:    make([]string, 0, 10),
		DataFileDest:  "/Users/corwin/go/src/github.com/American-Certified-Brands/tools/qr-short/test-data",
	}
	// gDebug["db002"] = true
	gLog = os.Stderr
}

func main() {
	var Note = flag.String("note", "", "User note")
	_ = Note

	var Cfg = flag.String("cfg", "cfg.json", "config file, default ./cfg.json")
	var Port = flag.String("port", "", "set port to listen on")
	var DataDir = flag.String("datadir", "", "set directory to put files in if storage is 'file'")
	var MaxCPU = flag.Bool("maxcpu", false, "set max number of CPUs.")
	var Store = flag.String("store", "", "which storage system to use, file, Redis.")
	var Debug = flag.String("debug", "", "comma list of flags")
	var AuthToken = flag.String("authtoken", "", "auth token for update/set")

	var err error

	flag.Parse()
	fns := flag.Args()
	if len(fns) > 0 {
		fmt.Printf("Usage: qr-short [--cfg fn] [--port ####] [--datadir path] [--maxcpu] [--store file|Redis] [--debug flag,flag...] [--authtoken token]\n")
		os.Exit(1)
	}

	if Cfg != nil && *Cfg != "" {
		gCfg, err = ReadConfig(*Cfg, gCfg)
		if err != nil {
			os.Exit(1)
		}
	}

	for _, dd := range gCfg.DebugFlags {
		gDebug[dd] = true
	}
	SetDebugFlags(Debug)
	storage.SetDebug(gDebug)

	if *Port != "" {
		gCfg.Port = *Port
	}
	if *DataDir != "" {
		gCfg.DataDir = *DataDir
	}
	if *AuthToken != "" {
		gCfg.AuthToken = *AuthToken
	}

	if *Store != "" && *Store == "Redis" {
		gCfg.StorageSystem = *Store
	} else if *Store != "" && *Store == "file" {
		gCfg.StorageSystem = *Store
	} else if *Store != "" {
		fmt.Fprintf(os.Stderr, "Invalid --store, must be 'file' or 'Redis', supplied [%s]\n", *Store)
		os.Exit(1)
	}

	if *MaxCPU {
		fmt.Printf("%sSetting to use ALL CPUs%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	var data storage.PersistentData

	if gCfg.StorageSystem == "file" {
		data, err = storage.NewFilesystem(getHomeDir.MustExpand(gCfg.DataDir), gLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal: Unable to initialize file system storage: %s\n", err)
			os.Exit(1)
		}
		if gCfg.CountHits {
			fmt.Fprintf(os.Stderr, "Warning: Unable to count hits with file system storage -- not implemented.\n")
		}
	} else if gCfg.StorageSystem == "Redis" {
		data, err = storage.NewRedisStore(gCfg.RedisHost, gCfg.RedisPort, gCfg.RedisAuth, gCfg.RedisPrefix, gCfg.CountHits, gLog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal: Unable to initialize Redis storage: %s\n", err)
			os.Exit(1)
		}
	}

	http.Handle("/enc/", HdlrEncode(data))                 // http.../url=ToUrl
	http.Handle("/enc", HdlrEncode(data))                  // http.../url=ToUrl
	http.Handle("/upd/", HdlrUpdate(data))                 // http.../url=ToUrl&id=Number
	http.Handle("/upd", HdlrUpdate(data))                  // http.../url=ToUrl&id=Number
	http.Handle("/dec/", HdlrDecode(data))                 // http.../id=Number
	http.Handle("/dec", HdlrDecode(data))                  // http.../id=Number
	http.Handle("/list/", HdlrList(data))                  //
	http.Handle("/list", HdlrList(data))                   //
	http.Handle("/bulkLoad", HdlrBulkLoad(data))           //
	http.Handle("/status", http.HandlerFunc(HandleStatus)) //
	http.Handle("/q/", HdlrRedirect(data))                 //

	if db11 {
		fmt.Printf("just before ListenAndServe gCfg=%s\n", godebug.SVarI(gCfg))
	}

	// FIXME - add server name/ip to listen to.
	err = http.ListenAndServe(":"+gCfg.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

var nReq = 0

// HandleStatus - server to respond with a working message if up.
func HandleStatus(www http.ResponseWriter, req *http.Request) {
	nReq++
	fmt.Fprintf(os.Stdout, "\n%sStatus: working.  Requests Served: %d.%s\n", MiscLib.ColorGreen, nReq, MiscLib.ColorReset)
	fmt.Fprintf(www, "Working.  %d requests.\n", nReq)
	return
}

// HdlrEncode returns a closure that handles /enc/ path.
func HdlrEncode(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		nReq++
		if db1 {
			fmt.Printf("Encode: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		if !CheckAuthToken(data, www, req) {
			www.WriteHeader(http.StatusUnauthorized) // 401
			fmt.Fprintf(www, "Error: not authorized.\n")
			return
		}
		found, urlStr := getVar("url", req)
		dataFound, dataStr := getVar("data", req)
		if found {
			enc, err := data.Insert(urlStr)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
				fmt.Fprintf(gLog, "Encode: list error %s, %s\n", err, godebug.LF())
				fmt.Fprintf(www, "Error: encode error: %s\n", err)
				os.Exit(1)
				return
			}
			if dataFound {
				fn := fmt.Sprintf("%s/%s", gCfg.DataFileDest, enc)
				ioutil.WriteFile(fn, []byte(dataStr+"\n"), 0644)
				fmt.Fprintf(os.Stderr, "Data Written To: %s = %s\n", fn, dataStr) // PJS test
				fmt.Fprintf(gLog, "Data Written To: %s = %s\n", fn, dataStr)
			}
			fmt.Fprintf(www, "%s", enc)
			fmt.Fprintf(gLog, "Encode: %s = %s\n", urlStr, enc)
			return
		}
		www.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(www, "Error: expected POST or GET with `url` parameter\n")
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrUpdate returns a closure that handles /upd/ path.
// This will update the specified URL.
func HdlrUpdate(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		nReq++
		if db1 {
			fmt.Printf("Update: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		if !CheckAuthToken(data, www, req) {
			www.WriteHeader(http.StatusUnauthorized) // 401
			fmt.Fprintf(www, "Error: not authorized.\n")
			return
		}

		foundUrl, urlStr := getVar("url", req)
		foundId, id := getVar("id", req)
		dataFound, dataStr := getVar("data", req)
		if foundUrl && foundId {
			enc, err := data.Update(urlStr, id)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
				fmt.Fprintf(gLog, "Update: list error %s, %s\n", err, godebug.LF())
				fmt.Fprintf(www, "Error: update error: %s\n", err)
				os.Exit(1)
				return
			}
			if dataFound {
				fn := fmt.Sprintf("%s/%s", gCfg.DataFileDest, enc)
				ioutil.WriteFile(fn, []byte(dataStr+"\n"), 0644)
				fmt.Fprintf(os.Stderr, "Data Written To: %s = %s\n", fn, dataStr) // PJS test
				fmt.Fprintf(gLog, "Data Written To: %s = %s\n", fn, dataStr)
			}
			fmt.Fprintf(www, "%s", enc)
			fmt.Fprintf(gLog, "Update Encode: %s = %s\n", urlStr, enc)
			return
		}
		www.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(www, "Error: expected POST or GET with `url` parameter\n")
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrDecode takes an ID and decoes it back to a URL.
func HdlrDecode(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		nReq++
		if db1 {
			fmt.Printf("Decode: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		var id string
		if strings.HasPrefix(req.URL.Path, "/dec/") {
			id = req.URL.Path[len("/dec/"):]
		} else {
			id = req.URL.Query().Get("url")
		}
		if id == "" {
			www.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(www, "URL Not Found.\n")
			return
		}

		urlStr, err := data.Fetch(id)
		if err != nil {
			www.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(www, "URL Not Found.  Error: %s\n", err)
			return
		}

		www.Write([]byte(urlStr))
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrRedirect is the real worker in this.  It takes a shortened URL
// with an ID and redirects it to its destination.
func HdlrRedirect(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		nReq++
		if db1 {
			fmt.Printf("Redirect: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		id := req.URL.Path[len("/q/"):]
		qry := req.URL.RawQuery

		fmt.Printf("id: [%s]\n", id)

		url, err := data.Fetch(id)
		if err != nil {
			www.WriteHeader(http.StatusNotFound)
			www.Write([]byte("URL Not Found. Error: " + err.Error() + "\n"))
			return
		}

		fmt.Printf("%sRedirect occurring from %s to %s%s\n", MiscLib.ColorCyan, id, url, MiscLib.ColorReset)

		// xyzzy1008 - What about passing along headers?
		// xyzzy1004 - config for 307 or 301 redirect.
		// h := www.Header()
		req.Header.Set("X-QR-Short", "Redirected By")
		req.Header.Set("X-QR-Short-OrigURL", req.RequestURI)
		// xyzzy1004 - TODO - add template generate redirect page with link if browser has redirect turned off.
		// xyzzy1004 - TODO - add JS id in template to do redirect if browser loads id and runs Ecma Script.
		http.Redirect(www, req, string(url)+"?"+qry, http.StatusTemporaryRedirect) // 307
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrList returns a closure that handles /list/ path.
func HdlrList(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		if db1 {
			fmt.Printf("List: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		if gDebug["db002"] {
			fmt.Fprintf(gLog, "In List at top: %s, %s\n", req.URL.Query(), godebug.LF())
		}
		if !CheckAuthToken(data, www, req) {
			www.WriteHeader(http.StatusUnauthorized) // 401
			fmt.Fprintf(gLog, "List: not authorized, %s\n", godebug.LF())
			fmt.Fprintf(www, "Error: not authorized.\n")
			return
		}
		if gDebug["db002"] {
			fmt.Fprintf(gLog, "GET params are: %s, %s\n", req.URL.Query(), godebug.LF())
		}
		if begStr := req.URL.Query().Get("beg"); begStr != "" {
			if endStr := req.URL.Query().Get("end"); endStr != "" {
				if gDebug["db002"] {
					fmt.Fprintf(gLog, "have big/end: %s, %s\n", req.URL.Query(), godebug.LF())
				}
				data, err := data.List(begStr, endStr)
				if err != nil {
					www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
					fmt.Fprintf(gLog, "List: list error %s, %s\n", err, godebug.LF())
					fmt.Fprintf(www, "Error: list error: %s\n", err)
					return
				}
				json := godebug.SVarI(data)

				// h := www.Header() // set type for return of JSON data
				www.Header().Set("Content-Type", "application/json")

				fmt.Fprintf(www, "%s", json)
				if gDebug["db003"] {
					fmt.Fprintf(gLog, "List: %s ... %s, data = %s\n", begStr, endStr, json)
				}
				return
			}
		}
		www.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(www, "Error: list error\n")
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrBulkLoad returns a closure that handles /enc/ path.
func HdlrBulkLoad(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		nReq++
		if db1 {
			fmt.Printf("BulkLoad: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		if !CheckAuthToken(data, www, req) {
			www.WriteHeader(http.StatusUnauthorized) // 401
			fmt.Fprintf(gLog, "BulkLoad: not authorized, %s\n", godebug.LF())
			fmt.Fprintf(www, "Error: not authorized.\n")
			return
		}

		type UpdateData struct {
			Data []struct {
				URL string `json:"url"`
				ID  string `json:"Id"`
			}
		}
		var respSet []storage.UpdateRespItem

		foundUpdate, updateStr := getVar("update", req)
		if foundUpdate {
			var update UpdateData
			if db2 {
				fmt.Printf("->%s<-, %s\n", updateStr, godebug.LF())
			}
			err := json.Unmarshal([]byte(updateStr), &update)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // 500 is this the correct error to return at this point?
				fmt.Fprintf(gLog, "BulkLoad: parse error: %s, %s, %s\n", update, err, godebug.LF())
				fmt.Fprintf(www, "Error: parse error: %s\n", err)
				return
			}
			for ii, dat := range update.Data {
				resp := data.UpdateInsert(dat.URL, dat.ID)
				resp.Pos = ii
				respSet = append(respSet, resp)
			}

			resp := godebug.SVarI(respSet)

			www.Header().Set("Content-Type", "application/json")
			www.Header().Set("Content-Length", fmt.Sprintf("%d", len(resp)))
			www.Header().Set("Length", fmt.Sprintf("%d", len(resp)))

			fmt.Fprintf(www, "%s", resp)
			fmt.Fprintf(gLog, "Bulk Load: %s = %s\n", updateStr, godebug.SVarI(respSet))
			return
		}
		www.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(www, "Error: list error\n")
	}
	return http.HandlerFunc(handleFunc)
}

// CheckAuthToken looks at either a header or a cookie to determine if the user is
// authorized.
func CheckAuthToken(data storage.PersistentData, www http.ResponseWriter, req *http.Request) bool {
	if gCfg.AuthToken == "-none-" {
		if gDebug["db-auth"] {
			fmt.Fprintf(gLog, "%sAuth Success - no authentication%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		return true
	}

	// look for cookie
	cookie, err := req.Cookie("Qr-Auth")
	if gDebug["db-auth"] {
		fmt.Printf("Cookie: %s\n", godebug.SVarI(cookie))
	}
	if err == nil {
		if cookie.Value == gCfg.AuthToken {
			if gDebug["db-auth"] {
				fmt.Fprintf(gLog, "%sAuth Success - cookie%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
			}
			return true
		}
	}

	// look for header
	// ua := r.Header.Get("User-Agent")
	auth := req.Header.Get("X-Qr-Auth")
	if gDebug["db-auth"] {
		fmt.Printf("Header: %s\n", godebug.SVarI(auth))
	}
	if auth == gCfg.AuthToken {
		if gDebug["db-auth"] {
			fmt.Fprintf(gLog, "%sAuth Success - header%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		return true
	}

	if gDebug["db-auth"] {
		fmt.Fprintf(gLog, "%sAuth Fail%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
	}
	return false
}

// ReadConfig reads in the configuration file and substitutes environment
// variables for passwords/auth-tokens.
func ReadConfig(fn string, in ConfigType) (rv ConfigType, err error) {
	rv = in
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: Unable to open %s for configuration, error=%s\n", fn, err)
		os.Exit(1)
	}
	err = json.Unmarshal(buf, &rv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: Unable to parse %s for configuration, error=%s\n", fn, err)
		es := jsonSyntaxErroLib.GenerateSyntaxError(string(buf), err)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", MiscLib.ColorYellow, es, MiscLib.ColorReset)
		os.Exit(1)
	}

	if strings.HasPrefix(rv.RedisAuth, "ENV$") {
		name := rv.RedisAuth[4:]
		val := os.Getenv(name)
		rv.RedisAuth = val
	}

	// AuthToken:     "ENV$QR_SHORT_AUTH_TOKEN"
	if strings.HasPrefix(rv.AuthToken, "ENV$") {
		name := rv.AuthToken[4:]
		val := os.Getenv(name)
		rv.AuthToken = val
	}

	if db11 {
		fmt.Printf("rv=%s, %s\n", godebug.SVarI(rv), godebug.LF())
	}

	return rv, nil
}

// SetDebugFlags convers from --debug csv,csv -> gDebug
func SetDebugFlags(Debug *string) {
	if Debug != nil && *Debug != "" {
		df := strings.Split(*Debug, ",")
		for _, dd := range df {
			if _, have := gDebug[dd]; have {
				gDebug[dd] = !gDebug[dd]
			} else {
				gDebug[dd] = true
			}
		}
	}
	if gDebug["db1"] {
		db1 = true
	}
	if gDebug["db2"] {
		db2 = true
	}
	if gDebug["db11"] {
		db11 = true
	}
	if gDebug["db12"] {
		db12 = true
	}
}

func getVar(name string, req *http.Request) (found bool, value string) {
	method := req.Method
	if method == "POST" {
		if str := req.PostFormValue(name); str != "" {
			value = str
			found = true
		}
	} else if method == "GET" {
		if str := req.URL.Query().Get(name); str != "" {
			value = str
			found = true
		}
	}
	if gDebug["db008"] {
		fmt.Fprintf(gLog, "Method %s Param %s Value %s: %s\n", method, name, value, godebug.LF())
	}
	return
}

var db1 = false
var db2 = false
var db11 = false
var db12 = false
