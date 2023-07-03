package main

import (
	"context"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	flagsheetv1 "github.com/stillmatic/flagsheet/gen/flagsheet/v1"
	"github.com/stillmatic/flagsheet/gen/flagsheet/v1/flagsheetv1connect"
)

func main() {
	client := flagsheetv1connect.NewFlagSheetServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	res, err := client.Evaluate(
		context.Background(),
		connect.NewRequest(&flagsheetv1.EvaluateRequest{
			Feature:  "my_key",
			EntityId: "my_id",
		}),
	)
	if err != nil {
		panic(err)
	}
	log.Println(res.Msg.Variant)
}
