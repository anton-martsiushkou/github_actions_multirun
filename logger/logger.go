package logger

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.DebugLevel)
	return l
}

type contextKey struct{}

func ToContext(ctx context.Context, l *logrus.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func FromContext(ctx context.Context) *logrus.Logger {
	l := ctx.Value(contextKey{})
	if l == nil {
		return New()
	}

	return l.(*logrus.Logger)
}
