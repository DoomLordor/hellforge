package apiserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/gorilla/mux"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/DoomLordor/hellforge/logger"
)

var (
	ConfiguratorNotSetup    = errors.New("configurator not setup")
	FailedToRegisterGateway = errors.New("failed to register gateway")
)

type APIServer struct {
	logger *logger.Logger
	config *config
	ready  atomic.Bool

	httpSystemServer *http.Server
	httpServer       *http.Server
	grpcServer       *grpc.Server

	registry *prometheus.Registry
}

func NewAPIServer(ctx context.Context, configurator Configurator) (*APIServer, error) {
	if configurator == nil {
		return nil, ConfiguratorNotSetup
	}

	options, err := configurator.Configurate(ctx)
	if err != nil {
		return nil, err
	}

	cfg := newConfig()
	for _, option := range options {
		option(cfg)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	registry.MustRegister(
		cfg.metrics...,
	)

	return &APIServer{
		logger:   logger.NewLogger("server"),
		config:   cfg,
		registry: registry,
	}, nil
}

func (s *APIServer) Run(ctx context.Context) error {
	err := s.Start(ctx)
	if err != nil {
		return err
	}

	s.Wait()

	err = s.Stop(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *APIServer) Start(ctx context.Context) error {
	err := s.grpcStart()
	if err != nil {
		s.logger.Err(err).Msg("grpc start error")
		return err
	}

	err = s.httpStart(ctx)
	if err != nil {
		s.logger.Err(err).Msg("http start error")
		return err
	}

	s.httpSystemStart()
	s.ready.Store(true)
	return nil
}

func (s *APIServer) grpcStart() error {
	if !s.config.grpc.enabled {
		return nil
	}

	listener, err := net.Listen("tcp", s.config.grpc.address())
	if err != nil {
		return err
	}

	metricsCollector := grpcprom.NewServerMetrics(grpcprom.WithServerHandlingTimeHistogram())
	s.registry.MustRegister(metricsCollector)

	systemInterceptors := newInterceptors(s.config.tracer)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(s.config.grpc.unaryInterceptors)+6)
	unaryInterceptors = append(
		unaryInterceptors,
		metricsCollector.UnaryServerInterceptor(),
		systemInterceptors.withRecoveryUnary(),
		systemInterceptors.withTracing(),
		systemInterceptors.timeInterceptor(),
		systemInterceptors.withErrorLoggingUnary(),
	)

	unaryInterceptors = append(unaryInterceptors, s.config.grpc.unaryInterceptors...)

	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(s.config.grpc.unaryInterceptors)+4)
	streamInterceptors = append(
		streamInterceptors,
		metricsCollector.StreamServerInterceptor(),
		systemInterceptors.withRecoveryStream(),
		systemInterceptors.withErrorLoggingStream(),
	)

	streamInterceptors = append(streamInterceptors, s.config.grpc.streamInterceptors...)

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			unaryInterceptors...,
		),
		grpc.ChainStreamInterceptor(
			streamInterceptors...,
		),
	)

	metricsCollector.InitializeMetrics(s.grpcServer)

	for _, implementation := range s.config.grpc.implementations {
		implementation.RegisterServer(s.grpcServer)
	}

	if s.config.grpc.withReflect {
		reflection.Register(s.grpcServer)
	}

	go func() {
		s.logger.Info().Msg("Server grpc start")
		err := s.grpcServer.Serve(listener)
		if err != nil {
			s.logger.Err(err).Send()
		}
	}()

	return nil
}

func (s *APIServer) httpSystemStart() {
	if !s.config.httpSystem.enabled {
		return
	}
	router := mux.NewRouter()
	metricsHandler := promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{EnableOpenMetrics: true, Registry: s.registry})
	router.Handle("/metrics", metricsHandler).Methods(http.MethodGet)
	router.HandleFunc("/live", livenessCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/ready", s.readinessCheckHandler).Methods(http.MethodGet)
	router.HandleFunc("/logger", setLogLevel).Methods(http.MethodPost)

	pprofRouter := router.PathPrefix("/debug/pprof").Subrouter()
	pprofRouter.HandleFunc("/", pprof.Index)
	pprofRouter.HandleFunc("/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/profile", pprof.Profile)
	pprofRouter.HandleFunc("/symbol", pprof.Symbol)

	pprofRouter.Handle("/goroutine", pprof.Handler("goroutine"))
	pprofRouter.Handle("/threadcreate", pprof.Handler("threadcreate"))
	pprofRouter.Handle("/mutex", pprof.Handler("mutex"))
	pprofRouter.Handle("/heap", pprof.Handler("heap"))
	pprofRouter.Handle("/block", pprof.Handler("block"))
	pprofRouter.Handle("/allocs", pprof.Handler("allocs"))

	s.httpSystemServer = &http.Server{
		Addr:         s.config.httpSystem.address(),
		WriteTimeout: time.Second * 100,
		ReadTimeout:  time.Second * 100,
		IdleTimeout:  time.Second * 100,
		Handler:      router,
	}

	go func() {
		s.logger.Info().Msg("Server system start")
		if err := s.httpSystemServer.ListenAndServe(); err != nil {
			s.logger.Err(err).Send()
		}
	}()
}

