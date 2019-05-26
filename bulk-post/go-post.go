package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/check-json-syntax/lib"
	"github.com/pschlump/godebug"
	ms "github.com/pschlump/templatestrings"
)

// TODO
// 1. --rpt - format return value - xyzzyRpt1
// 2. Make rpt in 2 formats
//		1. as a .zip with all the QR images and a CSV file with info on each.
//		2. as a "HTML" set file wwith a URL - to open from the server.

// ConfigType is the global configuration that is read in from cfg.json
type ConfigType struct {
	HostURLPort string   // http://localhost:9001 for example
	AuthToken   string   //
	DebugFlags  []string // List of debug flags - set by default	-- TODO
}

var gCfg ConfigType

func init() {
	gCfg = ConfigType{
		HostURLPort: "http://192.168.0.157:9001", // note the lack of a '/' at the end.
		AuthToken:   "$ENV$QR_SHORT_AUTH_TOKEN",
		DebugFlags:  make([]string, 0, 10),
	}
}

func main() {
	var err error

	var Cfg = flag.String("cfg", "cfg.json", "config file, default ./cfg.json")
	var Data = flag.String("data", "data.csv", "Input to bulk update")
	var Rpt = flag.String("rpt", "", "Format output as a report, - for stdout else file name")
	var StartID = flag.Int("startId", -1, "Starting ID value for automatic generation")
	var EndID = flag.Int("endId", -1, "Ending ID value for automatic generation")

	flag.Parse()
	fns := flag.Args()
	if len(fns) > 0 {
		fmt.Printf("Usage: bulk-post [--cfg cfg.json] --data data.csv [ --rpt output-file ]\n")
		os.Exit(1)
	}

	if Cfg != nil && *Cfg != "" {
		gCfg, err = ReadConfig(*Cfg, gCfg)
		if err != nil {
			os.Exit(1)
		}
	}

	rawData := ReadData(*Data, *StartID, *EndID)
	// rawData := `{"data":[{"url":"http://www.2c-why.com/demo3?id=3&x=","id":"3"},{"url":"http://www.2c-why.com/demo3?id=57&x=","id":"57"}]}`

	form := url.Values{}
	form.Set("update", rawData)

	encbody := form.Encode()
	body := bytes.NewBufferString(encbody)
	_, _ = body, rawData

	bodyStr := fmt.Sprintf("%v", encbody)
	fmt.Printf("body ->%s<- len %d\n", bodyStr, len(bodyStr))

	client := &http.Client{}

	req, err := http.NewRequest("POST", "http://192.168.0.157:9001/bulkLoad", strings.NewReader(bodyStr)) // Fixed!
	// req, err := http.NewRequest("GET", fmt.Sprintf("http://192.168.0.157:9001/bulkLoad?update=%s&_ran=%d", rawData, rand.Intn(999999999)), nil)

	req.Header.Add("User-Agent", "GoClient_1.10.2") // Informative but not needed
	req.Header.Add("X-Qr-Auth", gCfg.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded") //

	// fmt.Printf("req before send: %s\n", godebug.SVarI(req))

	resp, err := client.Do(req)
	defer resp.Body.Close()

	if *Rpt == "" {
		fmt.Printf("resp: %s err: %s AT:%s\n", godebug.SVarI(resp), err, godebug.LF())
	} else {
		// xyzzyRpt1 -- parse JSON and format as a report.
		fmt.Printf("Error: not implemented yet --rpt flag TODO\n")
	}
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
		// template process at this point?
		// http://www.2c-why.com/demo3?id={{.id36}},{{.id}}
		mdata := make(map[string]interface{})
		if startId >= 0 && endId > startId {
			mdata["URL"] = line[0]
			mdata["ID"] = line[1]
			for vv := startId; vv < endId; vv++ {
				mdata["id"] = vv
				mdata["id36"] = strconv.FormatUint(uint64(vv), 36) // Base 36, take count of # of files add 1, this is the code.
				URLfinal := ExecuteATemplate(line[0], mdata)
				IDfinal := ExecuteATemplate(line[1], mdata)
				data.Data = append(data.Data, DataType{
					URL: URLfinal,
					ID:  IDfinal,
				})
			}
		} else {
			mdata["URL"] = line[0]
			mdata["ID"] = line[1]
			v, err := strconv.ParseInt(line[1], 10, 64)
			if err != nil {
				fmt.Printf("Error: unable to parse int [%s] error %s line no: %d\n", line[1], err, line_no)
			}
			vv := int(v)
			mdata["id"] = vv
			mdata["id36"] = strconv.FormatUint(uint64(vv), 36) // Base 36, take count of # of files add 1, this is the code.
			URLfinal := ExecuteATemplate(line[0], mdata)
			data.Data = append(data.Data, DataType{
				URL: URLfinal,
				ID:  line[1],
			})
		}
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("raw data -->>%s<<--\n", dataJSON)
	return string(dataJSON)
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

	// AuthToken:     "ENV$QR_SHORT_AUTH_TOKEN"
	if strings.HasPrefix(rv.AuthToken, "$ENV$") {
		name := rv.AuthToken[4:]
		val := os.Getenv(name)
		rv.AuthToken = val
	}

	if db11 {
		fmt.Printf("rv=%s, %s\n", godebug.SVarI(rv), godebug.LF())
	}

	return rv, nil
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
		"dirname":     filepath.Dir, // xyzzyTemplateAdd - basename, dirname,
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
