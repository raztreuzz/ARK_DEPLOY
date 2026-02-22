package instances

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockRouteStore struct {
	routes map[string]struct {
		host string
		port int
	}
	putErr error
	getErr error
	delErr error
}

func newMockRouteStore() *mockRouteStore {
	return &mockRouteStore{
		routes: map[string]struct {
			host string
			port int
		}{},
	}
}

func (m *mockRouteStore) GetRoute(instanceID string) (string, int, bool, error) {
	if m.getErr != nil {
		return "", 0, false, m.getErr
	}
	v, ok := m.routes[instanceID]
	if !ok {
		return "", 0, false, nil
	}
	return v.host, v.port, true, nil
}

func (m *mockRouteStore) PutRoute(instanceID string, host string, port int) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.routes[instanceID] = struct {
		host string
		port int
	}{host: host, port: port}
	return nil
}

func (m *mockRouteStore) DeleteRoute(instanceID string) error {
	if m.delErr != nil {
		return m.delErr
	}
	delete(m.routes, instanceID)
	return nil
}

func TestRegisterRoute_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := newMockRouteStore()
	RegisterRoutes(r, store)

	payload := RegisterReq{
		InstanceID: "i-1",
		TargetHost: "100.103.96.26",
		TargetPort: 18080,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/instances/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRegisterRoute_InvalidPort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, newMockRouteStore())

	payload := RegisterReq{
		InstanceID: "i-1",
		TargetHost: "100.103.96.26",
		TargetPort: 70000,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/instances/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDeleteRoute_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := newMockRouteStore()
	store.delErr = errors.New("boom")
	RegisterRoutes(r, store)

	req := httptest.NewRequest(http.MethodDelete, "/instances/i-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
