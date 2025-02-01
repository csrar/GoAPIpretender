package GoAPIpretender

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type TestCapture struct {
	failed bool
	Errors []string
}

func (tc *TestCapture) Error(args ...interface{}) {
	tc.Errors = append(tc.Errors, args[0].(string)) // Capture error messages
}
func (tc *TestCapture) Errorf(format string, args ...any) {
	tc.Errors = append(tc.Errors, fmt.Sprintf(format, args...))
	tc.failed = true
}
func (tc *TestCapture) checkErrors(expectedErrors []string, t *testing.T) {
	for _, expected := range expectedErrors {
		found := false
		for _, actual := range tc.Errors {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error message '%s' not found in captured errors", expected)
		}
	}
}

func TestServerMock_StartAndStop(t *testing.T) {
	mock := NewDefaultMockServer()
	url := mock.Start()
	secondStartUrl := mock.Start()
	if url == "" {
		t.Fatal("Expected non-empty server URL")
	}
	mock.Stop()
	if mock.server != nil {
		t.Fatal("Expected server to be nil after Stop()")
	}
	if url != secondStartUrl {
		t.Fatalf("expected URLs after a second initialization should be the same, first call '%s', second call '%s'", url, secondStartUrl)
	}
}

func TestServerMock_MethodValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: invalid method, got: 'GET' expected: 'POST'"}
	mock := NewConfiguredMockServer(ServerMockConfig{
		Method: "POST",
		T:      tc,
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("GET", url, nil)
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_PathValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: invalid path, got: '/wrong-path' expected: '/expected-path'"}
	mock := NewConfiguredMockServer(ServerMockConfig{
		Path: "/expected-path",
		T:    tc,
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("GET", url+"/wrong-path", nil)
	http.DefaultClient.Do(req)
	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_HeaderValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: invalid X-Test-Header header: expected 'expected-value', got 'wrong-value'"}
	mock := NewConfiguredMockServer(ServerMockConfig{
		Headers: map[string]string{"X-Test-Header": "expected-value"},
		T:       tc,
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Test-Header", "wrong-value")
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_QueryParameterValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: invalid param1 parameter: expected 'value1', got 'wrong-value'"}
	mock := NewConfiguredMockServer(ServerMockConfig{
		Parameters: map[string]string{"param1": "value1"},
		T:          tc,
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("GET", url+"?param1=wrong-value", nil)
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_PayloadValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Mismatched JSON payload, got: '{\"key\":\"wrong-value\"}', expected: '{\"key\":\"value\"}'"}
	expectedPayload := []byte(`{"key":"value"}`)
	mock := NewDefaultMockServer().SetPayload(expectedPayload).SetMethod("POST").SetT(tc)
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"key":"wrong-value"}`)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}
func TestServerMock_MissmatchingPayloadValidation(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Unexpected payload received, got: '{\"key\":\"wrong-value\"}', expected: '{\"key\":\"value\"}'"}
	expectedPayload := []byte(`{"key":"value"}`)
	mock := NewDefaultMockServer().SetPayload(expectedPayload).SetMethod("POST").SetT(tc)
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"key":"wrong-value"}`)))
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}
func TestServerMock_InvalidJSONexpected(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Invalid expected JSON: '{\"key:\"value}'"}
	expectedPayload := []byte(`{"key:"value}`)
	mock := NewDefaultMockServer().SetPayload(expectedPayload).SetMethod("POST").SetT(tc)
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"key":"wrong-value"}`)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}
func TestServerMock_InvalidJSONpayload(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Invalid received JSON: '{\"key:\"wrong-value}'"}
	expectedPayload := []byte(`{"key":"value"}`)
	mock := NewDefaultMockServer().SetPayload(expectedPayload).SetMethod("POST").SetT(tc)
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"key:"wrong-value}`)))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_NotExpectedPayload(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Unexpected payload received, got: 'mock-payload', but none was expected"}

	mock := NewConfiguredMockServer(ServerMockConfig{}).SetT(tc)
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, strings.NewReader("mock-payload"))
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}
func TestServerMock_MissingPayload(t *testing.T) {
	tc := &TestCapture{}
	expectedErrors := []string{"GoAPIpretender: Expected a request payload, but none was received"}

	mock := NewConfiguredMockServer(ServerMockConfig{}).SetT(tc).SetPayload([]byte("expected-payload"))
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, nil)
	http.DefaultClient.Do(req)

	tc.checkErrors(expectedErrors, t)
}

