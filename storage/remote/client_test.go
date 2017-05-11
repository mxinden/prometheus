package remote

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
)

func TestStoreHTTPErrorHandling(t *testing.T) {
	tests := []struct {
		code        int
		shouldFail  bool
		recoverable bool
	}{
		{
			code:       200,
			shouldFail: false,
		},
		{
			code:        300,
			shouldFail:  true,
			recoverable: false,
		},
		{
			code:        404,
			shouldFail:  true,
			recoverable: false,
		},
		{
			code:        500,
			shouldFail:  true,
			recoverable: true,
		},
	}

	for i, test := range tests {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "test error", test.code)
			}),
		)

		serverURL, err := url.Parse(server.URL)
		if err != nil {
			panic(err)
		}

		c, err := NewClient(0, &clientConfig{
			url:     &config.URL{serverURL},
			timeout: model.Duration(time.Second),
		})

		err = c.Store(nil)
		if test.shouldFail {
			if err == nil {
				t.Fatalf("%d. Expected Store to fail, but it succeeded", i)
			}
			_, recoverable := err.(recoverableError)
			if test.recoverable != recoverable {
				t.Fatalf("%d. Unexpected recoverability of error; want %v, got %v", i, test.recoverable, recoverable)
			}
		}

		server.Close()
	}
}
