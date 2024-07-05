package cache_encoder

import (
	"bytes"
	"encoding/gob"
)

type GobEncoder struct{}

func NewGobEncoder() *GobEncoder {
	return &GobEncoder{}
}

func (ge *GobEncoder) Encode(input interface{}) (string, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(input)
	if err != nil {
		return "", err
	}
	return buffer.String(), err
}

func (ge *GobEncoder) Decode(input string, output interface{}) error {
	buffer := bytes.NewBufferString(input)
	enc := gob.NewDecoder(buffer)
	return enc.Decode(output)
}
