package apiserver

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	govalidator "github.com/bufbuild/protovalidate-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/DoomLordor/hellforge/helpers"
	"github.com/DoomLordor/hellforge/logger"
)

type interceptors struct {
	logger zerolog.Logger
	tracer trace.Tracer
}

func newInterceptors(tracer trace.Tracer) *interceptors {
	return &interceptors{
		logger: logger.NewLogger("interceptor"),
		tracer: tracer,
	}
}

func (i *interceptors) grpcPanicRecoveryHandler(interceptorType string) recovery.Option {
	return recovery.WithRecoveryHandler(func(p any) (err error) {
		i.logger.Err(errors.New("panic")).Str("interceptor-type", interceptorType).Msg(string(debug.Stack()))
		return status.Errorf(codes.Internal, "%s", p)
	})
}

func (i *interceptors) withRecoveryUnary() grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(i.grpcPanicRecoveryHandler("unary"))
}

func (i *interceptors) withRecoveryStream() grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor(i.grpcPanicRecoveryHandler("stream"))
}

func (i *interceptors) timeInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now().UnixMilli()
		i.logger.Info().Str("full-method", info.FullMethod).Msg("Start")

		defer func() {
			statusErr, _ := status.FromError(err)
			end := time.Now().UnixMilli() - start
			i.logger.Info().
				Ctx(ctx).
				Str("full-method", info.FullMethod).
				Uint64("code", uint64(statusErr.Code())).
				Int64("response-time", end).
				Msg("End")
		}()

		return handler(ctx, req)
	}
}

func (i *interceptors) withErrorLoggingUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			i.logging(ctx, info.FullMethod, err)
			return nil, err
		}

		return resp, nil
	}
}

func (i *interceptors) withErrorLoggingStream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)

		if err != nil {
			i.logging(ss.Context(), info.FullMethod, err)
			return err
		}

		return nil
	}
}

func (i *interceptors) logging(ctx context.Context, fullMethod string, err error) {
	statusErr, _ := status.FromError(err)
	massageField := ""
	switch statusErr.Code() {
	case codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied,
		codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unauthenticated:
		massageField = "warning"
	case codes.Unknown, codes.DeadlineExceeded, codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
		massageField = "error"
	}

	if massageField != "" {
		i.logger.Warn().Ctx(ctx).Str("full-method", fullMethod).Str(massageField, statusErr.Message()).Send()
	}
}

func (i *interceptors) withTracing() grpc.UnaryServerInterceptor {
	if i.tracer == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Obtain parent propagator if exists
			ctx = otel.GetTextMapPropagator().Extract(ctx, helpers.NewMetadataCarrier(md))
		}

		// Start new parent or child span
		ctx, span := i.tracer.Start(ctx, info.FullMethod)
		defer span.End()

		if ok {
			if sourceService, in := md["source-service"]; in && len(sourceService) > 0 {
				span.SetAttributes(attribute.String("source-service", sourceService[0]))
			}
		}

		resp, err = handler(ctx, req)
		// Mark span status
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		} else {
			span.SetStatus(otelcodes.Ok, "success")
		}

		return resp, err
	}
}

func (i *interceptors) withUnaryValidation() (grpc.UnaryServerInterceptor, error) {
	v, err := govalidator.New()
	if err != nil {
		return nil, err
	}

	return protovalidate.UnaryServerInterceptor(v), nil
}

func (i *interceptors) withStreamValidation() (grpc.StreamServerInterceptor, error) {
	v, err := govalidator.New()
	if err != nil {
		return nil, err
	}

	return protovalidate.StreamServerInterceptor(v), nil
}
