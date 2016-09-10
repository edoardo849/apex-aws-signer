package signer

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
)

// AWSSigningTransport implements http.RoundTripper. When set as Transport of http.Client, it signs HTTP requests with the latest AWS v4 signer and logs the requests
// No field is mandatory, but you can provide your own Transport or contextLogger by setting the Transport or Logger property. This library was inspired by https://github.com/motemen/go-loghttp/blob/master/loghttp.go
type AWSSigningTransport struct {
	Transport      http.RoundTripper
	Logger         *log.Entry
	awsServiceName string
	awsSigner      *v4.Signer
	awsRegion      string
}

// NewTransport creates a new instance of the AWSSigningTransport
func NewTransport(s *session.Session, serviceName string) *AWSSigningTransport {
	return &AWSSigningTransport{
		awsSigner:      v4.NewSigner(s.Config.Credentials),
		awsRegion:      *s.Config.Region,
		awsServiceName: serviceName,
	}
}

func readAndReplaceBody(req *http.Request) []byte {
	if req.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
// Signs the requests with AWS's v4 signer
func (t *AWSSigningTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	payload := bytes.NewReader(readAndReplaceBody(req))

	_, err := t.awsSigner.Sign(req, payload, t.awsServiceName, t.awsRegion, time.Now())
	if err != nil {
		t.logger().WithError(err).Error("Couldn't sign the request")
		return nil, err
	}

	t.logRequest(req)
	requestStarted := time.Now()
	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp, time.Since(requestStarted))

	return resp, err
}

func (t *AWSSigningTransport) logRequest(req *http.Request) {
	t.logger().Debugf("---> %s %s", req.Method, req.URL)
}

func (t *AWSSigningTransport) logResponse(resp *http.Response, d time.Duration) {
	t.logger().WithField("duration", d.String()).Debugf("<--- %d %s", resp.StatusCode, resp.Request.URL)
}

func (t *AWSSigningTransport) logger() *log.Entry {
	if t.Logger != nil {
		return t.Logger
	}
	return log.WithField("client", "AWSSigningTransport")
}

func (t *AWSSigningTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
