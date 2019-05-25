package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/American-Certified-Brands/config-sample/ReadConfig"
	"github.com/American-Certified-Brands/tools/get"
	"github.com/pschlump/godebug"
	ms "github.com/pschlump/templatestrings"
)

// ConfigType is the global configuration that is read in from cfg.json
type ConfigType struct {
	HostURLPort   string `json:"HostURLPort" default:"http://127.0.0.1:2004"`      // URL of qr-short server
	AuthToken     string `json:"qr_auth_token" default:"$ENV$QR_SHORT_AUTH_TOKEN"` // Auth key for taling to qr-short
	BaseServerURL string `json:"base_server_url" default:"http://127.0.0.1:9022"`  // QR Image Server (qr-micro-service, 127.0.0.1:9022?)
}

var gCfg ConfigType

var Cfg = flag.String("cfg", "cfg.json", "config file, default ./cfg.json")
var Data = flag.String("data", "data.csv", "Input to bulk update")
var StartID = flag.Int("startId", -1, "Starting ID value for automatic generation")
var EndID = flag.Int("endId", -1, "Ending ID value for automatic generation")
var Zip = flag.String("zip", "", "zipama")
var BaseURL = flag.String("baseurl", "http://www.agroledge.com", "URL to pull QR images from.")
var Server = flag.String("server", "http://127.0.0.1:2004", "Local or remote qr-short server.") // http://t432z.com for remote

func main() {

	flag.Parse()
	fns := flag.Args()
	if len(fns) > 0 {
		fmt.Printf("Usage: bulk-post2 [--cfg cfg.json] --data data.csv [ --rpt output-file ] [ --zip zip-file ] [ --baseurl http://www.example.com ]\n")
		os.Exit(1)
	}

	if Cfg == nil {
		fmt.Printf("--cfg is a required parameter\n")
		os.Exit(1)
	}

	err := ReadConfig.ReadFile(*Cfg, &gCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read confguration: %s error %s\n", *Cfg, err)
		os.Exit(1)
	}

	rawData := ReadData(*Data, *StartID, *EndID)

	fmt.Printf("rawData ->%s<-\n", rawData)

	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded") //
	fmt.Fprintf(os.Stderr, "Call To ->%s/bulkLoad<-\n", *Server)
	status, rv := get.DoPostHeader(fmt.Sprintf("%s/bulkLoad", *Server), []get.HeaderType{
		{Name: "X-Qr-Auth", Value: gCfg.AuthToken},
	}, "update", rawData)

	fmt.Printf("status %d err: %s\nbody: %s\nAT:%s\n", status, err, rv, godebug.LF())
}

type DataType struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

func ReadData(fn string, startId, endId int) string {
	csvFile, err := os.Open(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data, unable to open file: %s\n", err)
		os.Exit(1)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	type DataMeta struct {
		Data []DataType `json:"data"`
	}
	var data DataMeta
	line_no := 0
	for {
		line_no++
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading data, unable to open file: %s\n", err)
			os.Exit(1)
		}
		// fmt.Printf("line = %s\n", godebug.SVarI(line))
		// template process at this point?
		// http://www.2c-why.com/demo3?id={{.id36}},{{.id}}
		mdata := make(map[string]interface{})
		mdata["URL"] = line[2]
		mdata["ID"] = line[0]
		mdata["ID10"] = line[0]
		mdata["ID36"] = line[1]
		if line[0] == "" {
			nb, err := strconv.ParseInt(line[1], 36, 64)
			if err != nil {
			}
			s := fmt.Sprintf("%d", nb)
			line[0] = s
			mdata["ID"] = line[0]
			mdata["ID10"] = line[0]
		}
		v, err := strconv.ParseInt(line[0], 10, 64)
		if err != nil {
			fmt.Printf("Error: unable to parse int [%s] error %s line no: %d\n", line[1], err, line_no)
		}
		vv := int(v)
		mdata["id"] = vv
		mdata["id10"] = vv
		mdata["id36"] = strconv.FormatUint(uint64(vv), 36) // Base 36, take count of # of files add 1, this is the code.
		URLfinal := ExecuteATemplate(line[2], mdata)
		data.Data = append(data.Data, DataType{
			URL: URLfinal,
			ID:  fmt.Sprintf("%v", mdata["id"]),
		})
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("raw data -->>%s<<--\n", dataJSON)
	return string(dataJSON)
}

func ExecuteATemplate(tmpl string, data map[string]interface{}) (rv string) {
	funcMapTmpl := template.FuncMap{
		"PadR":        ms.PadOnRight,
		"PadL":        ms.PadOnLeft,
		"PicTime":     ms.PicTime,
		"FTime":       ms.StrFTime,
		"PicFloat":    ms.PicFloat,
		"nvl":         ms.Nvl,
		"Concat":      ms.Concat,
		"title":       strings.Title, // The name "title" is what the function will be called in the template text.
		"ifDef":       ms.IfDef,
		"ifIsDef":     ms.IfIsDef,
		"ifIsNotNull": ms.IfIsNotNull,
		"dirname":     filepath.Dir,
		"basename":    filepath.Base,
	}
	t := template.New("line-template").Funcs(funcMapTmpl)
	t, err := t.Parse(tmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(102): Invalid template: %s\n", err)
		return tmpl
	}

	// Create an io.Writer to write to a string
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	err = t.ExecuteTemplate(foo, "line-template", data)
	// err = t.ExecuteTemplate(os.Stdout, "line-template", data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error(103): Invalid template processing: %s\n", err)
		return tmpl
	}
	foo.Flush()
	rv = b.String() // Fetch the data back from the buffer
	return
}

var db1 = false
var db11 = false