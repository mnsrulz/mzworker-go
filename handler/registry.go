package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/mehdihadeli/go-mediatr"
)

type RequestFactory func(payload []byte) (any, error)

func jsonRequestFactory[T any](payload []byte) (any, error) {
	var request T
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, err
	}
	return &request, nil
}

type Deps struct {
	Executor QueryExecutor
}

type registryEntry struct {
	factory  RequestFactory
	dispatch func(ctx context.Context, req any) (any, error)
}

var (
	mu           sync.RWMutex
	registry     = map[string]registryEntry{}
	typeRegistry = map[reflect.Type]func(ctx context.Context, req any) (any, error){}
	pending      []func(Deps) error
)

// RegisterStruct registers a request type with its JSON factory and mediatr handler.
// T is the struct type (e.g., DynamicSQLQuery), TResp is response type.
// handlerFactory receives Deps (Executor injected only where needed).
func RegisterStruct[T any, TResp any](requestType string, handlerFactory func(Deps) mediatr.RequestHandler[*T, TResp]) {
	factory := jsonRequestFactory[T]
	mu.Lock()
	existing, ok := registry[requestType]
	if !ok {
		registry[requestType] = registryEntry{factory: factory}
	} else {
		existing.factory = factory
		registry[requestType] = existing
	}
	mu.Unlock()
	pending = append(pending, func(d Deps) error {
		h := handlerFactory(d)
		if err := mediatr.RegisterRequestHandler(h); err != nil {
			return err
		}
		dispatch := func(ctx context.Context, req any) (any, error) {
			return mediatr.Send[*T, TResp](ctx, req.(*T))
		}
		mu.Lock()
		e := registry[requestType]
		e.dispatch = dispatch
		registry[requestType] = e
		typeRegistry[reflect.TypeFor[*T]()] = dispatch
		mu.Unlock()
		return nil
	})
}

func GetRequestFactory(requestType string) (RequestFactory, bool) {
	mu.RLock()
	e, ok := registry[requestType]
	mu.RUnlock()
	if !ok || e.factory == nil {
		return nil, false
	}
	return e.factory, true
}

func Dispatch(ctx context.Context, req any) (any, error) {
	mu.RLock()
	fn, ok := typeRegistry[reflect.TypeOf(req)]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no dispatch registered for %T", req)
	}
	return fn(ctx, req)
}

func DispatchByType(ctx context.Context, requestType string, payload []byte) (any, error) {
	mu.RLock()
	e, ok := registry[requestType]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown request type: %s", requestType)
	}
	req, err := e.factory(payload)
	if err != nil {
		return nil, err
	}
	return e.dispatch(ctx, req)
}

func initRegistry(d Deps) error {
	for _, fn := range pending {
		if err := fn(d); err != nil {
			return err
		}
	}
	return nil
}
