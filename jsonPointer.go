package jsonPointer

import (
	"fmt"
	"strconv"
	"strings"
)

// If someone else is editing this, please feel free to make it conform to the spec in any ways I haven't 
// Also; making the code more efficient/cleaner would be appreciated too 

const (
	sep             string = "/"
	escChar         string = "~"
	escapedSep      string = "~1"
	escapedEscChar  string = "~0"
	newElementToken string = "-"
)

// A pointer to an element in a json structure
// Roughly following http://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-07
type Pointer string

// Given a series of strings will produce an equivelent jsonPointer.Pointer
func BuildPointer(path ...string) Pointer {
	tokens := make([]string, 0, len(path)+1)

	// Add the leading slash
	if len(path) > 0 {
		tokens = append(tokens, "")
	}

	for _, s := range path {
		tokens = append(tokens, StringToPathToken(s))
	}

	pointer := strings.Join(tokens, sep)

	return Pointer(pointer)
}

// Splits a jsonPointer.Pointer into an array of strings
func (ptr Pointer) Split() []string {
	if len(ptr) == 0 {
		return make([]string, 0, 0)
	}

	// Strip the leading /
	ptr = ptr[1:]

	tokens := strings.Split(string(ptr), sep)
	results := make([]string, 0, len(tokens))
	for _, token := range tokens {
		s := PathTokenToString(token)
		results = append(results, s)
	}

	return results
}

// Encodes a string as a path token by escaping / and ~ into ~0 and ~1
func StringToPathToken(s string) string {
	s = strings.Replace(s, escChar, escapedEscChar, -1)
	s = strings.Replace(s, sep, escapedSep, -1)
	return s
}

// Decodes a path token into a string by unescaping ~1 and ~0 into ~ and /
func PathTokenToString(s string) string {
	s = strings.Replace(s, escapedSep, sep, -1)
	s = strings.Replace(s, escapedEscChar, escChar, -1)
	return s
}

/* Delete from here down if you only want to build/split JsonPaths */

// Given a go representation of a json object (go-maps and go-slices) will return the element referenced by this jsonPointer.Pointer
func (ptr Pointer) Get(from interface{}) (result interface{}, err error) {
	tokens := ptr.Split()

	result = from
	for _, s := range tokens {
		if result, err = accessInterface(result, s); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Given a pointer to a go representation of a json object (go-maps and go-slices) will set a value referenced by this jsonPointer.Pointer
func (ptr Pointer) Set(root *interface{}, value interface{}) (err error) {
	tokens := ptr.Split()
	newRoot, err := setValueInternal(*root, tokens, value)
	if err != nil {
		return err
	}

	*root = newRoot
	return nil
}

// Given a pointer to a go representation of a json object (go-maps and go-slices) will set a value referenced by this jsonPointer.Pointer
func setValueInternal(current interface{}, tokens []string, value interface{}) (interface{}, error) {
	if len(tokens) < 1 {
		return nil, fmt.Errorf("When setting a value via a jsonPointer.Pointer the pointer must have atleast one path token")
	} else if len(tokens) == 1 {
		return setValueOnInterface(current, tokens[0], value)
	}

	// The gist of the rest of this:
	// 		Get the next element in the path specified by the pointer
	// 		Recursively call this on that element
	// 		Re-Add the next element to this one incase the next element has changed

	currentToken := tokens[0]
	next, err := accessInterface(current, currentToken)
	if err != nil {
		return nil, err
	}

	newNext, err := setValueInternal(next, tokens[1:], value)
	if err != nil {
		return nil, err
	}

	newCurrent, err := setValueOnInterface(current, currentToken, newNext)
	if err != nil {
		// Using panic here because we have already set the values in maps but not slices
		// The state of the json object is undefined
		panic(err.Error())
	}

	return newCurrent, nil
}

// Access a map or an interface given a string as an index
func accessInterface(itfce interface{}, token string) (interface{}, error) {
	switch itfce.(type) {
	case map[string]interface{}:
		return accessMap(itfce.(map[string]interface{}), token)
	case []interface{}:
		return accessSlice(itfce.([]interface{}), token)
	}

	return nil, fmt.Errorf("Could not access the %s value of element because it was a %T rather than a map[string]interface{} or []interface{}", token, itfce)
}

// Access a map given a string as an index
func accessMap(m map[string]interface{}, token string) (interface{}, error) {
	value, exists := m[token]
	if !exists {
		return nil, fmt.Errorf("Map key \"%s\" not found", token)
	}

	return value, nil
}

// Access a slice given a string as an index
func accessSlice(slice []interface{}, token string) (interface{}, error) {
	if token == newElementToken {
		return nil, fmt.Errorf("Index %s out of range, slice only has %d elements", token, len(slice))
	}

	// Other than - only unsigned integers are allowed
	uIndex, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return nil, err
	}

	index := int(uIndex)

	// Range check
	if index >= len(slice) {
		return nil, fmt.Errorf("Index %d out of range, slice only has %d elements", index, len(slice))
	}

	// Return the needed element
	return slice[index], nil
}

func setValueOnInterface(itfce interface{}, token string, value interface{}) (interface{}, error) {
	switch itfce.(type) {
	case map[string]interface{}:
		m := itfce.(map[string]interface{})
		m[token] = value
		return itfce, nil
	case []interface{}:
		slice := itfce.([]interface{})
		return setValueOnSlice(slice, token, value)
	}

	return nil, fmt.Errorf("Could not access the %s value of element because it was a %T rather than a map[string]interface{} or []interface{}", token, itfce)
}

// Sets a value in a slice given a string as an index
func setValueOnSlice(slice []interface{}, token string, value interface{}) (interface{}, error) {
	if token == newElementToken {
		slice = append(slice, value)
		return slice, nil
	}

	// Other than - only unsigned integers are allowed
	uIndex, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return nil, err
	}

	index := int(uIndex)

	// Range check
	if index >= len(slice) {
		return nil, fmt.Errorf("Index %d out of range, slice only has %d elements, use the index \"-\" to append a new element", index, len(slice))
	}

	// Set the new value
	slice[index] = value
	return slice, nil
}
