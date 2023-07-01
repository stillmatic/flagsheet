package featuresheet

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Yiling-J/theine-go"
	"github.com/bufbuild/connect-go"
	featuresheetv1 "github.com/stillmatic/featuresheet/gen/featuresheet/v1"
	"github.com/stillmatic/featuresheet/gen/featuresheet/v1/featuresheetv1connect"
)

type flagQuery struct {
	Feature  string
	EntityID string
}

type FlagClient struct {
	// flags is the feature flags client
	flags featuresheetv1connect.FeatureSheetServiceClient
	// cache stores key value pairs with their result
	cache    *theine.Cache[flagQuery, string]
	duration time.Duration
}

func NewFlagClient() *FlagClient {
	flagsURL := os.Getenv("FEATURESHEET_URL")
	if flagsURL == "" {
		panic("FEATURESHEET_URL env var must be set")
	}
	flagsClient := featuresheetv1connect.NewFeatureSheetServiceClient(http.DefaultClient, flagsURL)
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
	req := connect.NewRequest(&featuresheetv1.EvaluateRequest{
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
