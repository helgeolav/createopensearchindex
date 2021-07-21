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
