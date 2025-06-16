package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime/debug"

	"github.com/dgraph-io/ristretto/v2"
	grpcLogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/necroskillz/config-service/db"
	pb "github.com/necroskillz/config-service/grpc/gen"
	"github.com/necroskillz/config-service/services"
	"github.com/necroskillz/config-service/util/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpcServer *grpc.Server
	dbpool     *pgxpool.Pool
	cache      *ristretto.Cache[string, any]
}

func NewServer() *Server {
	return &Server{}
}

func interceptorLogger(l *slog.Logger) grpcLogging.Logger {
	return grpcLogging.LoggerFunc(func(ctx context.Context, lvl grpcLogging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (s *Server) Start() error {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	connectionString := os.Getenv("DATABASE_URL")

	if err := db.Migrate(connectionString); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.dbpool = dbpool
	s.cache = cache

	svc := services.InitializeServices(dbpool, cache)

	logger := logging.ConfigureSlog("Config Service gRPC/server")

	panicRecoveryHandler := func(p any) (err error) {
		logger.Error("recovered from panic", "panic", p, "stack", string(debug.Stack()))

		return status.Errorf(codes.Internal, "%s", p)
	}

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcLogging.UnaryServerInterceptor(interceptorLogger(logger)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(panicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			grpcLogging.StreamServerInterceptor(interceptorLogger(logger)),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(panicRecoveryHandler)),
		),
	)
	pb.RegisterConfigServiceServer(s.grpcServer, NewConfigurationServer(svc))

	slog.Info("Starting gRPC server on port", "port", os.Getenv("GRPC_PORT"))

	return s.grpcServer.Serve(listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()

	if s.dbpool != nil {
		s.dbpool.Close()
	}

	if s.cache != nil {
		s.cache.Close()
	}

	return nil
}
