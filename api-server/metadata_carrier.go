package apiserver

import (
	"google.golang.org/grpc/metadata"
)

// metadataCarrier type for using MD as open telemetry TextMapCarrier
type metadataCarrier struct {
	md metadata.MD
}

func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for key := range mc.md {
		keys = append(keys, key)
	}

	return keys
}

func (mc metadataCarrier) Get(key string) string {
	values := mc.md.Get(key)
	if len(values) > 0 {
		return values[0]
	}

	return ""
}

func (mc metadataCarrier) Set(key string, value string) {
	mc.md.Append(key, value)
}
