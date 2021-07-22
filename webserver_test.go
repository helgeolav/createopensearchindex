package main

import (
	"testing"
)

func TestGuessTypeOf(t *testing.T) {
	type args struct {
		input interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantResult string
	}{
		{"Test bool", args{input: true}, "boolean"},
		{"Test []bool", args{input: []bool{false}}, "boolean"},
		{"Test []int", args{input: []int{323}}, "integer"},
		{"Test []int as interface", args{input: []interface{}{323}}, "integer"},
		{"Test string", args{input: "a string"}, "text"},
		{"Test int", args{input: 5000}, "integer"},
		{"Test float", args{input: 5000.50}, "float"},
		{"Test args", args{args{}}, "unknown-main.args"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := GuessTypeOf(tt.args.input); gotResult != tt.wantResult {
				t.Errorf("GuessTypeOf() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestInputStruct_addKey(t *testing.T) {
	// create input
	i := make(InputStruct)
	i.addKey("root", "rootval")
	i.addKey("sub1.val1", "val1")
	i.addKey("sub1.val2", "val2")
	i.addKey("sub2.sub1.value", "value")
	// validate sub1
	if sub1, ok := i["sub1"]; ok {
		if t1, ok1 := sub1.(InputStruct); ok1 {
			if len(t1) != 2 {
				t.Error("sub1 wrong length")
			}
		} else {
			t.Errorf("sub1 not InputStruct")
		}
	} else {
		t.Error("Could not find sub1")
	}
	// validate sub2
	if sub2, ok := i["sub2"].(InputStruct); ok {
		if sub1, ok1 := sub2["sub1"].(InputStruct); ok1 {
			if len(sub1) != 1 {
				t.Errorf("sub2.sub1 wrong length")
			}
		} else {
			t.Errorf("sub2.sub1 not InputStruct")
		}
	} else {
		t.Errorf("sub2 failed")
	}
}

func Test_upgradeFieldType(t *testing.T) {
	type args struct {
		currentKey string
		newKey     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"from text to bool", args{currentKey: "text", newKey: "bool"}, "text"},
		{"from int to text", args{currentKey: "integer", newKey: "text"}, "text"},
		{"from text to int", args{currentKey: "text", newKey: "integer"}, "text"},
		{"from float to int", args{currentKey: "float", newKey: "integer"}, "float"},
		{"from int to float", args{currentKey: "integer", newKey: "float"}, "float"},
		{"from int to unknown", args{currentKey: "integer", newKey: "unkown"}, "integer"},
		{"from int to int", args{currentKey: "integer", newKey: "integer"}, "integer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := upgradeFieldType(tt.args.currentKey, tt.args.newKey); got != tt.want {
				t.Errorf("upgradeFieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}
