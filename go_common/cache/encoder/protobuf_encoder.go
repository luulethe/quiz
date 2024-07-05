package cache_encoder

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type ProtobufEncoder struct{}

func NewProtobufEncoder() *ProtobufEncoder {
	return &ProtobufEncoder{}
}

func (pbe *ProtobufEncoder) Encode(input interface{}) (string, error) {
	protoInput, ok := input.(proto.Message)
	if !ok {
		return "", fmt.Errorf("input is not proto: %v", input)
	}
	out, err := proto.Marshal(protoInput)
	return string(out), err
}

func (pbe *ProtobufEncoder) Decode(input string, output interface{}) error {
	protoOutput, ok := output.(proto.Message)
	if !ok {
		return fmt.Errorf("output is not proto: %v", protoOutput)
	}
	return proto.Unmarshal([]byte(input), protoOutput)
}
