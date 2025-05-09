package httplambda

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest(t *testing.T) {
	testCases := []struct {
		name  string
		event events.LambdaFunctionURLRequest
		want  *http.Request
	}{
		{
			name: "simple GET request",
			event: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
					},
				},
				Headers: map[string]string{
					"Host": "example.com",
				},
				RawPath: "/test-path",
			},
			want: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test-path",
				},
				Header: http.Header{
					"Host": []string{"example.com"},
				},
				Host: "example.com",
			},
		},

		{
			name: "simple POST request with body",
			event: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Headers: map[string]string{
					"Host":         "example.com",
					"Content-Type": "application/json",
				},
				RawPath: "/test-path",
				Body:    `{"key":"value"}`,
			},
			want: &http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test-path",
				},
				Header: http.Header{
					"Host":         []string{"example.com"},
					"Content-Type": []string{"application/json"},
				},
				Host: "example.com",
				Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
		},
		{
			name: "POST request with cookies",
			event: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Headers: map[string]string{
					"Host":         "example.com",
					"Content-Type": "application/json",
					"Cookie":       "session_id=abc123; user_id=42",
				},
				RawPath: "/test-path",
				Body:    `{"key":"value"}`,
			},
			want: &http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test-path",
				},
				Header: http.Header{
					"Host":         []string{"example.com"},
					"Content-Type": []string{"application/json"},
					"Cookie":       []string{"session_id=abc123; user_id=42"},
				},
				Host: "example.com",
				Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
		},
		{
			name: "POST request with base64-encoded body",
			event: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Headers: map[string]string{
					"Host":         "example.com",
					"Content-Type": "application/json",
				},
				RawPath:         "/test-path",
				Body:            "eyJrZXkiOiJ2YWx1ZSJ9", // Base64-encoded string of {"key":"value"}
				IsBase64Encoded: true,
			},
			want: &http.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test-path",
				},
				Header: http.Header{
					"Host":         []string{"example.com"},
					"Content-Type": []string{"application/json"},
				},
				Host: "example.com",
				Body: io.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httpRequest(context.Background(), tt.event)
			require.NoError(t, err)

			assert.Equal(t, tt.want.Method, got.Method)
			assert.Equal(t, tt.want.URL.String(), got.URL.String())
			assert.Equal(t, tt.want.Header, got.Header)
			assert.Equal(t, tt.want.Host, got.Host)

			var wantBody []byte
			if tt.want.Body != nil {
				wantBody, _ = io.ReadAll(tt.want.Body)
			}

			var gotBody []byte
			if got.Body != nil {
				gotBody, _ = io.ReadAll(got.Body)
			}

			assert.True(t, bytes.Equal(wantBody, gotBody))
		})
	}
}

func TestHTTPResponseEvent(t *testing.T) {
	testCases := []struct {
		name string
		w    *httptest.ResponseRecorder
		want events.LambdaFunctionURLResponse
	}{
		{
			name: "empty 200 OK",
			w: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusOK)
				return rec
			}(),
			want: events.LambdaFunctionURLResponse{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{},
				Body:       "",
			},
		},
		{
			name: "200 OK with response body",
			w: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusOK)
				_, _ = rec.Write([]byte(`{"message":"success"}`)) // Writing response body
				return rec
			}(),
			want: events.LambdaFunctionURLResponse{
				StatusCode:      http.StatusOK,
				Headers:         map[string]string{},
				Body:            `eyJtZXNzYWdlIjoic3VjY2VzcyJ9`, // Expected response body
				IsBase64Encoded: true,
			},
		},
		{
			name: "404 Not Found",
			w: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusNotFound)
				_, _ = rec.Write([]byte(`{"error":"resource not found"}`)) // Writing response body
				return rec
			}(),
			want: events.LambdaFunctionURLResponse{
				StatusCode:      http.StatusNotFound,
				Headers:         map[string]string{},
				Body:            `eyJlcnJvciI6InJlc291cmNlIG5vdCBmb3VuZCJ9`, // Expected response body
				IsBase64Encoded: true,
			},
		},
		{
			name: "200 OK with response body, repeated headers, and repeated Set-Cookie",
			w: func() *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				rec.Header().Add("X-Custom-Header", "value1")
				rec.Header().Add("X-Custom-Header", "value2")
				rec.Header().Add("Set-Cookie", "session_id=abc123; Path=/; HttpOnly")
				rec.Header().Add("Set-Cookie", "user_id=42; Path=/; HttpOnly")
				rec.WriteHeader(http.StatusOK)
				_, _ = rec.Write([]byte(`{"message":"success"}`)) // Writing response body
				return rec
			}(),
			want: events.LambdaFunctionURLResponse{
				StatusCode: http.StatusOK,
				Cookies: []string{
					"session_id=abc123; Path=/; HttpOnly",
					"user_id=42; Path=/; HttpOnly",
				},
				Headers: map[string]string{
					"X-Custom-Header": "value1,value2",
				},
				Body:            `eyJtZXNzYWdlIjoic3VjY2VzcyJ9`, // Expected response body
				IsBase64Encoded: true,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got := httpResponseEvent(tt.w)

			assert.Equal(t, tt.want.StatusCode, got.StatusCode)
			assert.Equal(t, tt.want.Cookies, got.Cookies)

			if len(tt.want.Headers) != 0 || len(got.Headers) != 0 {
				assert.Equal(t, tt.want.Headers, got.Headers)
			}

			var wantBody, gotBody []byte
			if tt.want.Body != "" {
				wantBody = []byte(tt.want.Body)
			}
			if got.Body != "" {
				gotBody = []byte(got.Body)
			}

			assert.True(t, bytes.Equal(wantBody, gotBody))
		})
	}
}
