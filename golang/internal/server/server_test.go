package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHealthz(t *testing.T) {
	t.Run("when GET /healthz then it returns 200 with body ok", func(t *testing.T) {
		srv := New(fake.NewSimpleClientset())

		
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		body, err := io.ReadAll(rec.Result().Body)
		assert.NoError(t, err)
		assert.Equal(t, "ok", string(body))
	})
}