func TestServerMock_CustomMock(t *testing.T) {
	expectedBody := "mock-message"
	expectedStatus := 201

	mock := NewConfiguredMockServer(ServerMockConfig{}).SetCustomHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatus)
		fmt.Fprintf(w, expectedBody)
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, nil)
	response, _ := http.DefaultClient.Do(req)
	responseBody := make([]byte, len(expectedBody))
	response.Body.Read(responseBody)
	if expectedBody != string(responseBody) {
		t.Errorf("received unnexpected response, got: '%s', expected: '%s'", string(responseBody), expectedBody)
	}
	if response.StatusCode != expectedStatus {
		t.Errorf("received unnexpected status, got:'%d' expected:'%d", response.StatusCode, expectedStatus)
	}
}

func TestServerMock_LogErrors(t *testing.T) {
	mock := NewConfiguredMockServer(ServerMockConfig{}).SetPayload([]byte("expected-payload")).SetMethod("GET")
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, nil)
	http.DefaultClient.Do(req)
}

func TestServerMock_ResponseStatusAndBody(t *testing.T) {
	expectedBody := []byte(`{"object":{"key":"value"},"array":[1,"text",false,null],"string":"hello","number":42,"boolean":true,"null":null}`)
	expectedHeaders := map[string]string{"Custom-Header": "success"}
	mock := NewConfiguredMockServer(ServerMockConfig{
		ResponseStatus: http.StatusCreated,
		ResponseBody:   expectedBody,
		ResponseHeader: expectedHeaders,
		Payload:        expectedBody,
		T:              t,
	})
	defer mock.Stop()
	url := mock.Start()

	req, _ := http.NewRequest("POST", url, strings.NewReader(string(expectedBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}

	body := make([]byte, len(expectedBody))
	resp.Body.Read(body)
	resp.Body.Close()
	if string(body) != string(expectedBody) {
		t.Errorf("Expected response body %s, got %s", expectedBody, body)
	}
	for key, value := range expectedHeaders {
		headerValue := resp.Header.Get(key)
		if value != headerValue {
			t.Errorf("Header '%s' returned an unexpected value, got: '%s', expected: '%s'", key, headerValue, value)
		}
	}
}

func TestServerMock_BuilderMethods(t *testing.T) {
	mock := NewDefaultMockServer().
		SetMethod("POST").
		SetPath("/test").
		SetResponseStatus(http.StatusAccepted).
		SetResponseHeader(map[string]string{"Content-Type": "application/json"}).
		SetResponseBody([]byte(`{"status":"ok"}`))

	if mock.config.Method != "POST" {
		t.Errorf("SetMethod failed, expected POST, got %s", mock.config.Method)
	}
	if mock.config.Path != "/test" {
		t.Errorf("SetPath failed, expected /test, got %s", mock.config.Path)
	}
	if mock.config.ResponseStatus != http.StatusAccepted {
		t.Errorf("SetResponseStatus failed, expected 202, got %d", mock.config.ResponseStatus)
	}
	if mock.config.ResponseHeader["Content-Type"] != "application/json" {
		t.Errorf("SetResponseHeader failed, expected application/json, got %s", mock.config.ResponseHeader["Content-Type"])
	}
	if string(mock.config.ResponseBody) != `{"status":"ok"}` {
		t.Errorf("SetResponseBody failed, expected `{\"status\":\"ok\"}`, got %s", string(mock.config.ResponseBody))
	}
}

func TestServerMock_SetT(t *testing.T) {
	mock := NewDefaultMockServer().SetT(t)
	if mock.config.T == nil {
		t.Error("SetT failed, expected non-nil testing.T reference")
	}
}

func TestServerMock_Stop(t *testing.T) {
	mock := NewConfiguredMockServer(ServerMockConfig{})
	mock.Start()

	mock.Stop()
	mock.Stop()

	if mock.server != nil {
		t.Error("Expected server to be nil after Stop()")
	}
}

func TestServerMock_Server(t *testing.T) {
	mock := NewConfiguredMockServer(ServerMockConfig{})
	httpServer := mock.Server()
	if httpServer != nil {
		t.Error("Expected http server to be nil before Start()")
	}
	mock.Start()
	httpServer = mock.Server()
	if httpServer == nil {
		t.Error(" http server should not be nil after Start()")
	}
}
