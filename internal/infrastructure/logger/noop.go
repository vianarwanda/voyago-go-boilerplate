package logger

import "context"

type NoOpLogger struct{}

var _ Logger = (*NoOpLogger)(nil)

func NewNoOpLogger() Logger { return &NoOpLogger{} }

func (l *NoOpLogger) WithContext(ctx context.Context) Logger  { return l }
func (l *NoOpLogger) WithField(key string, value any) Logger  { return l }
func (l *NoOpLogger) WithFields(fields map[string]any) Logger { return l }

func (l *NoOpLogger) Debug(message string) {}
func (l *NoOpLogger) Info(message string)  {}
func (l *NoOpLogger) Warn(message string)  {}
func (l *NoOpLogger) Error(message string) {}
