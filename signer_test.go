package signer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/edoardo849/apex-aws-signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {

	handler := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("200 OK"))
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	transport := signer.NewTransport(session.New(&aws.Config{Region: aws.String("eu-west-1")}), "")

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", ts.URL, nil)
	req.Header.Add("my-test-header", "my-test-value")
	resp, err := client.Do(req)

	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	header := resp.Request.Header

	// checks if AWS's Headers are set
	require.NotEmpty(t, header.Get("X-Amz-Date"))
	require.NotEmpty(t, header.Get("Authorization"))

	// checks if it preserves the original headers
	require.NotEmpty(t, header.Get("my-test-header"))
}
