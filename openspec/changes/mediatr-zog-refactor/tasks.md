## 1. Dependencies

- [x] 1.1 Add `github.com/mehdihadeli/go-mediatr` via `go get`
- [x] 1.2 Add `github.com/Oudwins/zog` via `go get`

## 2. Handler Package — Types and Schemas

- [x] 2.1 Create `handler/` directory
- [x] 2.2 Create `handler/types.go` with `DynamicSQLQuery` request struct (fields: `Query string`, `Limit int32`) and `QueryResponse` struct (fields: `Columns []string`, `Rows [][]any`)
- [x] 2.3 Define Zog schemas in `handler/types.go` — `DynamicSQLQuerySchema` with `z.Struct(z.Shape{...})` where keys match struct field names (PascalCase)

## 3. Handler Package — Factory Registry

- [x] 3.1 Create `handler/registry.go` with `RequestFactory` type (`func(payload []byte) (any, error)`)
- [x] 3.2 Implement `GetRequestFactory(requestType string) (RequestFactory, bool)` lookup function
- [x] 3.3 Register `"dynamic-sql-query"` factory that unmarshals JSON into `*DynamicSQLQuery`

## 4. Handler Package — Query Handler

- [x] 4.1 Create `handler/query_handler.go` with `QueryHandler` struct holding `dataDir string`
- [x] 4.2 Implement `NewQueryHandler(dataDir string) *QueryHandler` constructor
- [x] 4.3 Implement `Handle(ctx context.Context, req *DynamicSQLQuery) (*QueryResponse, error)` — validate with Zog, then call `executeQuery`, return `QueryResponse`

## 5. Handler Package — Pipeline Behavior

- [x] 5.1 Create `handler/validation.go` with `ValidationBehavior` struct
- [x] 5.2 Implement `Handle(ctx context.Context, request any, next mediatr.RequestHandlerFunc) (any, error)` — log request type, call `next()`, return result

## 6. AMQP Consumer — Mediatr Integration

- [x] 6.1 Update `amqp.go` `handleMessage` to extract `requestType` from JSON envelope
- [x] 6.2 Use `handler.GetRequestFactory` to create typed request from payload
- [x] 6.3 Add `dispatchMediatr(ctx, request)` function with type switch calling `mediatr.Send` for each request type
- [x] 6.4 Remove direct `executeQuery` call and inline `amqpRequest`/`amqpResponse` types

## 7. gRPC Server — Mediatr Integration

- [x] 7.1 Update `server.go` `ExecuteQuery` to create `*handler.DynamicSQLQuery` from gRPC request
- [x] 7.2 Call `mediatr.Send[*handler.DynamicSQLQuery, *handler.QueryResponse]` instead of direct `executeQuery`
- [x] 7.3 Convert `QueryResponse` to protobuf `ExecuteQueryResponse`

## 8. Main — Registration

- [x] 8.1 Import `handler` package and `go-mediatr`
- [x] 8.2 Register `ValidationBehavior` pipeline via `mediatr.RegisterRequestPipelineBehaviors`
- [x] 8.3 Register `QueryHandler` via `mediatr.RegisterRequestHandler[*handler.DynamicSQLQuery, *handler.QueryResponse]`

## 9. Verification

- [x] 9.1 Run `go build ./...` to verify compilation
- [x] 9.2 Run `go vet ./...` and existing tests to verify no regressions
