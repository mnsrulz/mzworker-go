package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/mnsrulz/mzworker-go/gen/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedOptionsQueryServiceServer
	dataDir string
}

func NewServer(dataDir string) *server {
	return &server{dataDir: dataDir}
}

func (s *server) ExecuteQuery(ctx context.Context, req *pb.ExecuteQueryRequest) (*pb.ExecuteQueryResponse, error) {
	log.Printf("Received query: %s (limit: %d)", req.Query, req.Limit)

	limit := enforceLimit(req.Limit)
	columns, rows, err := executeQuery(ctx, s.dataDir, req.Query, limit)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	pbRows := make([]*pb.Row, len(rows))
	for i, row := range rows {
		values := make([]*pb.Value, len(row))
		for j, v := range row {
			values[j] = toProtoValue(v)
		}
		pbRows[i] = &pb.Row{Values: values}
	}

	log.Printf("Query completed: %d rows", len(rows))
	return &pb.ExecuteQueryResponse{
		Columns: columns,
		Rows:    pbRows,
	}, nil
}

func toProtoValue(v interface{}) *pb.Value {
	if v == nil {
		return &pb.Value{Value: nil}
	}

	switch val := v.(type) {
	case string:
		return &pb.Value{Value: &pb.Value_StringVal{StringVal: val}}
	case float64:
		return &pb.Value{Value: &pb.Value_DoubleVal{DoubleVal: val}}
	case int64:
		return &pb.Value{Value: &pb.Value_IntVal{IntVal: val}}
	case bool:
		return &pb.Value{Value: &pb.Value_BoolVal{BoolVal: val}}
	default:
		return &pb.Value{Value: &pb.Value_StringVal{StringVal: fmt.Sprintf("%v", val)}}
	}
}
