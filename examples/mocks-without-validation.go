package examples

import (
	"net/http"
	"testing"

	"github.com/csrar/GoAPIpretender"
)

func TestMockServerWithoutValidation(t *testing.T) {
	mock := GoAPIpretender.NewDefaultMockServer().
		SetResponseStatus(http.StatusOK).
		SetResponseBody([]byte(`{"message": "no validation"}`))

	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestMockServerWithErrorResponse(t *testing.T) {
	mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
		ResponseStatus: http.StatusInternalServerError,
		ResponseBody:   []byte(`{"error": "server failure"}`),
	})
	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("GET", url+"/api/error", nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", resp.StatusCode)
	}
}
