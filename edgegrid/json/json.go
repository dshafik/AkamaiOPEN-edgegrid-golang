// Package json adds hooks that are automatically called before marshaling (PreMarshalJSON) and
// after unmarshaling (PostUnmarshalJSON). It does not do so recursively.
package json

import (
	gojson "encoding/json"
)

// Wraps encoding/json.Marshal, calls v.PreMarshalJSON() if it exists
//
// This should probably copy v, otherwise PreMarshalJSON is destructive
func Marshal(v interface{}) ([]byte, error) {
	if _, ok := v.(PreJSONMarshaler); ok {
		err := v.(PreJSONMarshaler).PreMarshalJSON()
		if err != nil {
			return nil, err
		}
	}

	return gojson.Marshal(v)
}

// Wraps encoding/json.Unmarshal, calls v.PostUnmarshalJSON() if it exists
func Unmarshal(data []byte, v interface{}) error {
	err := gojson.Unmarshal(data, v)
	if err != nil {
		return err
	}

	if _, ok := v.(PostJSONUnmarshaler); ok {
		err := v.(PostJSONUnmarshaler).PostUnmarshalJSON()
		if err != nil {
			return err
		}
	}

	return nil
}

type PreJSONMarshaler interface {
	PreMarshalJSON() error
}

func ImplementsPreJSONMarshaler(v interface{}) bool {
	_, ok := v.(PreJSONMarshaler)
	return ok
}

type PostJSONUnmarshaler interface {
	PostUnmarshalJSON() error
}

func ImplementsPostJSONUnmarshaler(v interface{}) (interface{}, bool) {
	v, ok := v.(PostJSONUnmarshaler)
	return v, ok
}
