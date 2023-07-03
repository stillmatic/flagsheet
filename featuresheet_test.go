package flagsheet_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stillmatic/flagsheet"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
)

func stringPtr(s string) *string {
	return &s
}

const testSpreadsheetID = "15_oV5NcvYK7wK3VVD5ol6KVkWHzPLFl22c1QyLYplpU"

func TestSheet(t *testing.T) {
	data, err := os.ReadFile("client_secret.json")
	assert.NoError(t, err)

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	assert.NoError(t, err)

	client := conf.Client(context.TODO())
	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := flagsheet.NewFlagSheet(service, testSpreadsheetID, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, spreadsheet)
	fv, err := spreadsheet.Evaluate("my_key", stringPtr("my_id"))
	assert.Nil(t, err)
	assert.NotEmpty(t, fv)
	assert.Equal(t, "foo", string(fv))
}

func BenchmarkEvaluate(b *testing.B) {
	data, err := os.ReadFile("client_secret.json")
	assert.NoError(b, err)

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	assert.NoError(b, err)

	client := conf.Client(context.TODO())
	service := spreadsheet.NewServiceWithClient(client)
	spreadsheet, err := flagsheet.NewFlagSheet(service, testSpreadsheetID, 1*time.Second)
	assert.NoError(b, err)
	assert.NotNil(b, spreadsheet)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fv, err := spreadsheet.Evaluate("my_key", stringPtr("my_id"))
		assert.Nil(b, err)
		assert.NotEmpty(b, fv)
		assert.Equal(b, "foo", string(fv))
	}
}
