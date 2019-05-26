package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/American-Certified-Brands/config-sample/ReadConfig"
	"github.com/American-Certified-Brands/tools/get"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	ms "github.com/pschlump/templatestrings"
)

// ConfigType is the global configuration that is read in from cfg.json
type ConfigType struct {
	HostURLPort   string `json:"HostURLPort" default:"http://127.0.0.1:2004"`      // URL of qr-short server
	AuthToken     string `json:"qr_auth_token" default:"$ENV$QR_SHORT_AUTH_TOKEN"` // Auth key for taling to qr-short
	BaseServerURL string `json:"base_server_url" default:"http://127.0.0.1:9022"`  // QR Image Server (qr-micro-service, 127.0.0.1:9022?) // the QR Image Server
	QR_MS_Url     string `json:"QR_MS_Url" default:"http://127.0.0.1:9022"`        // Server that can allocate new QR codes.
	QR_MS_AuthKey string `json:"QR_MS_AuthKey" default:"$ENV$QR_MS_AUTH_TOKEN"`
}

var gCfg ConfigType

var Cfg = flag.String("cfg", "cfg.json", "config file, default ./cfg.json")
var Data = flag.String("data", "data.csv", "Input to bulk update")
var Zip = flag.String("zip", "", "zipama")
var Server = flag.String("server", "http://127.0.0.1:2004", "Local or remote qr-short server.") // http://t432z.com for remote

var db_flag = make(map[string]bool)

func main() {

	flag.Parse()
	fns := flag.Args()
	if len(fns) > 0 {
		fmt.Printf("Usage: bulk-post2 [--cfg cfg.json] --data data.csv [ --zip zip-file ] [ --baseurl http://www.example.com ]\n")
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

	if db_flag["db4"] {
		fmt.Printf("Cfg ->%s<-\n", godebug.SVarI(gCfg))
	}

	rawData, raw := ReadData(*Data)

	if db_flag["db5"] {
		fmt.Printf("rawData ->%s<- raw ->%s<-\n", rawData, godebug.SVarI(raw))
	}

	if db_flag["db5"] {
		fmt.Fprintf(os.Stderr, "Call To ->%s/bulkLoad<-\n", *Server)
	}
	status, rv := get.DoPostHeader(fmt.Sprintf("%s/bulkLoad", *Server), []get.HeaderType{
		{Name: "X-Qr-Auth", Value: gCfg.AuthToken},
	}, "update", rawData)

	if db_flag["db5"] {
		fmt.Printf("status %d err: %s\nbody: %s\nAT:%s\n", status, err, rv, godebug.LF())
	}

	// for each QR Read
	//		Go get the QR Image - valiate that the image is available.
	//			Save image into ./xxx for later zipping
	//		Go decode the QR - get the destination
	//			Check that it is working
	//		Generate index.csv for this.

	nSucc := 0
	nErr := 0
	nSuccDest := 0
	nErrDest := 0
	zipfn := make([]string, 0, len(raw)+1)
	var noEx string
	if strings.HasSuffix(*Zip, ".zip") {
		noEx = RmExt(*Zip)
	} else {
		noEx = *Zip
		*Zip = fmt.Sprintf("%s.zip", *Zip)
	}
	for ii, vv := range raw {
		uri := fmt.Sprintf("%s/qr/qr_%05d/q%05d.4.png", gCfg.BaseServerURL, vv.Id10n/100, vv.Id10n)
		status, img := get.DoGet(uri)
		if status == 200 {
			nSucc++
			if *Zip != "" {
				os.MkdirAll(fmt.Sprintf("./%s", noEx), 0755)
				fn := fmt.Sprintf("./%s/q%05d.4.png", noEx, vv.Id10n)
				ioutil.WriteFile(fn, []byte(img), 0644)
				fn = fmt.Sprintf("./q%05d.4.png", vv.Id10n)
				zipfn = append(zipfn, fn)
				nSucc++
			}
			statusDest, _ := get.DoGet(vv.Url)
			if statusDest == 200 {
				nSuccDest++
			} else {
				nErrDest++
			}
		} else {
			nErr++
			fmt.Printf("%sLine: %d Unable to get ->%s<- image for QR code, status=%d.%s\n", MiscLib.ColorRed, ii+1, uri, status, MiscLib.ColorReset)
		}
	}
	ioutil.WriteFile(fmt.Sprintf("./%s/index.json", noEx), []byte(godebug.SVarI(raw)), 0644)
	zipfn = append(zipfn, "./index.json")
	if nErr == 0 {
		fmt.Printf("%sSuccessfuil fetch of %d QR code images.%s\n", MiscLib.ColorGreen, nSucc, MiscLib.ColorReset)
		fmt.Printf("%sDestiations %d succ %d error%s\n", MiscLib.ColorGreen, nSuccDest, nErrDest, MiscLib.ColorReset)
	}
	if *Zip != "" {
		// Make the .zip file now
		os.Chdir(fmt.Sprintf("./%s", noEx))
		mkZip(fmt.Sprintf("../%s", *Zip), zipfn...)
	}
}

func mkZip(outZipFn string, files ...string) {
	if err := ZipFiles(outZipFn, files); err != nil {
		fmt.Printf("Unable to crate .zip file [%s] error [%s]\n", outZipFn, err)
	}
}

type DataType struct {
	URL string `json:"url"`
	ID  string `json:"id"`
}

type DataReturn struct {
	Url   string
	Id10  string
	Id10n int
	Id36  string
}

func ReadData(fn string) (jsonData string, raw []DataReturn) {
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
		if db_flag["ShowInputLine"] {
			fmt.Printf("line = %s\n", godebug.SVarI(line))
		}
		// http://www.2c-why.com/demo3?id={{.id36}},{{.id}}
		u10 := line[0]
		u36 := line[1]
		UrlTmpl := line[2]
		mdata := make(map[string]interface{}) // Data for template

		mdata["LineNo"] = line_no
		mdata["URL"] = UrlTmpl
		mdata["ID"] = u10
		mdata["ID10"] = u10
		mdata["ID36"] = u36

		if u10 == "" {
			nb, err := strconv.ParseInt(u36, 36, 64)
			if err != nil {
			}
			u10 = fmt.Sprintf("%d", nb)
			mdata["ID"] = u10
			mdata["ID10"] = u10
		}
		v, err := strconv.ParseInt(u10, 10, 64)
		if err != nil {
			fmt.Printf("Error: unable to parse int [%s] error %s line no: %d\n", u36, err, line_no)
		}
		u10n := int(v)
		mdata["id"] = u10n
		mdata["id10"] = u10n
		mdata["id36"] = strconv.FormatUint(uint64(u10n), 36) // Base 36, take count of # of files add 1, this is the code.
		URLfinal := ExecuteATemplate(UrlTmpl, mdata)
		data.Data = append(data.Data, DataType{
			URL: URLfinal,
			ID:  u10,
		})
		raw = append(raw, DataReturn{
			Url:   URLfinal,
			Id10:  u10,
			Id10n: u10n,
			Id36:  u36,
		})
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading data: %s\n", err)
		os.Exit(1)
	}
	if db_flag["db3"] {
		fmt.Printf("raw data -->>%s<<--\n", dataJSON)
	}
	return string(dataJSON), raw
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

func RmExt(filename string) string {
	var extension = filepath.Ext(filename)
	var name = filename[0 : len(filename)-len(extension)]
	return name
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

var db1 = false
var db11 = false
