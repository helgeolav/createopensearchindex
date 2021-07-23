package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

// Doc for explicit mapping: https://opensearch.org/docs/opensearch/rest-api/create-index/#explicit-mapping
// Doc for index template: https://opensearch.org/docs/opensearch/index-templates/
// Doc for field types: https://opensearch.org/docs/search-plugins/ppl/datatypes/

// InputStruct is used to get generic mappings
type InputStruct map[string]interface{}

// ConfigFile is the input file that we read
type ConfigFile struct {
	Input           InputStruct   `json:"input,omitempty"`
	Patterns        []string      `json:"patterns,omitempty"`
	SupportedFields []string      `json:"supported_fields,omitempty"`
	WebCollector    *WebCollector `json:"web_collecor,omitempty"`
}

var (
	inputFile           = flag.String("input", "", "name of input file")
	outputFile          = flag.String("output", "", "name of output file")
	isTemplate          = flag.Bool("template", false, "true if template")
	mode                = flag.String("mode", "createindex", "type of operation")
	myInputFile         ConfigFile
	supportedFieldTypes = []string{"text", "integer", "ip", "geo_point", "float"}
	errorCount          = 0 // used with os.Exit
)

func loadConfig() error {
	var err error
	// load it
	inputBytes, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// parse it
	err = json.Unmarshal(inputBytes, &myInputFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if len(myInputFile.SupportedFields) > 0 {
		supportedFieldTypes = myInputFile.SupportedFields
	}
	return err
}

func main() {
	flag.Parse()
	switch *mode {
	case "createindex":
		// check that we have an input file
		if len(*inputFile) == 0 {
			flag.PrintDefaults()
			errorCount++
		}
		// load config
		if err := loadConfig(); err != nil {
			fmt.Println(err)
			errorCount++
		}
		result := TemplateGenerator()
		err := saveOutput(result)
		if err != nil {
			fmt.Println(err)
			errorCount++
		}
	case "webserver":
		RunWebServer()
	default:
		flag.PrintDefaults()
		errorCount++
	}
	os.Exit(errorCount)
}

// saveOutput saves result to file if specified in outputFile or to stdout
func saveOutput(o interface{}) error {
	res, err := json.Marshal(&o)
	if err == nil {
		if len((*outputFile)) > 0 {
			err = ioutil.WriteFile(*outputFile, res, 0644)
		} else {
			fmt.Println(string(res))
		}
	}
	return err
}
