// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: flagsheet/v1/flagsheet.proto

package flagsheetv1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	v1 "github.com/stillmatic/flagsheet/gen/flagsheet/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion0_1_0

const (
	// FlagSheetServiceName is the fully-qualified name of the FlagSheetService service.
	FlagSheetServiceName = "flagsheet.v1.FlagSheetService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// FlagSheetServiceEvaluateProcedure is the fully-qualified name of the FlagSheetService's Evaluate
	// RPC.
	FlagSheetServiceEvaluateProcedure = "/flagsheet.v1.FlagSheetService/Evaluate"
)

// FlagSheetServiceClient is a client for the flagsheet.v1.FlagSheetService service.
type FlagSheetServiceClient interface {
	Evaluate(context.Context, *connect_go.Request[v1.EvaluateRequest]) (*connect_go.Response[v1.EvaluateResponse], error)
}

// NewFlagSheetServiceClient constructs a client for the flagsheet.v1.FlagSheetService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewFlagSheetServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) FlagSheetServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &flagSheetServiceClient{
		evaluate: connect_go.NewClient[v1.EvaluateRequest, v1.EvaluateResponse](
			httpClient,
			baseURL+FlagSheetServiceEvaluateProcedure,
			opts...,
		),
	}
}

// flagSheetServiceClient implements FlagSheetServiceClient.
type flagSheetServiceClient struct {
	evaluate *connect_go.Client[v1.EvaluateRequest, v1.EvaluateResponse]
}

// Evaluate calls flagsheet.v1.FlagSheetService.Evaluate.
func (c *flagSheetServiceClient) Evaluate(ctx context.Context, req *connect_go.Request[v1.EvaluateRequest]) (*connect_go.Response[v1.EvaluateResponse], error) {
	return c.evaluate.CallUnary(ctx, req)
}

// FlagSheetServiceHandler is an implementation of the flagsheet.v1.FlagSheetService service.
type FlagSheetServiceHandler interface {
	Evaluate(context.Context, *connect_go.Request[v1.EvaluateRequest]) (*connect_go.Response[v1.EvaluateResponse], error)
}

// NewFlagSheetServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewFlagSheetServiceHandler(svc FlagSheetServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(FlagSheetServiceEvaluateProcedure, connect_go.NewUnaryHandler(
		FlagSheetServiceEvaluateProcedure,
		svc.Evaluate,
		opts...,
	))
	return "/flagsheet.v1.FlagSheetService/", mux
}

// UnimplementedFlagSheetServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedFlagSheetServiceHandler struct{}

func (UnimplementedFlagSheetServiceHandler) Evaluate(context.Context, *connect_go.Request[v1.EvaluateRequest]) (*connect_go.Response[v1.EvaluateResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("flagsheet.v1.FlagSheetService.Evaluate is not implemented"))
}