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
	Input            InputStruct       `json:"input,omitempty"`             // this are the fields we are working on
	Patterns         []string          `json:"patterns,omitempty"`          // patterns to be copied to the template
	SupportedFields  []string          `json:"supported_fields,omitempty"`  // user configurable set of supported fields
	WebCollector     *WebCollector     `json:"web_collecor,omitempty"`      // used by web server to output more statistics after collection
	TemplateSettings *TemplateSettings `json:"template_settings,omitempty"` // optional template settings that are copied to output if present
}

const (
	typeNested   = "nested"    // name of the nested structs
	typeString   = "text"      // name of string type
	typeInt      = "integer"   // name of integer type
	typeIP       = "ip"        // name of IP
	typeDate     = "date"      // name of date
	typeFloat    = "float"     // name of float
	typeGeo      = "geo_point" // name of geo point
	typeKeyword  = "keyword"   // name of keyword type
	typeTemplate = "template"  // for template name in index templates
	typeMappings = "mappings"  // for mappings name in index templates
	typeSettings = "settings"  // for settings name in index templates
)

var (
	inputFile           = flag.String("input", "", "name of input file")
	outputFile          = flag.String("output", "", "name of output file")
	isTemplate          = flag.Bool("template", false, "true if template")
	mode                = flag.String("mode", "createindex", "type of operation")
	myInputFile         ConfigFile
	supportedFieldTypes = []string{typeString, typeInt, typeIP, typeGeo, typeFloat, typeDate, typeNested, typeKeyword}
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
		result := IndexMapperGenerator()
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
	res, err := json.MarshalIndent(&o, "", "  ")
	if err == nil {
		if len((*outputFile)) > 0 {
			err = ioutil.WriteFile(*outputFile, res, 0644)
		} else {
			fmt.Println(string(res))
		}
	}
	return err
}
