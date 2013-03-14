package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
)

var input_file = flag.String("file", "", "the name of the file that contains the json")
var struct_name = flag.String("struct", "JsonStruct", "the desired name of the struct")


func generateTypes(obj map[string]interface{}, layers int) string {
	structure := "struct {"

	keys := make([]string, 0, len(obj))
	for key, _ := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		curType := reflect.TypeOf(obj[key])
		indentation := "\t"
		for i := 0; i < layers; {
			indentation += "\t"
			i++
		}

		nested, isNested := obj[key].(map[string]interface{})
		if isNested {
			//This is a nested object
			structure += "\n" + indentation + key + " " + generateTypes(nested, layers+1) + "}"
		} else {
			//Check if this is an array
			if objects, ok := obj[key].([]interface{}); ok {
				types := make(map[reflect.Type]bool, 0)
				for _, o := range objects {
					types[reflect.TypeOf(o)] = true
				}

				if len(types) == 1 {
					structure += "\n" + indentation + key + " []" + reflect.TypeOf(objects[0]).Name()
				} else {
					structure += "\n" + indentation + key + " []interface{}"
				}
			} else if curType == nil {
				structure += "\n" + indentation + key + " " + "*interface{}"
			} else {
				structure += "\n" + indentation + key + " " + curType.Name()
			}
		}
	}
	return structure
}

//Given a JSON string representation of an object and a name structName,
//generate the struct definition of the struct, and give it the specified name
func generate(jsn []byte, structName string) (js_s string, err error) {
	result := map[string]interface{}{}
	if err = json.Unmarshal(jsn, &result); err != nil {
		return
	}

	typeString := generateTypes(result, 0)

	typeString = "package main \n type " + structName + " " + typeString + "}"

	fset := token.NewFileSet() // positions are relative to fset

	formatted, err := parser.ParseFile(fset, "", typeString, parser.ParseComments)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, formatted)
	js_s = buf.String()
	return
}

func main() {
	flag.Parse()

	if *input_file == "" {
		//If '-file' was not provided, use the first command-line argument, if one exists
		//If no command-line arguments were provided, panic
		if len(os.Args) > 1 {
			*input_file = os.Args[1]
		} else {
			flag.Usage()
			fmt.Fprintln(os.Stderr, "No input file specified")
			os.Exit(1)
		}
	}

	js, err := ioutil.ReadFile(*input_file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading", *input_file)
		os.Exit(1)
	}

	if output, err := generate(js, *struct_name); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing json", err)
		os.Exit(1)
	} else {
		fmt.Print(output)
	}
}
