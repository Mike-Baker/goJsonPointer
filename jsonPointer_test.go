package jsonPointer

import (
	"encoding/json"
	"fmt"
	"testing"
)

var ietfTestDocument string = `
{
	"foo": ["bar", "baz"],
	"": 0,
	"a/b": 1,
	"c%d": 2,
	"e^f": 3,
	"g|h": 4,
	"i\\j": 5,
	"k\"l": 6,
	" ": 7,
	"m~n": 8
}`

func TestFromWebsite(t *testing.T) {
	var testJson interface{}
	json.Unmarshal([]byte(ietfTestDocument), &testJson)

	// Mising the "" test because ordering on the json object is hard
	testCases := map[string]string{
		//"":       testJsonString,
		"/foo":   `["bar","baz"]`,
		"/foo/0": "\"bar\"",
		"/":      "0",
		"/a~1b":  "1",
		"/c%d":   "2",
		"/e^f":   "3",
		"/g|h":   "4",
		"/i\\j":  "5",
		"/k\"l":  "6",
		"/ ":     "7",
		"/m~0n":  "8",
	}

	for pointer, expected := range testCases {
		testPointerGet(t, testJson, pointer, expected)
	}
}

func testPointerGet(t *testing.T, document interface{}, pointer, expected string) {
	p := Pointer(pointer)
	result, err := p.Get(document)
	if err != nil {
		t.Error(err.Error())
	}

	resultBytes, _ := json.Marshal(result)
	resultString := string(resultBytes)
	if resultString != expected {
		t.Errorf("Key \"%s\" was supposed to produce value %s, instead produced value %s", pointer, expected, resultString)
	}
}

func TestEmptyPointer(t *testing.T) {
	var testJson interface{}
	json.Unmarshal([]byte(ietfTestDocument), &testJson)

	p := Pointer("")
	result, err := p.Get(testJson)
	if err != nil {
		t.Error(err.Error())
	}

	if !ifacePointersEqual(result, testJson) {
		t.Errorf("Key \"%s\" was supposed to produce value %p, instead produced value %p", "", testJson, result)
	}
}

// There has to be a better way of doing this
func ifacePointersEqual(a, b interface{}) bool {
	return fmt.Sprintf("%p", a) == fmt.Sprintf("%p", b)
}

func TestSetValue(t *testing.T) {
	type testCase struct {
		setPtr Pointer
		value  interface{}
		getPtr Pointer
	}

	var testDocumentString string = `
	{
		"": [],
		"array": [1, 2, 3],
		"object": {
			"array": ["a", "b", "c"],
			"foo": "bar"
		}
	}`

	testCases := []testCase{
		testCase{BuildPointer("array", "-"), 5, BuildPointer("array", "3")},
		testCase{BuildPointer("object", "array", "-"), "d", BuildPointer("object", "array", "3")},
		testCase{BuildPointer("", "-"), "abc", BuildPointer("", "0")},
		testCase{BuildPointer(""), "abc", BuildPointer("")},
		testCase{BuildPointer("array", "0"), "-1", BuildPointer("array", "0")},
	}

	var testJson interface{}
	json.Unmarshal([]byte(testDocumentString), &testJson)

	for _, test := range testCases {
		testSetValue(t, testJson, test.setPtr, test.getPtr, test.value)
	}
}

func testSetValue(t *testing.T, document interface{}, setPtr, getPtr Pointer, value interface{}) {
	err := setPtr.Set(&document, value)
	if err != nil {
		t.Error("Set error:", err.Error())
		return
	}

	result, err := getPtr.Get(document)
	if err != nil {
		t.Error("Get error:", err.Error())
		return
	}

	if !ifacePointersEqual(value, result) {
		t.Errorf("Pointer \"%s\" was supposed to produce value %v, instead has produced value %v", string(getPtr), value, result)
	}
}

func TestDecoding(t *testing.T) {
	encoded := "~01"
	expected := "~1"
	decoded := PathTokenToString(encoded)
	if decoded != expected {
		t.Errorf("Decoding \"%s\" was supposed to produce \"%s\", instead has produced value \"%s\"", encoded, expected, decoded)
	}
}
