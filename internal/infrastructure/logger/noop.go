package logger

import "context"

type noOpLogger struct{}

var _ Logger = (*noOpLogger)(nil)

func NewNoOpLogger() Logger { return &noOpLogger{} }

func (l *noOpLogger) WithContext(ctx context.Context) Logger  { return l }
func (l *noOpLogger) WithField(key string, value any) Logger  { return l }
func (l *noOpLogger) WithFields(fields map[string]any) Logger { return l }

func (l *noOpLogger) Debug(message string) {}
func (l *noOpLogger) Info(message string)  {}
func (l *noOpLogger) Warn(message string)  {}
func (l *noOpLogger) Error(message string) {}
