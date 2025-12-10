package helpers

import (
	"google.golang.org/grpc/metadata"
)

// MetadataCarrier type for using MD as open telemetry TextMapCarrier
type MetadataCarrier struct {
	md metadata.MD
}

func NewMetadataCarrier(md metadata.MD) *MetadataCarrier {
	return &MetadataCarrier{md: md}
}

func (mc MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for key := range mc.md {
		keys = append(keys, key)
	}

	return keys
}

func (mc MetadataCarrier) Get(key string) string {
	if values := mc.md.Get(key); len(values) > 0 {
		return values[0]
	}

	return ""
}

func (mc MetadataCarrier) Set(key string, value string) {
	mc.md.Append(key, value)
}
