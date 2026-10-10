package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"github.com/ikondratev/event-service/internal/eventpb"
	"github.com/ikondratev/event-service/internal/settings"
	"google.golang.org/grpc"
)

type Listener struct {
	engine   *grpc.Server
	settings *settings.Settings
	log      *slog.Logger
}

func NewListener(settings *settings.Settings, log *slog.Logger, handler *Server) *Listener {
	engine := grpc.NewServer()
	eventpb.RegisterEventServiceServer(engine, handler)
	return &Listener{
		engine:   engine,
		settings: settings,
		log:      log,
	}
}

func (l *Listener) Start(errCh chan<- error) {
	lis, err := net.Listen("tcp", l.settings.Upstream.Port)
	if err != nil {
		errCh <- err
		return
	}
	l.log.Info("grpc started", "addr", l.settings.Upstream.Port)
	errCh <- l.engine.Serve(lis)
}

func (l *Listener) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		l.engine.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		l.engine.Stop()
		return ctx.Err()
	}
}

func IsStopped(err error) bool {
	return errors.Is(err, grpc.ErrServerStopped)
}
