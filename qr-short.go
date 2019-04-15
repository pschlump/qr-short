package main

// Copyright (C) Philip Schlump 2018-2019.

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/American-Certified-Brands/config-sample/ReadConfig"
	"github.com/American-Certified-Brands/tools/GetVar"
	"github.com/American-Certified-Brands/tools/qr-short/storage"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/getHomeDir"
	"github.com/pschlump/godebug"
	MonAliveLib "github.com/pschlump/mon-alive/lib" // "github.com/pschlump/mon-alive/lib"
	"github.com/pschlump/radix.v2/redis"
)

// xyzzy2000 Wed Mar 20 16:52:43 MDT 2019 -- PJS -- count number of redirects
// xyzzy2001 User login - Keep "user_id" assoc with data.
// 		auth_token -> user_id on data changes.
//		qa:auth_token -> user_id
// xyzzy2003 Drop file storage

// ConfigType is the global configuration that is read in from cfg.json
type ConfigType struct {
	DataDir          string `default:"~/data"` // ~/data
	HostPort         string `default:":2004"`  // 2004
	StorageSystem    string `default:"Redis"`  // --store file (default), --store Redis
	RedisConnectHost string `json:"redis_host" default:"$ENV$REDIS_HOST"`
	RedisConnectAuth string `json:"redis_auth" default:"$ENV$REDIS_AUTH"`
	RedisConnectPort string `json:"redis_port" default:"6379"`
	RedisPrefix      string `default:"qr"`                       // default "qr"
	AuthToken        string `default:"$ENV$QR_SHORT_AUTH_TOKEN"` // authorize update/set of redirects
	CountHits        bool   `default:"false"`                    // Count number of times referenced
	DataFileDest     string `default:"./test-data"`              // Where to store data when it is passed
	LogFileName      string `json:"log_file_name"`
	DebugFlag        string `json:"db_flag"`

	// Default file for TLS setup (Should include path), both must be specified.
	// These can be over ridden on the command line.
	TLS_crt string `json:"tls_crt" default:""`
	TLS_key string `json:"tls_key" default:""`
}

var gCfg ConfigType
var logFilePtr *os.File
var db_flag map[string]bool
var isTLS bool
var wg sync.WaitGroup
var httpServer *http.Server
var shutdownWaitTime = time.Duration(1)
var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", 0)
	isTLS = false
	db_flag = make(map[string]bool)
	db_flag["db002"] = true
	db_flag["db-auth"] = true
	logFilePtr = os.Stderr
}

var Note = flag.String("note", "", "User note")
var Cfg = flag.String("cfg", "cfg.json", "config file, default ./cfg.json")
var Cli = flag.String("cli", "", "Run as a CLI command intead of a server")
var DataDir = flag.String("datadir", "", "set directory to put files in if storage is 'file'")
var MaxCPU = flag.Bool("maxcpu", false, "set max number of CPUs.")
var Store = flag.String("store", "", "which storage system to use, file, Redis.")
var AuthToken = flag.String("authtoken", "", "auth token for update/set")
var TLS_crt = flag.String("tls_crt", "", "TLS Signed Publick Key")
var TLS_key = flag.String("tls_key", "", "TLS Signed Private Key")
var DbFlag = flag.String("db_flag", "", "Additional Debug Flags")
var HostPort = flag.String("hostport", ":2004", "Host/Port to listen on")

