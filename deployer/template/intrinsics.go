package template

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// decodeRef returns the value of a Ref
func decodeRef(value string) (string, error) {
	// goformation stores intrinsic functions as base64 encoded strings
	var decoded []byte
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// The string value is not base64 encoded, so it's not an intrinsic so just pass it back
		return "", fmt.Errorf("Not a Reference")
	}

	var intrinsic map[string]string
	if err := json.Unmarshal([]byte(decoded), &intrinsic); err != nil {
		// The string value is not JSON, so it's not an intrinsic so just pass it back
		return "", fmt.Errorf("Not a (or only) a Ref")
	}

	// An intrinsic should be an object, with a single key containing a valid intrinsic name
	if len(intrinsic) != 1 {
		return "", fmt.Errorf("Incorrect intrinsic")
	}

	if intrinsic["Ref"] == "" {
		return "", fmt.Errorf("Not a Ref")
	}

	return intrinsic["Ref"], nil
}

// decodeRef returns the value of a Ref
func decodeGetAtt(value string) ([]string, error) {
	// goformation stores intrinsic functions as base64 encoded strings
	var decoded []byte

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// The string value is not base64 encoded, so it's not an intrinsic so just pass it back
		return []string{}, fmt.Errorf("Not a Get Attribute")
	}

	var intrinsic map[string][]string
	if err := json.Unmarshal([]byte(decoded), &intrinsic); err != nil {
		// The string value is not JSON, so it's not an intrinsic so just pass it back
		return []string{}, fmt.Errorf("Not a (or only) a GetAtt")
	}

	// An intrinsic should be an object, with a single key containing a valid intrinsic name
	if len(intrinsic) != 1 {
		return []string{}, fmt.Errorf("Incorrect intrinsic")
	}

	if len(intrinsic["Fn::GetAtt"]) < 1 {
		return []string{}, fmt.Errorf("Not a GetAtt")
	}

	return intrinsic["Fn::GetAtt"], nil
}
