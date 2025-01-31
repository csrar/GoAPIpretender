package examples

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/csrar/GoAPIpretender"
)

// Example 1: Basic Mock Server with GET Request
func TestBasicMockServer(t *testing.T) {
	mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
		Path:           "/api/data",
		Method:         "GET",
		ResponseStatus: http.StatusOK,
		ResponseBody:   []byte(`{"message": "success"}`),
		T:              t, // Ensures validation errors fail the test
	})
	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("GET", url+"/api/data", nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

// Example 2: Mock Server with Headers Validation
func TestMockServerWithHeaders(t *testing.T) {
	mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
		Path:           "/api/secure",
		Method:         "POST",
		Headers:        map[string]string{"Authorization": "Bearer token123"},
		ResponseStatus: http.StatusOK,
		ResponseBody:   []byte(`{"status": "authorized"}`),
		T:              t,
	})
	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("POST", url+"/api/secure", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

// Example 3: Mock Server with Query Parameters
func TestMockServerWithQueryParams(t *testing.T) {
	mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
		Path:           "/api/query",
		Method:         "GET",
		Parameters:     map[string]string{"id": "42"},
		ResponseStatus: http.StatusOK,
		ResponseBody:   []byte(`{"result": "found"}`),
		T:              t,
	})
	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/query?id=42", url), nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
