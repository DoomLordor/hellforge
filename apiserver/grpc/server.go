package grpc

import (
	"net"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DoomLordor/hellforge/logger"
)

type Grps interface {
	RegisterServer(grpcServer *grpc.Server)
}

type Server struct {
	config     Config
	logger     *logger.Logger
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(config Config) *Server {
	return &Server{
		config:     config,
		logger:     logger.NewLogger("server-grpc"),
		grpcServer: nil,
	}
}

func (s *Server) Configuration(grps []Grps, tracer trace.Tracer) error {
	var err error
	s.listener, err = net.Listen("tcp", s.config.BindAddress())
	if err != nil {
		return err
	}

	metricsCollector := grpcprom.NewServerMetrics()
	prometheus.MustRegister(metricsCollector)

	middlewares := NewMiddlewares(logger.NewLogger("middlewares-grpc"), tracer)

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metricsCollector.UnaryServerInterceptor(),
			recovery.UnaryServerInterceptor(middlewares.RecoveryMiddleware()),
			middlewares.TracingMiddleware(),
			middlewares.TimeMiddleware(),
			middlewares.LoggingMiddleware(),
		),
		grpc.ChainStreamInterceptor(
			metricsCollector.StreamServerInterceptor(),
			recovery.StreamServerInterceptor(middlewares.RecoveryMiddleware()),
			middlewares.LoggingStreamMiddleware(),
		),
	)

	for _, imp := range grps {
		imp.RegisterServer(s.grpcServer)
	}

	reflection.Register(s.grpcServer)

	return nil
}

func (s *Server) Start() {
	if !s.Active() {
		return
	}
	s.logger.Info().Msg("Server grpc start")
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.logger.Err(err).Send()
	}
}

func (s *Server) Stop() {
	if !s.Active() {
		return
	}
	s.grpcServer.GracefulStop()
}

func (s *Server) Active() bool {
	return s.config.Active
}