func main() {

	var err error

	flag.Parse() // Parse CLI arguments to this, --cfg <name>.json

	fns := flag.Args()
	if *Cli != "" {
		GetVar.SetCliOpts(Cli, fns)
	} else if len(fns) != 0 {
		fmt.Printf("Usage: qr-short [--cfg fn] [--port ####] [--datadir path] [--maxcpu] [--store file|Redis] [--debug flag,flag...] [--authtoken token]\n")
		os.Exit(1)
	}

	if Cfg == nil {
		fmt.Printf("--cfg is a required parameter\n")
		os.Exit(1)
	}

	// ------------------------------------------------------------------------------
	// Read in Configuraiton
	// ------------------------------------------------------------------------------
	err = ReadConfig.ReadFile(*Cfg, &gCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read confguration: %s error %s\n", *Cfg, err)
		os.Exit(1)
	}

	// ------------------------------------------------------------------------------
	// Logging File
	// ------------------------------------------------------------------------------
	if gCfg.LogFileName != "" {
		bn := path.Dir(gCfg.LogFileName)
		os.MkdirAll(bn, 0755)
		fp, err := filelib.Fopen(gCfg.LogFileName, "a")
		if err != nil {
			log.Fatalf("log file confiured, but unable to open, file[%s] error[%s]\n", gCfg.LogFileName, err)
		}
		LogFile(fp)
	}

	// ------------------------------------------------------------------------------
	// TLS parameter check / setup
	// ------------------------------------------------------------------------------
	if *TLS_crt == "" && gCfg.TLS_crt != "" {
		TLS_crt = &gCfg.TLS_crt
	}
	if *TLS_key == "" && gCfg.TLS_key != "" {
		TLS_key = &gCfg.TLS_key
	}

	if *TLS_crt != "" && *TLS_key == "" {
		log.Fatalf("Must supply both .crt and .key for TLS to be turned on - fatal error.")
	} else if *TLS_crt == "" && *TLS_key != "" {
		log.Fatalf("Must supply both .crt and .key for TLS to be turned on - fatal error.")
	} else if *TLS_crt != "" && *TLS_key != "" {
		if !filelib.Exists(*TLS_crt) {
			log.Fatalf("Missing file ->%s<-\n", *TLS_crt)
		}
		if !filelib.Exists(*TLS_key) {
			log.Fatalf("Missing file ->%s<-\n", *TLS_key)
		}
		isTLS = true
	}

	// ------------------------------------------------------------------------------
	// Debug Flag Processing
	// ------------------------------------------------------------------------------
	if gCfg.DebugFlag != "" {
		ss := strings.Split(gCfg.DebugFlag, ",")
		// fmt.Printf("gCfg.DebugFlag ->%s<-\n", gCfg.DebugFlag)
		for _, sx := range ss {
			// fmt.Printf("Setting ->%s<-\n", sx)
			db_flag[sx] = true
		}
	}
	if *DbFlag != "" {
		ss := strings.Split(*DbFlag, ",")
		// fmt.Printf("gCfg.DebugFlag ->%s<-\n", gCfg.DebugFlag)
		for _, sx := range ss {
			// fmt.Printf("Setting ->%s<-\n", sx)
			db_flag[sx] = true
		}
	}
	if db_flag["dump-db-flag"] {
		fmt.Fprintf(os.Stderr, "%sDB Flags Enabled Are:%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		for x := range db_flag {
			fmt.Fprintf(os.Stderr, "%s\t%s%s\n", MiscLib.ColorGreen, x, MiscLib.ColorReset)
		}
	}
	GetVar.SetDbFlag(db_flag)

	// ---- Copy to new verion ---- // ---- Copy to new verion ---- // ---- Copy to new verion ---- // ---- Copy to new verion ----
	for _, dd := range strings.Split(gCfg.DebugFlag, ",") {
		db_flag[dd] = true
	}
	SetDebugFlags()
	storage.SetDebug(db_flag)

	if *HostPort != "" {
		gCfg.HostPort = *HostPort
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
		data, err = storage.NewFilesystem(getHomeDir.MustExpand(gCfg.DataDir), logFilePtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal: Unable to initialize file system storage: %s\n", err)
			os.Exit(1)
		}
		if gCfg.CountHits {
			fmt.Fprintf(os.Stderr, "Warning: Unable to count hits with file system storage -- not implemented.\n")
		}
	} else if gCfg.StorageSystem == "Redis" {
		data, err = storage.NewRedisStore(gCfg.RedisConnectHost, gCfg.RedisConnectPort, gCfg.RedisConnectAuth, gCfg.RedisPrefix, gCfg.CountHits, logFilePtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal: Unable to initialize Redis storage: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Internal error >%s< should be 'file' or 'Redis'\n", gCfg.StorageSystem)
		os.Exit(1)
	}

	// xyzzy - AUTH /getAuth/?un=UU&pw=YY -> Auth Token / Cookie

	mux := http.NewServeMux()
	mux.Handle("/api/v1/status", http.HandlerFunc(HandleStatus))          //
	mux.Handle("/api/v1/exit-server", http.HandlerFunc(HandleExitServer)) //

	mux.Handle("/enc/", HdlrEncode(data))                 // http.../url=ToUrl					Auth Req
	mux.Handle("/enc", HdlrEncode(data))                  // http.../url=ToUrl					Auth Req
	mux.Handle("/upd/", HdlrUpdate(data))                 // http.../url=ToUrl&id=Number		Auth Req
	mux.Handle("/upd", HdlrUpdate(data))                  // http.../url=ToUrl&id=Number		Auth Req
	mux.Handle("/dec/", HdlrDecode(data))                 // http.../id=Number
	mux.Handle("/dec", HdlrDecode(data))                  // http.../id=Number
	mux.Handle("/list/", HdlrList(data))                  // http...?beg=NUmber&end=Number		Auth Req.
	mux.Handle("/list", HdlrList(data))                   // http...?beg=NUmber&end=Number		Auth Req.
	mux.Handle("/bulkLoad", HdlrBulkLoad(data))           //
	mux.Handle("/status", http.HandlerFunc(HandleStatus)) //
	mux.Handle("/q/", HdlrRedirect(data))                 //
	mux.Handle("/", http.FileServer(http.Dir("www")))

	// ------------------------------------------------------------------------------
	// Setup signal capture
	// ------------------------------------------------------------------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// ------------------------------------------------------------------------------
	// Live Monitor Setup
	// ------------------------------------------------------------------------------
	monClient, err7 := RedisClient()
	fmt.Printf("err7=%v AT: %s\n", err7, godebug.LF())
	mon := MonAliveLib.NewMonIt(func() *redis.Client { return monClient }, func(conn *redis.Client) {})
	mon.SendPeriodicIAmAlive("QR-Short-MS")

	// ------------------------------------------------------------------------------
	// Setup / Run the HTTP Server.
	// ------------------------------------------------------------------------------
	if isTLS {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}
		httpServer = &http.Server{
			Addr:         *HostPort,
			Handler:      mux,
			TLSConfig:    cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
	} else {
		httpServer = &http.Server{
			Addr:    *HostPort,
			Handler: mux,
		}
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		if isTLS {
			fmt.Fprintf(os.Stderr, "%sListening on https://%s%s\n", MiscLib.ColorGreen, *HostPort, MiscLib.ColorReset)
			if err := httpServer.ListenAndServeTLS(*TLS_crt, *TLS_key); err != nil {
				logger.Fatal(err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "%sListening on http://%s%s\n", MiscLib.ColorGreen, *HostPort, MiscLib.ColorReset)
			if err := httpServer.ListenAndServe(); err != nil {
				logger.Fatal(err)
			}
		}
	}()

	// ------------------------------------------------------------------------------
	// Catch signals from [Contro-C]
	// ------------------------------------------------------------------------------
	select {
	case <-stop:
		fmt.Fprintf(os.Stderr, "\nShutting down the server... Received OS Signal...\n")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime*time.Second)
		defer cancel()
		err := httpServer.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Error on shutdown: [%s]\n", err)
		}
	}

	// ------------------------------------------------------------------------------
	// Wait for HTTP server to exit.
	// ------------------------------------------------------------------------------
	wg.Wait()
}

