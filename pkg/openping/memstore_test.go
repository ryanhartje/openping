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
	dms.Update(url, testDoc)
	dms.Update(url, testDoc)

	assert.Equal(
		t,
		len(dms.Documents[url]),
		2,
		"Error, inserted 2 documents but have a document store size of %v",
		len(dms.Documents[url]),
	)
}