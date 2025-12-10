package natsgateway

import (
	pb "github.com/DoomLordor/hellforge/pkg/api/gateway/v1"
)

func ConvertMetadataFromProto(raw map[string]*pb.StringArray) map[string][]string {
	md := make(map[string][]string, len(raw))
	for k, v := range raw {
		md[k] = v.Items
	}

	return md
}

func ConvertMetadataToProto(raw map[string][]string) map[string]*pb.StringArray {
	md := make(map[string]*pb.StringArray, len(raw))
	for k, v := range raw {
		md[k] = &pb.StringArray{Items: v}
	}

	return md
}
