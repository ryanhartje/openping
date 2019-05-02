package ping

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

// GetRequest is a helper function that sets up the request. This is broken off into a help to improve testability.
func GetRequest(url string) (request *http.Request, err error) {
	return http.NewRequest("GET", url, nil)
}

// GetDocument returns the HTML document to be stored in the document store for further analysis.
// This will be refactored to use channels
// Sets up HTTP trace, sets a timeout of 30 seconds, mimics a user agent,
func GetDocument(request *http.Request) (uptime Uptime, latency Latency, meta Metadata, size ContentSizes, err error) {
	var dns0, dns1, tls0, tls1, ttfb0, ttfb1, conn0, conn1 time.Time
	trace := &httptrace.ClientTrace{
		ConnectStart: func(string, string) {
			ttfb0 = time.Now()
		},
		GotFirstResponseByte: func() {
			ttfb1 = time.Now()
		},
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			dns0 = time.Now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dns1 = time.Now()
		},
		TLSHandshakeStart: func() {
			tls0 = time.Now()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			tls1 = time.Now()
		},
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	request.Header.Set("User-Agent", "MobileSafari/604.1 CFNetwork/978.0.7 Darwin/18.5.0")
	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	if _, err := http.DefaultTransport.RoundTrip(request); err != nil {
		return Uptime{time.Now(), 0, 0, request.RequestURI}, Latency{}, Metadata{}, ContentSizes{}, err
	}

	url := request.URL.Host
	conn0 = time.Now()
	response, err := client.Do(request)
	// Use this timestamp to populate our structs so they have a common timestamp.
	timestamp := time.Now().UTC()
	conn1 = time.Now()
	defer response.Body.Close()
	if err != nil {
		return Uptime{time.Now(), 0, response.StatusCode, url}, Latency{}, Metadata{}, ContentSizes{}, err
	}
	doc, _ := ioutil.ReadAll(response.Body)

	// Setup Uptime model
	uptime.RC = response.StatusCode
	if response.StatusCode == 200 {
		uptime.Up = 1 // true
	} else {
		uptime.Up = 0 //false
	}
	uptime.Timestamp = timestamp
	uptime.URL = url

	// Setup Metadata model
	meta.Document = string(doc)
	meta.Bytes = len(doc)
	meta.SHASum = fmt.Sprintf("%x", sha256.Sum256(doc))
	meta.Timestamp = timestamp
	meta.URL = url

	// Setup Latency and print off latency times if anything seems odd
	latency.DNSLookup = dns1.Sub(dns0)
	latency.TotalLatency = conn1.Sub(conn0)
	latency.TLSHandshake = tls1.Sub(tls0)
	latency.TTFB = ttfb1.Sub(ttfb0)
	latency.Timestamp = timestamp
	latency.URL = url

	// This is pretty hideous, open to suggestions / new ideas
	if !(latency.DNSLookup > (0 * time.Second)) ||
		!(latency.TotalLatency > (0 * time.Second)) ||
		!(latency.TTFB > (0 * time.Second)) {

		log.Printf("Possible lookup issue for URL: %v", request.URL)
		log.Printf("DNS Time info: %v", latency.DNSLookup)
		log.Printf("Latency: %v", latency.TotalLatency)
		log.Printf("TLS Handshake time: %v", latency.TLSHandshake)
		log.Printf("TTFB time: %v", latency.TTFB)
	}

	return uptime, latency, meta, ContentSizes{}, nil
}
