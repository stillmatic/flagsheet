package featuresheet

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/datadog/mmh3"
	"gopkg.in/Iwark/spreadsheet.v2"
)

type (
	FeatureValue string
	FlagName     string
)

type FeatureVariant struct {
	Value      FeatureValue
	Percentage int
}

type Feature struct {
	Key        string
	LayerName  string
	VariantMap map[string]FeatureVariant
}

// Layers encompass related and possibly interacting features.
// For example, if you are testing multiple changes to the signup screen,
// you should group them into a signup layer, as they can interact.
type Layer struct {
	Name    string
	Version int
	// buckets maps a particular bucket to the feature value.
	// This is an array of size 1000, where each index is a bucket
	// and the value is the feature value.
	// We use an array because it is significantly faster than a map.
	buckets []FeatureValue
	// cnt is an internal tracker of how many buckets have been filled.
	cnt int
}

// featureSheet is an internal representation for goroutine purposes.
type featureSheet struct {
	sheetID    string
	service    *spreadsheet.Service
	expiration time.Duration

	mu      sync.RWMutex
	janitor *janitor
	lmap    map[string]Layer
	fmap    map[string]Feature
}

type FeatureSheet struct {
	*featureSheet
}

// Evaluate returns the feature variant for a given flagName and id
// if the feature does not exist, it returns an empty string and false
func (f *featureSheet) Evaluate(key string, id *string) (FeatureValue, error) {
	feature, ok := f.fmap[key]
	if !ok {
		return "", fmt.Errorf("feature %s not found", key)
	}
	// get the layer -- this should not error
	layer, ok := f.lmap[feature.LayerName]
	if !ok {
		return "", fmt.Errorf("layer %s not found", feature.LayerName)
	}
	// get the bucket - essentially hash(id) % 100
	var bucket int

	// if id is nil, pick a random number
	if id == nil {
		bucket = rand.Intn(100)
	} else {
		// build hash input with bytes.Buffer
		// this should be very fast
		var bb bytes.Buffer
		bb.WriteString(*id)
		bb.WriteString("-")
		bb.WriteString(layer.Name)
		bb.WriteString("-")
		bb.WriteString(strconv.Itoa(layer.Version))
		h := mmh3.Hash32(bb.Bytes())
		bucket = int(h % 100)
	}
	// get the feature value
	fv := layer.buckets[bucket]
	return fv, nil
}

func (f *featureSheet) Refresh() error {
	// get spreadsheet
	spreadsheet, err := f.service.FetchSpreadsheet(f.sheetID)
	if err != nil {
		return fmt.Errorf("failed to fetch spreadsheet: %v", err)
	}
	featureMap := make(map[string]Feature)
	layerMap := make(map[string]Layer)

	// assume second sheet is layers
	layerRows := spreadsheet.Sheets[1].Rows
	for i, row := range layerRows {
		if i == 0 {
			continue
		}
		layerName := row[0].Value
		layerVersion, err := strconv.Atoi(row[1].Value)
		if err != nil {
			return fmt.Errorf("failed to parse layer version - must be int: %v", err)
		}
		layerMap[layerName] = Layer{
			Name:    layerName,
			Version: layerVersion,
			buckets: make([]FeatureValue, 1000),
		}
	}

	// assume first sheet is flags
	rows := spreadsheet.Sheets[0].Rows
	for i, row := range rows {
		if i == 0 {
			continue
		}
		featureKey := row[0].Value
		layerName := row[1].Value
		featureVariantKey := row[2].Value
		pct, err := strconv.Atoi(row[3].Value)
		if err != nil {
			return fmt.Errorf("failed to parse percentage - must be int: %v", err)
		}
		// get layer
		layer, ok := layerMap[layerName]
		if !ok {
			return fmt.Errorf("layer %s does not exist", layerName)
		}
		if pct+layer.cnt > 1000 {
			return fmt.Errorf("layer %s does not have enough buckets", layerName)
		}
		// add to layer
		for i := 0; i < pct; i++ {
			layer.buckets[layer.cnt] = FeatureValue(featureVariantKey)
			layer.cnt++
		}
		// add to feature map
		feature, ok := featureMap[featureKey]
		if !ok {
			feature = Feature{
				Key:        featureKey,
				LayerName:  layerName,
				VariantMap: make(map[string]FeatureVariant),
			}
		}
		feature.VariantMap[featureVariantKey] = FeatureVariant{
			Value:      FeatureValue(featureVariantKey),
			Percentage: pct,
		}
		featureMap[featureKey] = feature
		layerMap[layerName] = layer
	}
	// validate
	for _, layer := range layerMap {
		if layer.cnt > 1000 {
			return fmt.Errorf("layer %s has too many buckets", layer.Name)
		}
	}
	// lock and update
	f.mu.Lock()
	f.fmap = featureMap
	f.lmap = layerMap
	f.mu.Unlock()
	return nil
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *featureSheet) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			err := c.Refresh()
			if err != nil {
				log.Printf("failed to refresh feature sheet: %v", err)
			}
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor(c *FeatureSheet) {
	c.janitor.stop <- true
}

func runJanitor(c *featureSheet, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.Run(c)
}

func NewFeatureSheet(service *spreadsheet.Service, sheetID string, duration time.Duration) (*FeatureSheet, error) {
	fs := &featureSheet{
		sheetID:    sheetID,
		service:    service,
		expiration: duration,
	}
	if err := fs.Refresh(); err != nil {
		return nil, err
	}
	FS := &FeatureSheet{fs}
	if duration > 0 {
		runJanitor(fs, duration)
		runtime.SetFinalizer(FS, stopJanitor)
	}

	return FS, nil
}
