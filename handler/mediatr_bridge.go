package handler

import "context"

// Dispatch routes an already-typed request through the mediatr pipeline.
// The request must be a pointer type registered via RegisterStruct (e.g., *DynamicSQLQuery).
func DispatchBridge(ctx context.Context, req any) (any, error) {
	return Dispatch(ctx, req)
}

// DispatchByTypeBridge creates a typed request from payload via the registry factory
// and dispatches it through the pipeline. Convenience for AMQP envelope handling.
func DispatchByTypeBridge(ctx context.Context, requestType string, payload []byte) (any, error) {
	return DispatchByType(ctx, requestType, payload)
}
