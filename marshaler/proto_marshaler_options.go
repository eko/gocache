package marshaler

import (
	"google.golang.org/protobuf/encoding/protojson"
)

type ProtoMarshalerOption func(*ProtoMarshaler)

func WithMarshalerOption(opt protojson.MarshalOptions) ProtoMarshalerOption {
	return func(pm *ProtoMarshaler) {
		pm.marshalOpts = opt
	}
}

func WithUnmarshalerOption(opt protojson.UnmarshalOptions) ProtoMarshalerOption {
	return func(pm *ProtoMarshaler) {
		pm.unmarshalOpts = opt
	}
}
