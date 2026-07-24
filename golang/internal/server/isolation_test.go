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
	return New(cs, stubPinger{}, isolation.NewManager(cs))
}

func postIso(srv *Server) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/isolation", strings.NewReader(isoBody)))
	return rec
}

func deleteIso(srv *Server, id string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/isolation/"+id, nil))
	return rec
}

func TestCreateIsolation(t *testing.T) {
	t.Run("when the pair is posted then it returns 201, and a repeat returns 200 (idempotent)", func(t *testing.T) {
		srv := newSrv()
		assert.Equal(t, http.StatusCreated, postIso(srv).Code)
		assert.Equal(t, http.StatusOK, postIso(srv).Code)
	})

	t.Run("when the body is missing namespaces then it returns 400", func(t *testing.T) {
		srv := newSrv()
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/isolation", strings.NewReader(`{}`)))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestListIsolation(t *testing.T) {
	t.Run("when GET /isolation then it returns 200", func(t *testing.T) {
		srv := newSrv()
		postIso(srv)
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/isolation", nil))
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestDeleteIsolation(t *testing.T) {
	t.Run("when the id is unknown then it returns 404", func(t *testing.T) {
		assert.Equal(t, http.StatusNotFound, deleteIso(newSrv(), "does-not-exist").Code)
	})

	t.Run("when a real isolation id is deleted then it returns 200", func(t *testing.T) {
		srv := newSrv()
		postIso(srv)
		id := isolation.DeriveID(sampleReq())
		assert.Equal(t, http.StatusOK, deleteIso(srv, id).Code)
	})
}