func (s *APIServer) httpStart(ctx context.Context) error {
	if !s.config.http.enabled {
		return nil
	}

	m := newMiddlewares(s.config.tracer, s.config.http.erc)
	s.registry.MustRegister(
		m.metrics.requestCount,
		m.metrics.responseCount,
		m.metrics.latency,
	)

	router := mux.NewRouter()
	router.Use(m.recoveryMiddleware)
	router.Use(m.headersMiddleware)

	apiRouter := router.PathPrefix("/api").Subrouter()
	s.httpConfigurationApi(m, apiRouter)

	err := s.httpConfigurationSwagger(router.PathPrefix("/swagger").Subrouter())
	if err != nil {
		return err
	}

	err = s.httpConfigurationGateway(ctx, router)
	if err != nil {
		return err
	}

	s.httpServer = &http.Server{
		Addr:         s.config.http.address(),
		WriteTimeout: s.config.http.writeTimeout,
		ReadTimeout:  s.config.http.readTimeout,
		IdleTimeout:  s.config.http.idleTimeout,
		Handler:      router,
	}

	go func() {
		s.logger.Info().Msg("Server http start")
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Err(err).Send()
		}
	}()

	return nil
}

func (s *APIServer) httpConfigurationApi(m *middlewares, router *mux.Router) {
	for _, a := range s.config.http.api {
		routeMap := a.RegistrationApi()
		for prefix, routes := range routeMap {
			sub := router.PathPrefix(prefix).Subrouter()

			for _, route := range routes {
				r := sub.Path(route.Pattern)
				path, _ := r.GetPathTemplate()

				handlerFunc := m.tracingMiddleware(route.HandlerFunc)

				if route.Metrics {
					handlerFunc = m.metricsMiddleware(path, handlerFunc)
				}

				handler := m.handleApiWrap(handlerFunc, route.ContentType, route.DataType)

				r.Handler(handler).Methods(route.Methods...)
			}
		}

		//routeMapWs := a.RegistrationWs()
		//for prefix, routes := range routeMapWs {
		//	sub := routerWs.PathPrefix(prefix).Subrouter()
		//
		//	for _, route := range routes {
		//		handler := m.HandleWsWrapper(route.HandlerFunc)
		//		sub.Handle(route.Pattern, handler).Methods(http.MethodGet)
		//	}
		//}
	}
}

func (s *APIServer) httpConfigurationSwagger(router *mux.Router) error {
	//если нет файла сваггера, сохраняем дефолтный что бы сваггер работал
	_, err := os.Open(s.config.http.swaggerFile)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile("swagger.json", []byte(swaggerDefault), 0644)
		if err != nil {
			return fmt.Errorf("failed to write default swagger.json: %w", err)
		}
	}

	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, s.config.http.swaggerFile)
	})

	router.PathPrefix("/").Handler(httpSwagger.Handler(
		httpSwagger.URL("swagger.json"),
		httpSwagger.BeforeScript(plugin),
		httpSwagger.Plugins([]string{"UrlMutatorPlugin"}),
		httpSwagger.UIConfig(map[string]string{
			"onComplete": fmt.Sprintf(`() => {window.ui.setBasePath('%s');}`, s.config.http.baseSwaggerPath),
		}),
	))

	return nil
}

func (s *APIServer) httpConfigurationGateway(ctx context.Context, router *mux.Router) error {
	if !s.config.grpc.enabled || !s.config.grpc.gatewayEnabled {
		return nil
	}

	runtimeMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	muxOptions := make([]runtime.ServeMuxOption, 0, len(s.config.grpc.muxOptions))
	muxOptions = append(muxOptions, s.config.grpc.muxOptions...)
	grpcAddress := fmt.Sprintf("localhost:%d", s.config.grpc.port)

	withErr := false
	for _, implementation := range s.config.grpc.implementations {
		err := implementation.RegisterHandlerFromEndpoint(ctx, runtimeMux, grpcAddress, opts)
		if err != nil {
			withErr = true
			s.logger.Err(err).Msg("failed to register gateway")
		}
	}

	if withErr {
		return FailedToRegisterGateway
	}

	router.PathPrefix("/").Handler(runtimeMux)
	return nil
}

func (s *APIServer) readinessCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, fmt.Sprintf(`{"live": %t}`, s.ready.Load()))
}

func (s *APIServer) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	errs := make([]error, 0, len(s.config.closers)+2)
	errs = append(errs, s.stop(ctx)...)

	for _, closeFunc := range s.config.closers {
		err := closeFunc(ctx)
		if err != nil {
			errs = append(errs, err)
			s.logger.Err(err).Send()
		}
	}

	s.logger.Info().Msg("Server stop")
	return errors.Join(errs...)
}

func (s *APIServer) stop(ctx context.Context) []error {
	errs := make([]error, 0, 2)

	s.ready.Store(false)
	if s.config.http.enabled {
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			errs = append(errs, err)
			s.logger.Err(err).Send()
		}
	}

	if s.config.grpc.enabled {
		s.grpcServer.GracefulStop()
	}

	if s.config.httpSystem.enabled {
		err := s.httpSystemServer.Shutdown(ctx)
		if err != nil {
			errs = append(errs, err)
			s.logger.Err(err).Send()
		}
	}

	return errs
}

func (s *APIServer) Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
