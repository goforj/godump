package godump

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDebugClient(t *testing.T, debug bool, transport http.RoundTripper) (*http.Client, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	if transport == nil {
		transport = http.DefaultTransport
	}
	tp := NewHTTPDebugTransport(transport)
	tp.Dumper().writer = &buf
	tp.SetDebug(debug)
	return &http.Client{Transport: tp}, &buf
}

func startTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func mustNewRequest(t *testing.T, method, reqURL string, body io.Reader) *http.Request {
	t.Helper()
	//nolint:noctx // no context needed for this unit test: synthetic request
	req, err := http.NewRequest(method, reqURL, body)
	require.NoError(t, err)
	return req
}

func mustDoRequest(t *testing.T, client *http.Client, req *http.Request) {
	t.Helper()
	resp, err := client.Do(req)
	require.NoError(t, err)
	if resp != nil {
		t.Cleanup(func() { resp.Body.Close() })
	}
}

var ErrSimulatedNetwork = errors.New("simulated network error")

func TestHTTPDebugTransport_WithDebugEnabled(t *testing.T) {
	client, buf := newDebugClient(t, true, nil)

	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "TestValue")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"success":true}`))
		assert.NoError(t, err, "failed to write response")
	})

	req := mustNewRequest(t, http.MethodGet, server.URL, http.NoBody)
	mustDoRequest(t, client, req)

	output := stripANSI(buf.String())
	t.Logf("Captured dump:\n%s", output)

	require.Contains(t, output, "Transaction =>")
	require.Contains(t, output, "Request =>")
	require.Contains(t, output, "Response =>")
	require.Contains(t, output, `"success":true`)
}

func TestHTTPDebugTransport_WithDebugDisabled(t *testing.T) {
	var buf bytes.Buffer

	tp := NewHTTPDebugTransport(http.DefaultTransport)
	tp.Dumper().writer = &buf

	client := &http.Client{Transport: tp}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	//nolint:noctx // no context needed for this unit test: synthetic request
	req, _ := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	output := stripANSI(buf.String())
	t.Logf("Captured dump:\n%s", output)

	require.NotContains(t, output, "Transaction =>", "did not expect 'Transaction =>' in dump when debug disabled")
}

func TestHTTPDebugTransport_RoundTripError(t *testing.T) {
	client, buf := newDebugClient(t, true, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, ErrSimulatedNetwork
	}))

	req := mustNewRequest(t, http.MethodGet, "http://example.invalid", http.NoBody)
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Expect simulated network error
	require.Error(t, err)
	require.ErrorContains(t, err, "simulated network error")

	// Ensure no response body to close
	require.Nil(t, resp)

	output := stripANSI(buf.String())
	t.Logf("Captured dump (error case):\n%s", output)

	require.NotContains(t, output, "Transaction =>", "did not expect Transaction block when RoundTrip failed")
}

func TestHTTPDebugTransport_SetDebugToggle(t *testing.T) {
	client, buf := newDebugClient(t, false, nil)

	server := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "DynamicTest")
		w.WriteHeader(http.StatusOK)
	})

	// Debug disabled
	mustDoRequest(t, client, mustNewRequest(t, http.MethodGet, server.URL, http.NoBody))
	output := stripANSI(buf.String())
	t.Logf("Dump with debug disabled:\n%s", output)
	require.NotContains(t, output, "Transaction =>", "should not log when debug disabled")

	// Enable debug
	tr, ok := client.Transport.(*HTTPDebugTransport)
	require.True(t, ok, "Transport should be HTTPDebugTransport")
	tr.SetDebug(true)
	buf.Reset()

	mustDoRequest(t, client, mustNewRequest(t, http.MethodGet, server.URL, http.NoBody))
	output = stripANSI(buf.String())
	t.Logf("Dump with debug enabled:\n%s", output)
	require.Contains(t, output, "Transaction =>", "should log when debug enabled")
}

// roundTripFunc lets us use a function as a RoundTripper in tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var ErrSimulatedTransportFailure = errors.New("simulated transport failure")

func TestHTTPDebugTransport_PassThroughRoundTripError(t *testing.T) {
	client, _ := newDebugClient(t, false, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, ErrSimulatedTransportFailure
	}))

	req := mustNewRequest(t, http.MethodGet, "http://example.invalid", http.NoBody)
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Assert error behavior
	require.Error(t, err)
	require.ErrorContains(t, err, "HTTPDebugTransport: pass-through round trip failed")
	require.ErrorIs(t, err, ErrSimulatedTransportFailure)

	// Response should be nil
	require.Nil(t, resp)
}

func TestHTTPDebugTransport_RequestDumpFailure(t *testing.T) {
	client, _ := newDebugClient(t, true, nil)

	// Malformed request: URL exists but has no Scheme/Host
	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{}, // No scheme/host triggers DumpRequestOut failure
		Header: http.Header{},
	}

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Response should be nil
	require.Nil(t, resp)

	// Assert error behavior
	require.Error(t, err)
	require.ErrorContains(t, err, "HTTPDebugTransport: failed to dump request")
}

type errorBody struct{}

// errSimulatedBodyReadFailure simulates a failure when reading the response body.
var errSimulatedBodyReadFailure = errors.New("simulated body read failure")

func (errorBody) Read(p []byte) (int, error) { return 0, errSimulatedBodyReadFailure }
func (errorBody) Close() error               { return nil }

func TestHTTPDebugTransport_ResponseDumpFailure(t *testing.T) {
	client, _ := newDebugClient(t, true, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(errorBody{}), // Simulates body read failure
		}, nil
	}))

	req := mustNewRequest(t, http.MethodGet, "http://example.invalid", http.NoBody)
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Assert response dump failure
	require.Error(t, err)
	require.ErrorContains(t, err, "HTTPDebugTransport: failed to dump response")
	require.ErrorIs(t, err, errSimulatedBodyReadFailure)

	// Response should be nil because dump failed
	require.Nil(t, resp)
}
