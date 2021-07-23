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

// TODO: remove it - seems not to be needed and used
// Keyword is used under FieldMapper to give som additional properties
type Keyword struct {
	Type        string `json:"type,omitempty"` // will be of value keyword
	IgnoreAbove int    `json:"ignore_above,omitempty"`
}

// Mappings is used to create an intermediate in the output
type Mappings struct {
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// FieldMapper is struct used as map[keyword]FieldMapper
type FieldMapper struct {
	Type       string                 `json:"type"`                 // kind of field
	Properties map[string]interface{} `json:"properties,omitempty"` // used if we have subtypes
	Fields     *Keyword               `json:"fields,omitempty"`     // used to define keyword
}

// TemplateTop is the top level struct for creating templates
type TemplateTop struct {
	IndexPatterns []string               `json:"index_patterns,omitempty"` // list of indices using this template
	Template      map[string]interface{} `json:"template"`                 // The template definitions
	Priority      int                    `json:"priority,omitempty"`       // can be used to specify template priority, higher is better
}

// TemplateSettings are settings that can be specified in the configuration (input) file and sent out with templates
type TemplateSettings struct {
	Shards   int `json:"number_of_shards,omitempty"`   // number of shards
	Replicas int `json:"number_of_replicas,omitempty"` // number of replicas
}

// createField is the workhorse, paring a struct and creating the inner field definition that is common for both template and non-templates. This method is called recursively.
func createField(name string, value interface{}) interface{} {
	switch vt := value.(type) {
	case string:
		if SupportedFieldType(vt) {
			res := FieldMapper{
				Type: vt,
			}
			//TODO: does not seem to be needed / supported for creation
			//if vt == typeString {
			//	res.Fields = &Keyword{
			//		Type:        "keyword",
			//		IgnoreAbove: 256,
			//	}
			//}
			return res
		} else {
			fmt.Fprintln(os.Stderr, name, "has unsupported type", vt)
			errorCount++
			res := FieldMapper{
				Type: vt,
			}
			return res
		}
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range vt {
			result[k] = createField(k, v)
		}
		return FieldMapper{
			Type:       typeNested,
			Properties: result,
		}
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

// IndexMapperGenerator generates either a mapping that can be used to create an index or
// a template that can be applied to many indices.
func IndexMapperGenerator() interface{} {
	// generate template
	var result interface{}
	if *isTemplate {
		myResult := TemplateTop{
			IndexPatterns: myInputFile.Patterns,
			Template:      make(map[string]interface{}),
		}
		tmp := Mappings{Properties: createIndex()}
		myResult.Template[typeMappings] = tmp
		if myInputFile.TemplateSettings != nil {
			myResult.Template[typeSettings] = *myInputFile.TemplateSettings
		}
		result = myResult
	} else {
		tmp := Mappings{Properties: createIndex()}
		tmp2 := make(map[string]interface{})
		tmp2["mappings"] = tmp
		result = tmp2
	}
	// send output to stdout
	return result
}
