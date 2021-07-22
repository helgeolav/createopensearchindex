package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

var httpListenPort = flag.String("listenport", ":8888", "what the webserver listens to")

type CollectedData struct {
	Type  string `json:"type,omitempty"`
	Count int    `json:"count,omitempty"`
}

// Key is used during parsing to keep name and type
type Key struct {
	Name string
	Type string
}

// WebCollector contains the main structure for the web server
type WebCollector struct {
	FailedHttp  uint64                   `json:"failed_http,omitempty"`  // number of times an error occured on web parsing
	SuccessHttp uint64                   `json:"success_http,omitempty"` // successful HTTP calls received
	TotalKeys   uint64                   `json:"total_keys,omitempty"`   // total number of keys found
	TotalAdd    uint64                   `json:"total_add,omitempty"`    // total number of times Add was called
	mtx         sync.Mutex               `json:"-"`
	Data        map[string]CollectedData `json:"data,omitempty"`
}

func NewWebCollector() *WebCollector {
	return &WebCollector{
		Data: make(map[string]CollectedData),
	}
}

// Stop stops the web server and prints the output before returning
func (wc *WebCollector) Stop() {
	result := ConfigFile{WebCollector: wc}
	// process data
	result.Input = wc.GetResult()
	// save to output file
	outBytes, err := json.Marshal(&result)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(*outputFile) > 0 {
		err = ioutil.WriteFile(*outputFile, outBytes, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println(string(outBytes))
	}
}

// fieldTypePriorityList is priorities of types, meaning a field can be "upgraded" to the left on the list
var fieldTypePriorityList = []string{"text", "float", "integer", "bool"}

// upgradeFieldType checks if the field type have to change to a better matching type
// based on detection of types.
func upgradeFieldType(currentKey, newKey string) string {
	// if no change
	if currentKey == newKey {
		return currentKey
	}
	// if empty newKey
	if len(newKey) == 0 {
		return currentKey
	}
	const notFound = -1
	// func to find the position in the list, or notFound if not found
	getPos := func(name string) int {
		x := 0
		for _, v := range fieldTypePriorityList {
			if v == name {
				return x
			}
			x++
		}
		return notFound
	}
	// see if field is in list
	cu := getPos(currentKey)
	ne := getPos(newKey)
	// if one of the keys are not in the list return the current key
	if cu == notFound || ne == notFound {
		return currentKey
	}
	if cu < ne {
		return currentKey
	} else {
		return newKey
	}
}

// Add increases the number for all the given keys
func (wc *WebCollector) Add(keys []Key) {
	atomic.AddUint64(&wc.TotalKeys, uint64(len(keys)))
	if len(keys) == 0 {
		return
	}
	atomic.AddUint64(&wc.TotalAdd, 1)
	wc.mtx.Lock()
	for _, key := range keys {
		if keyData, ok := wc.Data[key.Name]; ok {
			keyData.Count++
			keyData.Type = upgradeFieldType(keyData.Type, key.Type)
			wc.Data[key.Name] = keyData
		} else {
			d := CollectedData{Count: 1, Type: key.Type}
			wc.Data[key.Name] = d
		}
	}
	wc.mtx.Unlock()
}

// webPing responds OK with a short text
func webPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// RunWebServer starts the webserver. This method does not return.
func RunWebServer() {
	wc := NewWebCollector()
	// start ctrl+c handler
	ctrlc := make(chan os.Signal)
	signal.Notify(ctrlc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrlc
		fmt.Println("\rCtrl+C pressed in Terminal, stopping")
		wc.Stop()
		os.Exit(0)
	}()
	// webserver handler
	webserverFunc := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			atomic.AddUint64(&wc.FailedHttp, 1)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(body) == 0 || r.Method != http.MethodPost {
			atomic.AddUint64(&wc.FailedHttp, 1)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// go through body
		var inputBody map[string]interface{}
		err = json.Unmarshal(body, &inputBody)
		if err != nil {
			atomic.AddUint64(&wc.FailedHttp, 1)
			w.WriteHeader(http.StatusConflict)
			return
		}
		// return
		keys := findKeywords("", inputBody)
		wc.Add(keys)
		atomic.AddUint64(&wc.SuccessHttp, 1)
		w.WriteHeader(http.StatusOK)
	}
	// setup webserver
	http.HandleFunc("/post", webserverFunc)
	http.HandleFunc("/ping", webPing)
	fmt.Println("Webserver started")
	http.ListenAndServe(*httpListenPort, nil)
}

// findKeywords returns a list of keywords
func findKeywords(srcParent string, input map[string]interface{}) []Key {
	var result []Key
	var parent string
	for k, v := range input {
		key := Key{Name: srcParent + k}
		switch t := v.(type) {
		case map[string]interface{}:
			if len(srcParent) > 0 {
				parent = srcParent + k + "."
			} else {
				parent = k + "."
			}
			res := findKeywords(parent, t)
			result = append(result, res...)
		default:
			key.Type = GuessTypeOf(t)
			result = append(result, key)
		}
	}
	return result
}

// GuessTypeOf tries to guess the type of input, returns type if not sure
func GuessTypeOf(input interface{}) (result string) {
	switch t := input.(type) {
	case []interface{}:
		if len(t) > 0 {
			return GuessTypeOf(t[0])
		}
	case float64, float32, []float32, []float64:
		// float can be returned on integers, check content
		str := fmt.Sprintf("%v", t)
		_, err := strconv.Atoi(str)
		if err == nil {
			result = "integer"
		} else {
			result = "float"
		}
	case int, uint, int32, int64, uint64, uint32, []int, []uint, []int32, []int64:
		result = "integer"
	case bool, []bool:
		result = "boolean"
	case string, []string:
		result = "text"
	default:
		result = "unknown-" + reflect.TypeOf(input).String()
	}
	return
}

// addKey creates the struct based on name
func (i InputStruct) addKey(name, fieldType string) {
	splits := strings.Split(name, ".")
	// simple case - no . in name
	if len(splits) < 2 {
		i[name] = fieldType
		return
	}
	endName := splits[len(splits)-1]
	// we have a dot
	var ptr interface{}
	ptr = i
	for x := 0; x < len(splits)-1; x++ {
		name := splits[x]
		if ptr2, ok := ptr.(InputStruct)[name]; !ok {
			n := make(InputStruct)
			ptr.(InputStruct)[name] = n
			ptr = n
		} else {
			ptr = ptr2.(InputStruct)
		}
	}
	ptr.(InputStruct)[endName] = fieldType
}

// GetResult returns an InputStruct with the parsed data
func (wc *WebCollector) GetResult() InputStruct {
	result := make(InputStruct)
	wc.mtx.Lock()
	myData := wc.Data
	wc.mtx.Unlock()

	for k, v := range myData {
		result.addKey(k, v.Type)
	}

	return result
}