var nReq = 0

// HandleStatus - server to respond with a working message if up.
func HandleStatus(www http.ResponseWriter, req *http.Request) {
	nReq++
	fmt.Fprintf(os.Stdout, "\n%sStatus: working.  Requests Served: %d.%s\n", MiscLib.ColorGreen, nReq, MiscLib.ColorReset)
	www.Header().Set("Content-Type", "text/html; charset=utf-8")
	www.WriteHeader(http.StatusOK) // 401
	// fmt.Fprintf(www, "Working.  %d requests. (Version 0.0.18, Mod Date: Thu Feb  7 06:49:19 MST 2019)\n", nReq)
	fmt.Fprintf(www, "Working.  Version v0.0.19 %d Requests. (Mod Date: Sat Mar 23 08:56:52 MDT 2019)\n", nReq)
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
		found, urlStr := GetVar.GetVar("url", www, req)
		dataFound, dataStr := GetVar.GetVar("data", www, req)
		if found {
			enc, err := data.Insert(urlStr)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
				fmt.Fprintf(logFilePtr, "Encode: list error %s, %s\n", err, godebug.LF())
				fmt.Fprintf(www, "Error: encode error: %s\n", err)
				os.Exit(1)
				return
			}
			if dataFound {
				fn := fmt.Sprintf("%s/%s", gCfg.DataFileDest, enc)
				ioutil.WriteFile(fn, []byte(dataStr+"\n"), 0644)
				fmt.Fprintf(os.Stderr, "Data Written To: %s = %s\n", fn, dataStr) // PJS test
				fmt.Fprintf(logFilePtr, "Data Written To: %s = %s\n", fn, dataStr)
			}
			fmt.Fprintf(www, "%s", enc)
			fmt.Fprintf(logFilePtr, "Encode: %s = %s\n", urlStr, enc)
			return
		}
		www.WriteHeader(http.StatusBadRequest)
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

		foundUrl, urlStr := GetVar.GetVar("url", www, req)
		foundId, id := GetVar.GetVar("id", www, req)
		dataFound, dataStr := GetVar.GetVar("data", www, req)
		if foundUrl && foundId {
			enc, err := data.Update(urlStr, id)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
				fmt.Fprintf(logFilePtr, "Update: list error %s, %s\n", err, godebug.LF())
				fmt.Fprintf(www, "Error: update error: %s\n", err)
				os.Exit(1)
				return
			}
			if dataFound {
				fn := fmt.Sprintf("%s/%s", gCfg.DataFileDest, enc)
				ioutil.WriteFile(fn, []byte(dataStr+"\n"), 0644)
				fmt.Fprintf(os.Stderr, "Data Written To: %s = %s\n", fn, dataStr) // PJS test
				fmt.Fprintf(logFilePtr, "Data Written To: %s = %s\n", fn, dataStr)
			}
			fmt.Fprintf(www, "%s", enc)
			fmt.Fprintf(logFilePtr, "Update Encode: %s = %s\n", urlStr, enc)
			return
		}

		fmt.Printf("%sFailed foundUrl=%v [%s] foundId=%v [%s] foundData=%v [%s]%s\n", MiscLib.ColorRed, foundUrl, urlStr, foundId, id, dataFound, dataStr, MiscLib.ColorReset)

		www.WriteHeader(http.StatusBadRequest)
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
			id = req.URL.Query().Get("id")
		}
		if id == "" {
			www.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(www, "URL Not Found.\n")
			fmt.Fprintf(os.Stderr, "URL Not Found.\n")
			return
		}
		if db1 {
			fmt.Printf("Decode: id=%s, %s\n", id, godebug.LF())
		}

		urlStr, err := data.Fetch(id)
		if err != nil {
			www.WriteHeader(http.StatusExpectationFailed)
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
		fmt.Printf("AT: %s qry ->%s<-\n", godebug.LF(), qry)

		fmt.Printf("id: [%s]\n", id)

		URL, err := data.Fetch(id)
		if err != nil {
			fmt.Printf("%sRedirect occurring from [%s] to [%s] -- failed to find in Redis%s\n", MiscLib.ColorCyan, id, URL, MiscLib.ColorReset)
			www.WriteHeader(http.StatusNotFound)
			www.Write([]byte("URL Not Found. Error: " + err.Error() + "\n"))
			return
		}
		/*
			URL, err = url.QueryUnescape(URL)
			if err != nil {
				fmt.Printf("%sRedirect occurring from [%s] to [%s] -- error:%s%s\n", MiscLib.ColorCyan, id, URL, err, MiscLib.ColorReset)
				www.WriteHeader(500)
				www.Write([]byte("URL Not Found. Error: " + err.Error() + "\n"))
				return
			}
		*/

		fmt.Printf("%sRedirect occurring from [%s] to [%s]%s\n", MiscLib.ColorCyan, id, URL, MiscLib.ColorReset)

		req.Header.Set("X-QR-Short", "Redirected By")
		req.Header.Set("X-QR-Short-OrigURL", req.RequestURI)

		// xyzzy2000 -- PJS -- count number of redirects
		data.IncrementRedirectCount(id)

		// Take care of URLs that arlready have prameters in them.
		uu := string(URL)
		sep := "?"
		if strings.Contains(URL, "?") {
			sep = "&"
		}
		if qry != "" {
			uu += sep + qry
		}

		http.Redirect(www, req, uu, http.StatusTemporaryRedirect) // 307
	}
	return http.HandlerFunc(handleFunc)
}

