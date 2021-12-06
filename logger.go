package consul

import "google.golang.org/grpc/grpclog"

// Logger is used to log any errors during async message processing.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type noopLogger struct{}

// Errorf does nothing.
func (n noopLogger) Errorf(_ string, _ ...interface{}) {
}

// Infof does nothing.
func (n noopLogger) Infof(_ string, _ ...interface{}) {
}

// ensure Logger method set is a subset of grpclog interface.
var _ Logger = grpclog.LoggerV2(nil)

// grpcGlobalLogger is a wrapper around grpclog package methods,
// introduced because grpclog doesn't export Logger instance.
type grpcGlobalLogger struct{}

func (g grpcGlobalLogger) Errorf(format string, args ...interface{}) {
	grpclog.Errorf(format, args...)
}

func (g grpcGlobalLogger) Infof(format string, args ...interface{}) {
	grpclog.Infof(format, args...)
}
