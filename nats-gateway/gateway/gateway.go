package gateway

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/DoomLordor/hellforge/helpers"
	"github.com/DoomLordor/hellforge/logger"
	"github.com/DoomLordor/hellforge/nats-gateway"
	pb "github.com/DoomLordor/hellforge/pkg/api/gateway/v1"
)

type Gateway interface {
	Run(ctx context.Context) error
	Shutdown() error
}

type gateway struct {
	logger        zerolog.Logger
	subject       string
	queueGroup    string
	natsConn      *nats.Conn
	grpcConn      *grpc.ClientConn
	tracer        trace.Tracer
	config        *config
	adapterMap    map[string]Adapter
	wg            sync.WaitGroup
	unsubscribers []func() error
}

func NewGateway(grpcAddr, subject, queueGroup string, natsConn *nats.Conn, options ...Option) (Gateway, error) {
	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	dialOpts = append(dialOpts, cfg.dialOpts...)

	grpcConn, err := grpc.NewClient(grpcAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &gateway{
		logger:        logger.NewLogger("nats-gateway"),
		subject:       subject,
		queueGroup:    queueGroup,
		natsConn:      natsConn,
		grpcConn:      grpcConn,
		tracer:        helpers.ProviderToTracer(cfg.provider, "nats-gateway"),
		config:        cfg,
		adapterMap:    cfg.getAdaptersMap(),
		wg:            sync.WaitGroup{},
		unsubscribers: make([]func() error, 0, 10),
	}, nil
}

func (g *gateway) Run(ctx context.Context) error {
	ch := make(chan *nats.Msg, g.config.limit)
	sub, err := g.natsConn.ChanQueueSubscribe(g.subject, g.queueGroup, ch)
	if err != nil {
		return err
	}

	g.unsubscribers = append(g.unsubscribers, func() error {
		defer close(ch)
		return sub.Unsubscribe()
	})

	g.wg.Go(func() {
		done := ctx.Done()
		for {
			select {
			case <-done:
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				go g.call(msg)
			}
		}
	})

	return nil
}

func (g *gateway) call(msg *nats.Msg) {

	ctx := context.Background()

	resp := g.handle(ctx, msg.Data)

	body, err := proto.Marshal(convertResponseToProto(resp))
	if err != nil {
		g.logger.Err(err).Msg("Error marshalling response")
		return
	}

	err = msg.Respond(body)
	if err != nil {
		g.logger.Err(err).Msg("Error respond message")
		return
	}
}

func (g *gateway) handle(ctx context.Context, body []byte) (response *Response) {
	defer func() {
		r := recover()
		if r != nil {
			g.logger.Err(errors.New("panic")).
				Str("panic", fmt.Sprintf("%v", r)).
				Msgf("recovered stack: %s", string(debug.Stack()))
			response = &Response{
				Status: status.New(codes.Internal, "Internal Server Error (panic)"),
			}
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, g.config.handleTimeout)
	defer cancel()

	if len(body) > g.config.maxMessageSize {
		return &Response{
			Status: status.New(codes.Canceled, "message too large"),
		}
	}

	var request pb.NatsGatewayRequest
	err := proto.Unmarshal(body, &request)
	if err != nil {
		g.logger.Err(err).Msg("Error unmarshalling request")
		return &Response{
			Status: status.New(codes.InvalidArgument, "Error unmarshalling request"),
		}
	}

	md := natsgateway.ConvertMetadataFromProto(request.Metadata)
	ctx = metadata.NewOutgoingContext(ctx, md)
	ctx = otel.GetTextMapPropagator().Extract(ctx, helpers.NewMetadataCarrier(md))
	if g.tracer != nil {
		var span trace.Span
		ctx, span = g.tracer.Start(ctx, "call-consumer")
		span.SetAttributes(attribute.String("method", request.Method))
		defer func() {
			if response.Status == nil {
				span.SetStatus(otelcodes.Ok, "success")
			} else {
				span.RecordError(err)
				span.SetStatus(otelcodes.Error, response.Status.Err().Error())
			}

			span.End()
		}()
	}

	adapter, ok := g.adapterMap[request.Method]
	if !ok {
		return &Response{
			Status: status.New(codes.Unimplemented, "method not found"),
		}
	}

	response, err = adapter.Handle(ctx, g.grpcConn, request.Body)
	if err != nil {
		g.logger.Err(err).Send()
		return &Response{
			Status: status.New(codes.Internal, err.Error()),
		}
	}

	return response
}

func (g *gateway) Shutdown() error {
	var multiErr *multierror.Error
	for _, closer := range g.unsubscribers {
		multiErr = multierror.Append(multiErr, closer())
	}

	g.wg.Wait()

	g.natsConn.Close()
	multiErr = multierror.Append(multiErr, g.grpcConn.Close())

	return multiErr.ErrorOrNil()
}
