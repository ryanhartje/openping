package ping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	status := Status{
		timestamp: time.Now(),
		state:     "OK",
	}
	assert.Equal(t, status.state, "OK")
}

func TestDocumentMemoryStore(t *testing.T) {
	dms := NewDocumentStore()
	url := "lol://test"
	testDoc := `
	<html><head><title>Test</title></head><body><h1>OMG Great Test!</h1></body></html>
	`
	uptime := Uptime{time.Now(), true, 200, "localhost"}
	meta := Metadata{Document: testDoc, URL: url}
	dms.Update(uptime, Latency{}, meta, ContentSizes{})
	dms.Update(uptime, Latency{}, meta, ContentSizes{})

	assert.Equal(
		t,
		len(dms.Metas),
		2,
		"Error, inserted 2 documents but have a document store size of %v",
		len(dms.Metas),
	)
	assert.Equal(
		t,
		url,
		meta.URL,
		ContentSizes{},
	)
}
