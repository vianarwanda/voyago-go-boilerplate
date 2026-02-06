package ctxkey

import "context"

type key struct{}

var (
	kTx        = key{}
	kRequestID = key{}
)

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(kRequestID).(string); ok {
		return id
	}
	return ""
}

func GetTransaction(ctx context.Context) any {
	if ctx == nil {
		return nil
	}
	return ctx.Value(kTx)
}

func SetTransaction(ctx context.Context, tx any) context.Context {
	return context.WithValue(ctx, kTx, tx)
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, kRequestID, id)
}
