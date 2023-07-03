package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"

	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	"github.com/stillmatic/flagsheet"
	fsv1 "github.com/stillmatic/flagsheet/gen/flagsheet/v1"
	"github.com/stillmatic/flagsheet/gen/flagsheet/v1/flagsheetv1connect"
)

const (
	flagSheetVersionKey   = "FlagSheet-Version"
	flagSheetVersionValue = "v1"
)

type FlagSheetServer struct {
	fs *flagsheet.FlagSheet
}

func (s *FlagSheetServer) Evaluate(
	ctx context.Context,
	req *connect.Request[fsv1.EvaluateRequest],
) (*connect.Response[fsv1.EvaluateResponse], error) {
	fv, err := s.fs.Evaluate(req.Msg.Feature, &req.Msg.EntityId)
	if err != nil {
		return nil, connect.NewError(
			http.StatusNotFound,
			err,
		)
	}
	res := connect.NewResponse(&fsv1.EvaluateResponse{
		Variant: string(fv),
	})
	res.Header().Set(flagSheetVersionKey, flagSheetVersionValue)
	return res, nil
}

func ok(_ http.ResponseWriter, _ *http.Request) {}

func main() {
	// copy these from client_secret.json
	if os.Getenv("GCP_PROJECT_ID") == "" {
		panic("GCP_PROJECT_ID env var must be set")
	}
	serviceAccountJSON := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                  os.Getenv("GCP_PROJECT_ID"),
		"private_key_id":              os.Getenv("GCP_PRIVATE_KEY_ID"),
		"private_key":                 os.Getenv("GCP_PRIVATE_KEY"),
		"client_email":                os.Getenv("GCP_CLIENT_EMAIL"),
		"client_id":                   os.Getenv("GCP_CLIENT_ID"),
		"auth_uri":                    os.Getenv("GCP_AUTH_URI"),
		"token_uri":                   os.Getenv("GCP_TOKEN_URI"),
		"auth_provider_x509_cert_url": os.Getenv("GCP_AUTH_PROVIDER_CERT_URL"),
		"client_x509_cert_url":        os.Getenv("GCP_CLIENT_CERT_URL"),
	}
	serviceAccountJSONBytes, err := json.Marshal(serviceAccountJSON)
	if err != nil {
		panic(err)
	}
	conf, err := google.JWTConfigFromJSON(serviceAccountJSONBytes, spreadsheet.Scope)
	if err != nil {
		panic(err)
	}
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		panic("SPREADSHEET_ID env var must be set")
	}

	client := conf.Client(context.Background())
	service := spreadsheet.NewServiceWithClient(client)
	fs, err := flagsheet.NewFlagSheet(service, spreadsheetID, 10*time.Second)
	if err != nil {
		panic(err)
	}

	// serving
	s := &FlagSheetServer{
		fs: fs,
	}
	mux := http.NewServeMux()
	path, handler := flagsheetv1connect.NewFlagSheetServiceHandler(s)
	mux.Handle(path, handler)
	checker := grpchealth.NewStaticChecker(
		"flagsheet.v1.FlagSheetService",
	)
	mux.Handle(grpchealth.NewHandler(checker))
	mux.Handle("/health", http.HandlerFunc(ok))
	portNum := os.Getenv("PORT")
	if portNum == "" {
		portNum = "8080"
	}
	http.ListenAndServe(
		":"+portNum,
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
