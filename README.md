GoJsonPointer
=============

An implementation of the [JSONPointer][#jsonpointer] draft in the Go programming language.


Features
--------
Has functions for turning string slices into JSON Pointers and vice versa.
JSON Pointers can be used to get/set values in the Go representation of abstract javscript (nested []interface{} and map[string]interface{}).
Encodes / and ~ characters in tokens.

Sample Usage
------------
```go
p := jsonPointer.BuildPointer("foo", "bar", 6)
value, err := p.Get(jsonDocument)
err := p.Set(&jsonDocument, "Hello")
```

Project Status
--------------
It passeses the tests such as they are at the moment.
Pull Requests/Bug reports welcome.

[#jsonpointer]:http://tools.ietf.org/html/draft-ietf-appsawg-json-pointer-08
