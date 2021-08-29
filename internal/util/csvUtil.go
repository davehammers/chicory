package util

import (
	"encoding/csv"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type csvUtil struct {
	CsvLines    [][]string
	C           chan string
	csvFileName string
	urlIdx      int
}

func New(filename string) *csvUtil {
	return &csvUtil{
		csvFileName: filename,
	}
}

func CommandParams() (csvFileName *string, err error) {
	csvFileName = flag.String("f", "", "CSV file with urls produced from Superset")
	flag.Parse()
	switch {
	case csvFileName == nil:
		err = fmt.Errorf("CSV filename required")
	case *csvFileName == "":
		err = fmt.Errorf("CSV filename missing")
	}
	return
}

func (x *csvUtil) Open() (err error) {
	x.C = make(chan string)
	f, err := os.Open(x.csvFileName)
	if err != nil {
		return
	}
	defer f.Close()
	x.CsvLines, err = csv.NewReader(f).ReadAll()
	if err != nil {
		log.Error(err)
		return
	}
	if len(x.CsvLines) == 0 {
		err = fmt.Errorf("No records found in CSV file %s", x.csvFileName)
		return
	}

	// find the URL column
	for idx, val := range x.CsvLines[0] {
		if val = strings.ToLower(val); val == "url" {
			x.urlIdx = idx
			break
		}
	}
	go func() {
		for idx, val := range x.CsvLines {
			switch idx {
			case 0:
				// skip headings
			default:
				sourceURL := val[x.urlIdx] // GET the url string
				if sourceURL == "" {
					continue
				}
				x.C <- sourceURL
			}
		}
		close(x.C)
	}()
	return
}


func (x *csvUtil) Write() (err error) {
	f, err := os.Create(x.csvFileName + ".csv")
	csv.NewWriter(f).WriteAll(x.CsvLines)
	f.Close()
	return
}
