package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/DoomLordor/hellforge/helpers"
	"github.com/DoomLordor/hellforge/logger"
	"github.com/DoomLordor/hellforge/nats-gateway"
	pb "github.com/DoomLordor/hellforge/pkg/api/gateway/v1"
)

type Client interface {
	grpc.ClientConnInterface
	Close()
}

type client struct {
	logger   zerolog.Logger
	subject  string
	natsConn *nats.Conn
	config   *config
}

func NewClient(subject string, natsConn *nats.Conn, options ...Option) Client {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	return &client{
		logger:   logger.NewLogger("nats-client"),
		subject:  subject,
		natsConn: natsConn,
		config:   cfg,
	}
}

func (c *client) Invoke(ctx context.Context, method string, args, reply any, _ ...grpc.CallOption) (err error) {
	start := time.Now()
	var span trace.Span
	if c.config.tracer != nil {
		ctx, span = c.config.tracer.Start(ctx, "nats-client-invoke")
		span.SetAttributes(attribute.String("method", method))
		defer func() {
			if err != nil {
				span.SetStatus(otelcodes.Error, err.Error())
			} else {
				span.SetStatus(otelcodes.Ok, "success")
			}

			span.End()
		}()
	}

	protoArgs, ok := args.(proto.Message)
	if !ok {
		return ErrArgsNotProtoType
	}

	protoReply, ok := reply.(proto.Message)
	if !ok {
		return ErrReplyNotProtoType
	}

	requestData, err := proto.Marshal(protoArgs)
	if err != nil {
		return fmt.Errorf("proto args marshal error: %w", err)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	otel.GetTextMapPropagator().Inject(ctx, helpers.NewMetadataCarrier(md))

	request := &pb.NatsGatewayRequest{
		Method:   method,
		Body:     requestData,
		Metadata: natsgateway.ConvertMetadataToProto(md),
	}

	body, err := proto.Marshal(request)
	if err != nil {
		return fmt.Errorf("proto request marshal error: %w", err)
	}

	if len(body) > c.config.maxMessageSize {
		return ErrMaxMessageSize
	}

	if span != nil {
		span.SetAttributes(attribute.String("prepare", time.Since(start).String()))
	}

	msg, err := c.natsConn.Request(c.subject, body, c.config.timeout)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}

	var response pb.NatsGatewayResponse
	err = proto.Unmarshal(msg.Data, &response)
	if err != nil {
		return fmt.Errorf("proto response umarshal error: %w", err)
	}

	_ = grpc.SetHeader(ctx, natsgateway.ConvertMetadataFromProto(response.Metadata))

	if response.Status != nil {
		return status.ErrorProto(response.Status)
	}

	err = proto.Unmarshal(response.Body, protoReply)
	if err != nil {
		return fmt.Errorf("proto reply umarshal error: %w", err)
	}

	return nil
}

func (c *client) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}

func (c *client) Close() {
	c.natsConn.Close()
}
