package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// TransformationRule represents a function that transforms a value based on a schema key type.
type TransformationRule func(interface{}) interface{}

// TransformRules is a map that associates each schema key type with its corresponding transformation function.
var (
	TransformRules = map[string]TransformationRule{
		"S":    FormatString,
		"N":    FormatNum,
		"BOOL": FormatBool,
		"NULL": FormatNull,
		"M":    FormatMap,
		"L":    FormatList,
	}
)

func main() {
	// Parse command-line flags
	schemaFlag := flag.String("config", "schema.json", "Used to read the json file")
	flag.Parse()

	// Read and parse the schema file
	inputMap, err := ParseSchema(*schemaFlag)
	if err != nil {
		fmt.Println("error :", err)
		return
	}

	// Transform the JSON according to the schema rules
	output := TransformJSON(inputMap)

	// Marshal the transformed JSON and print it
	out, err := json.Marshal(output)
	if err != nil {
		fmt.Println("error :", err)
		return
	}
	fmt.Println(string(out))
}

// TransformJSON recursively applies transformation rules to the input JSON.
func TransformJSON(inputMap map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	// Iterate over each key-value pair in the input JSON
	for key, value := range inputMap {
		key = sanitizeKey(key)
		if key == "" {
			continue
		}

		outMap := make(map[string]interface{})

		// Check if the value is a map (object)
		if val, ok := value.(map[string]interface{}); ok {
			// Iterate over each key-value pair in the nested map
			for k, v := range val {
				k = sanitizeKey(k)
				// Apply transformation rule if one exists for the key type
				if rule, ok := TransformRules[k]; ok {
					outMap[key] = rule(v)
				}
			}
		}

		// Merge transformed values into the output map
		if len(outMap) > 0 {
			for k, v := range outMap {
				output[k] = v
			}
		}
	}

	return output
}

// FormatString transforms string values, converting RFC3339 formatted strings to Unix Epoch.
func FormatString(v interface{}) interface{} {
	strVal := v.(string)
	if t, err := time.Parse(time.RFC3339, strVal); err == nil {
		return t.Unix()
	}
	return strVal
}

// FormatNum transforms numeric values, parsing them into float64.
func FormatNum(v interface{}) interface{} {
	numStr := v.(string)
	num := 0.0
	if val, err := strconv.ParseFloat(numStr, 64); err == nil {
		num = val
	}
	return num
}

// FormatBool transforms boolean values.
func FormatBool(v interface{}) interface{} {
	boolStr := v.(string)
	switch boolStr {
	case "1", "t", "true":
		return true
	default:
		return false
	}
}

// FormatNull transforms null values.
func FormatNull(v interface{}) interface{} {
	return nil // Always returns nil for NULL type
}

// FormatMap recursively transforms nested maps (objects).
func FormatMap(v interface{}) interface{} {
	submap := v.(map[string]interface{})
	return TransformJSON(submap)
}

// FormatList transforms list values (arrays).
func FormatList(v interface{}) interface{} {
	listValue := v.([]interface{})
	outList := make([]interface{}, 0)
	for _, listItem := range listValue {
		if val, ok := listItem.(map[string]interface{}); ok {
			outList = append(outList, TransformJSON(val))
		}
	}
	return outList
}

// sanitizeKey trims leading and trailing whitespace from a key.
func sanitizeKey(key string) string {
	return strings.TrimSpace(key)
}

// ParseSchema reads and parses the JSON schema file.
func ParseSchema(fileName string) (map[string]interface{}, error) {
	// Check if the file is a JSON file
	if !strings.Contains(fileName, ".json") {
		return nil, errors.New("config file is not a JSON file")
	}

	// Read the contents of the file
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into a map
	var output map[string]interface{}
	if err := json.Unmarshal(fileBytes, &output); err != nil {
		return nil, err
	}

	return output, nil
}
