package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SowmyaRahul/tyk-sre-assignment/internal/isolation"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

const isoBody = `{"a":{"namespace":"team-a","podSelector":{"app":"checkout"}},
                  "b":{"namespace":"team-b","podSelector":{"app":"payments"}}}`

// sampleReq mirrors isoBody so tests can derive the isolation id.
func sampleReq() isolation.Request {
	return isolation.Request{
		A: isolation.Workload{Namespace: "team-a", PodSelector: map[string]string{"app": "checkout"}},
		B: isolation.Workload{Namespace: "team-b", PodSelector: map[string]string{"app": "payments"}},
	}
}

func newSrv() *Server {
	cs := fake.NewSimpleClientset()
	return New(cs, stubPinger{}, isolation.NewManager(cs), "s3cret")
}

func postIso(srv *Server, token string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/isolation", strings.NewReader(isoBody))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func deleteIso(srv *Server, id, token string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/isolation/"+id, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestCreateIsolation(t *testing.T) {
	t.Run("when there is no bearer token then it returns 401", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, postIso(newSrv(), "").Code)
	})

	t.Run("when a valid token is used then it returns 201, and a repeat returns 200 (idempotent)", func(t *testing.T) {
		srv := newSrv()
		assert.Equal(t, http.StatusCreated, postIso(srv, "s3cret").Code)
		assert.Equal(t, http.StatusOK, postIso(srv, "s3cret").Code)
	})
}

func TestListIsolation(t *testing.T) {
	t.Run("when GET /isolation then it returns 200 without a token (read-only is open)", func(t *testing.T) {
		// Given a server with one isolation applied
		srv := newSrv()
		postIso(srv, "s3cret")
		// When we GET /isolation with no token
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/isolation", nil))
		// Then it is 200
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestDeleteIsolation(t *testing.T) {
	t.Run("when there is no bearer token then it returns 401", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, deleteIso(newSrv(), "anything", "").Code)
	})

	t.Run("when the id is unknown then it returns 404", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, deleteIso(newSrv(), "does-not-exist", "s3cret").Code)
	})

	t.Run("when a real isolation id is deleted then it returns 200", func(t *testing.T) {
		// Given an applied isolation
		srv := newSrv()
		postIso(srv, "s3cret")
		id := isolation.DeriveID(sampleReq())
		// When we DELETE it with a token, then it is 200
		assert.Equal(t, http.StatusOK, deleteIso(srv, id, "s3cret").Code)
	})
}
