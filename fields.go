package main

import (
	"fmt"
	"os"
)

// check if field type is supported
func SupportedFieldType(kind string) bool {
	for _, v := range supportedFieldTypes {
		if v == kind {
			return true
		}
	}
	return false
}

type Keyword struct {
	Type        string `json:"type,omitempty"` // will be of value keyword
	IgnoreAbove int    `json:"ignore_above,omitempty"`
}

type Mappings struct {
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// FieldMapper is struct used as map[keyword]FieldMapper
type FieldMapper struct {
	Type       string                 `json:"type"`                 // kind of field
	Properties map[string]interface{} `json:"properties,omitempty"` // used if we have subtypes
	Fields     Keyword                `json:"fields,omitempty"`     // used to define keyword
}

// createField is the workhorse, paring a struct and creating the inner field definition that is common for both template and non-templates. This method is called recursively.
func createField(name string, value interface{}) interface{} {
	switch vt := value.(type) {
	case string:
		if SupportedFieldType(vt) {
			res := FieldMapper{
				Type:       vt,
				Properties: nil,
				Fields: Keyword{
					Type:        "keyword",
					IgnoreAbove: 256,
				},
			}
			return res
		} else {
			fmt.Fprintln(os.Stderr, name, "has unsupported type", vt)
			errorCount++
		}
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range vt {
			result[k] = createField(k, v)
		}
		return result
	}
	return nil
}

// createIndex produces the mapping part - all the fields
func createIndex() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range myInputFile.Input {
		result[k] = createField(k, v)
	}
	return result
}

func TemplateGenerator() interface{} {
	// generate template
	var result interface{}
	if *isTemplate {
		// TODO: add support for templates
		fmt.Fprintln(os.Stderr, "templates are not supported yet")
		errorCount++
	} else {
		tmp := Mappings{Properties: createIndex()}
		tmp2 := make(map[string]interface{})
		tmp2["mappings"] = tmp
		result = tmp2
	}
	// send output to stdout
	return result
}