// HdlrList returns a closure that handles /list/ path.
func HdlrList(data storage.PersistentData) http.Handler {
	handleFunc := func(www http.ResponseWriter, req *http.Request) {
		if db1 {
			fmt.Printf("List: %s, %s\n", godebug.SVarI(req), godebug.LF())
		}
		if db_flag["db002"] {
			fmt.Fprintf(logFilePtr, "In List at top: %s, %s\n", req.URL.Query(), godebug.LF())
		}
		if !CheckAuthToken(data, www, req) {
			www.WriteHeader(http.StatusUnauthorized) // 401
			fmt.Fprintf(logFilePtr, "List: not authorized, %s\n", godebug.LF())
			fmt.Fprintf(www, "Error: not authorized.\n")
			return
		}
		if db_flag["db002"] {
			fmt.Fprintf(logFilePtr, "GET params are: %s, %s\n", req.URL.Query(), godebug.LF())
		}
		if begStr := req.URL.Query().Get("beg"); begStr != "" {
			if endStr := req.URL.Query().Get("end"); endStr != "" {
				if db_flag["db002"] {
					fmt.Fprintf(logFilePtr, "have big/end: %s, %s\n", req.URL.Query(), godebug.LF())
				}
				data, err := data.List(begStr, endStr)
				if err != nil {
					www.WriteHeader(http.StatusInternalServerError) // is this the correct error to return at this point?
					fmt.Fprintf(logFilePtr, "List: list error %s, %s\n", err, godebug.LF())
					fmt.Fprintf(www, "Error: list error: %s\n", err)
					return
				}
				json := godebug.SVarI(data)

				// h := www.Header() // set type for return of JSON data
				www.Header().Set("Content-Type", "application/json")

				fmt.Fprintf(www, "%s", json)
				if db_flag["db003"] {
					fmt.Fprintf(logFilePtr, "List: %s ... %s, data = %s\n", begStr, endStr, json)
				}
				return
			}
		}
		www.WriteHeader(http.StatusBadRequest)
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
			fmt.Fprintf(logFilePtr, "BulkLoad: not authorized, %s\n", godebug.LF())
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

		foundUpdate, updateStr := GetVar.GetVar("update", www, req)
		if foundUpdate {
			var update UpdateData
			if db2 {
				fmt.Printf("->%s<-, %s\n", updateStr, godebug.LF())
			}
			err := json.Unmarshal([]byte(updateStr), &update)
			if err != nil {
				www.WriteHeader(http.StatusInternalServerError) // 500 is this the correct error to return at this point?
				fmt.Fprintf(logFilePtr, "BulkLoad: parse error: %s, %s, %s\n", update, err, godebug.LF())
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
			fmt.Fprintf(logFilePtr, "Bulk Load: %s = %s\n", updateStr, godebug.SVarI(respSet))
			return
		}
		www.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(www, "Error: list error\n")
	}
	return http.HandlerFunc(handleFunc)
}

// CheckAuthToken looks at either a header or a cookie to determine if the user is
// authorized.
func CheckAuthToken(data storage.PersistentData, www http.ResponseWriter, req *http.Request) bool {
	if db_flag["db-auth"] {
		fmt.Printf("In CheckAuthToken: Looking for [%s]\n", gCfg.AuthToken)
	}
	if gCfg.AuthToken == "-none-" {
		if db_flag["db-auth"] {
			fmt.Fprintf(logFilePtr, "%sAuth Success - no authentication%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		return true
	}

	// look for cookie
	cookie, err := req.Cookie("Qr-Auth")
	if db_flag["db-auth"] {
		fmt.Printf("Cookie: %s\n", godebug.SVarI(cookie))
	}
	if err == nil {
		if cookie.Value == gCfg.AuthToken {
			if db_flag["db-auth"] {
				fmt.Fprintf(logFilePtr, "%sAuth Success - cookie%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
			}
			return true
		}
	}

	// look for header
	// ua := r.Header.Get("User-Agent")
	auth := req.Header.Get("X-Qr-Auth")
	if db_flag["db-auth"] {
		fmt.Printf("Header: %s\n", godebug.SVarI(auth))
	}
	if auth == gCfg.AuthToken {
		if db_flag["db-auth"] {
			fmt.Fprintf(logFilePtr, "%sAuth Success - header%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		return true
	}

	auth_key_found, auth_key := GetVar.GetVar("auth_key", www, req)
	if db_flag["db-auth"] {
		fmt.Printf("Variable: %s\n", auth_key)
	}
	if auth_key_found && auth_key == gCfg.AuthToken {
		if db_flag["db-auth"] {
			fmt.Fprintf(logFilePtr, "%sAuth Success - header%s\n", MiscLib.ColorGreen, MiscLib.ColorReset)
		}
		return true
	}

	if db_flag["db-auth"] {
		fmt.Fprintf(logFilePtr, "%sAuth Fail%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
	}
	return false
}

// SetDebugFlags convers from db_flag values to db? Variables.
func SetDebugFlags() {
	if db_flag["db1"] {
		db1 = true
	}
	if db_flag["db2"] {
		db2 = true
	}
	if db_flag["db11"] {
		db11 = true
	}
	if db_flag["db12"] {
		db12 = true
	}
}

// LogFile sets the output log file to an open file.  This will turn on logging of SQL statments.
func LogFile(f *os.File) {
	logFilePtr = f
}

// HandleExitServer - graceful server shutdown.
func HandleExitServer(www http.ResponseWriter, req *http.Request) {

	// if !IsAuthKeyValid(www, req) {
	if !CheckAuthToken(nil, www, req) {
		return
	}
	if isTLS {
		www.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
	www.Header().Set("Content-Type", "application/json; charset=utf-8")

	//	// fmt.Printf("AT: %s - gCfg.AuthKey = [%s]\n", godebug.LF(), gCfg.AuthKey)
	//	found, auth_key := GetVar.GetVar("auth_key", www, req)
	//	if gCfg.AuthKey != "" {
	//		// fmt.Printf("AT: %s - configed AuthKey [%s], found=%v ?auth_key=[%s]\n", godebug.LF(), gCfg.AuthKey, found, auth_key)
	//		if !found || auth_key != gCfg.AuthKey {
	//			// fmt.Printf("AT: %s\n", godebug.LF())
	//			www.WriteHeader(http.StatusUnauthorized) // 401
	//			return
	//		}
	//	}
	//	// fmt.Printf("AT: %s\n", godebug.LF())

	www.WriteHeader(http.StatusOK) // 200
	fmt.Fprintf(www, `{"status":"success"}`+"\n")

	go func() {
		// Implement graceful exit with auth_key
		fmt.Fprintf(os.Stderr, "\nShutting down the server... Received /exit-server?auth_key=...\n")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime*time.Second)
		defer cancel()
		err := httpServer.Shutdown(ctx)
		if err != nil {
			fmt.Printf("Error on shutdown: [%s]\n", err)
		}
	}()
}

// xyzzy - fix this -- really should be by function debug names.
// SetDebugFlags(db_flag)
var db1 = false
var db2 = false
var db11 = false
var db12 = false
