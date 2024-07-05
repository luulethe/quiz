package cache_encoder

import (
	"encoding/json"
)

type JSONEncoder struct{}

func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

func (js *JSONEncoder) Encode(input interface{}) (string, error) {
	r, err := json.Marshal(input)
	return string(r), err
}

func (js *JSONEncoder) Decode(input string, output interface{}) error {
	return json.Unmarshal([]byte(input), output)
}
