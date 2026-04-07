package logging

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	ServiceName    = "simulador"
	ServiceVersion = "0.1.0"
)

// Setup inicializa logging estruturado com OpenTelemetry.
// Retorna o logger e uma funcao de cleanup para flush do provider.
func Setup(verbose, quiet bool) (*slog.Logger, func()) {
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName),
			semconv.ServiceVersion(ServiceVersion),
		),
	)

	var level slog.Level
	switch {
	case verbose:
		level = slog.LevelDebug
	case quiet:
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
	)

	otelHandler := otelslog.NewHandler(ServiceName,
		otelslog.WithLoggerProvider(provider),
	)

	jsonHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(newMultiHandler(jsonHandler, otelHandler))

	cleanup := func() {
		_ = provider.Shutdown(context.Background())
	}

	return logger, cleanup
}

// multiHandler distribui log records para multiplos handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return newMultiHandler(handlers...)
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return newMultiHandler(handlers...)
}
