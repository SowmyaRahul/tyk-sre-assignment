package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

// stubPinger lets us exercise both readiness branches deterministically.
type stubPinger struct{ err error }

func (s stubPinger) Ping() error { return s.err }

func TestHealthz(t *testing.T) {
	t.Run("when the API server is down then /healthz still returns 200 ok (liveness ignores the API)", func(t *testing.T) {
		// Given a server whose pinger reports the API is down
		srv := New(fake.NewSimpleClientset(), stubPinger{err: errors.New("api down")}, nil, "")

		// When we GET /healthz
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		// Then it is still 200 ok — liveness must not depend on the API server
		assert.Equal(t, http.StatusOK, rec.Code)
		body, err := io.ReadAll(rec.Result().Body)
		assert.NoError(t, err)
		assert.Equal(t, "ok", string(body))
	})
}

func TestReadyz(t *testing.T) {
	t.Run("when the API server is reachable then /readyz returns 200", func(t *testing.T) {
		srv := New(fake.NewSimpleClientset(), stubPinger{}, nil, "")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("when the API server is unreachable then /readyz returns 503", func(t *testing.T) {
		srv := New(fake.NewSimpleClientset(), stubPinger{err: errors.New("conn refused")}, nil, "")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})
}