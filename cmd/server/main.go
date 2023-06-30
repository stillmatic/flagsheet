package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"

	"github.com/stillmatic/featuresheet"
	fsv1 "github.com/stillmatic/featuresheet/gen/featuresheet/v1"
	"github.com/stillmatic/featuresheet/gen/featuresheet/v1/featuresheetv1connect"
)

type FeatureSheetServer struct {
	fs *featuresheet.FeatureSheet
}

func (s *FeatureSheetServer) Evaluate(
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
	res.Header().Set("FeatureSheet-Version", "v1")
	return res, nil
}

func main() {
	data, err := os.ReadFile("client_secret.json")
	if err != nil {
		panic(err)
	}

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	if err != nil {
		panic(err)
	}
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		panic("SPREADSHEET_ID env var must be set")
	}

	client := conf.Client(context.TODO())
	service := spreadsheet.NewServiceWithClient(client)
	fs, err := featuresheet.NewFeatureSheet(service, spreadsheetID, 10*time.Second)
	if err != nil {
		panic(err)
	}
	s := &FeatureSheetServer{
		fs: fs,
	}
	mux := http.NewServeMux()
	path, handler := featuresheetv1connect.NewFeatureSheetServiceHandler(s)
	mux.Handle(path, handler)
	portNum := os.Getenv("PORT")
	if portNum == "" {
		portNum = "8080"
	}
	http.ListenAndServe(
		":"+portNum,
		h2c.NewHandler(mux, &http2.Server{}),
	)
}
