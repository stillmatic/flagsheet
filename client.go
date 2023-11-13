package flagsheet

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Yiling-J/theine-go"
	"github.com/bufbuild/connect-go"
	flagsheetv1 "github.com/stillmatic/flagsheet/gen/flagsheet/v1"
	"github.com/stillmatic/flagsheet/gen/flagsheet/v1/flagsheetv1connect"
)

type flagQuery struct {
	Feature  string
	EntityID string
}

type FlagClient struct {
	// flags is the feature flags client
	flags flagsheetv1connect.FlagSheetServiceClient
	// cache stores key value pairs with their result
	cache    *theine.Cache[flagQuery, string]
	duration time.Duration
}

func NewFlagClient() *FlagClient {
	flagsURL := os.Getenv("flagsheet_URL")
	if flagsURL == "" {
		panic("flagsheet_URL env var must be set")
	}
	flagsClient := flagsheetv1connect.NewFlagSheetServiceClient(http.DefaultClient, flagsURL)
	cache, err := theine.NewBuilder[flagQuery, string](1024).Build()
	if err != nil {
		panic(err)
	}
	return &FlagClient{
		flags:    flagsClient,
		cache:    cache,
		duration: 10 * time.Second,
	}
}

func (f *FlagClient) Evaluate(ctx context.Context, feature string, entityID string) (string, error) {
	query := flagQuery{
		Feature:  feature,
		EntityID: entityID,
	}
	val, ok := f.cache.Get(query)
	// cache hit
	if ok {
		return val, nil
	}
	// cache miss, call and set cache
	req := connect.NewRequest(&flagsheetv1.EvaluateRequest{
		Feature:  feature,
		EntityId: entityID,
	})
	res, err := f.flags.Evaluate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("could not evaluate feature: %w", err)
	}
	f.cache.SetWithTTL(query, res.Msg.Variant, 1, f.duration)
	return res.Msg.Variant, nil
}
