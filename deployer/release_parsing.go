package deployer

import (
	"bytes"
	"encoding/json"
)

// The goal here is to raise an error if a key is sent that is not supported.
// This should stop many dangerous problems, like misspelling a parameter.

// xRelease is a Release type that can be parsed without overrideing the UnmarshalJSON method

// UnmarshalJSON should error if there is something unexpected
func (release *Release) UnmarshalJSON(data []byte) error {
	type xRelease Release
	var releaseX xRelease
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force error if unknown field is found

	if err := dec.Decode(&releaseX); err != nil {
		return err
	}

	*release = Release(releaseX)
	return nil
}
