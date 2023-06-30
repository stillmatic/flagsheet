package main

import (
	"context"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	featuresheetv1 "github.com/stillmatic/featuresheet/gen/featuresheet/v1"
	"github.com/stillmatic/featuresheet/gen/featuresheet/v1/featuresheetv1connect"
)

func main() {
	client := featuresheetv1connect.NewFeatureSheetServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	res, err := client.Evaluate(
		context.Background(),
		connect.NewRequest(&featuresheetv1.EvaluateRequest{
			Feature:  "my_key",
			EntityId: "my_id",
		}),
	)
	if err != nil {
		panic(err)
	}
	log.Println(res.Msg.Variant)
}
