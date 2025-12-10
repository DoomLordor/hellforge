package gateway

import (
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/DoomLordor/hellforge/nats-gateway"
	pb "github.com/DoomLordor/hellforge/pkg/api/gateway/v1"
)

type serverMetadata struct {
	HeaderMD  metadata.MD
	TrailerMD metadata.MD
}

type Response struct {
	Body     []byte
	Status   *status.Status
	Metadata map[string][]string
}

func convertResponseToProto(response *Response) *pb.NatsGatewayResponse {
	md := make(map[string]*pb.StringArray, len(response.Metadata))
	for k, v := range response.Metadata {
		md[k] = &pb.StringArray{Items: v}
	}

	return &pb.NatsGatewayResponse{
		Body:     response.Body,
		Status:   response.Status.Proto(),
		Metadata: natsgateway.ConvertMetadataToProto(response.Metadata),
	}
}
