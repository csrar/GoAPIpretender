# GoAPIpretender

*A lightweight, configurable API mock server for Go tests.*

[![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## Overview

**GoAPIpretender** is a **mock HTTP server** designed for **testing API interactions** in Go. It allows developers to:

- Simulate API responses **without external dependencies**
- Validate HTTP **methods, paths, headers, parameters, and payloads**
- Log validation errors **without breaking tests (unless explicitly configured)**
- Ensure **API clients strictly follow expected contract**

---

## Installation

Use `go get` to install GoAPIpretender:

```sh
go get github.com/csrar/GoAPIpretender
```

---

## Usage

###  1 - Basic Mock Server Setup

```go
package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/csrar/GoAPIpretender"
)

func TestAPI(t *testing.T) {
	mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
		Path: "/api/test",
		Method: "POST",
		ResponseStatus: http.StatusCreated,
		ResponseBody: []byte(`{"message": "success"}`),
		T: t, // ✅ Ensures validation errors fail the test
	})
	defer mock.Stop()

	url := mock.Start()
	req, _ := http.NewRequest("POST", url+"/api/test", nil)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}
}
```

**This test creates a mock server, validates the request, and ensures the response is correct.**

---

##  Validation Behavior

**GoAPIpretender** validates incoming API requests **based on the expected configuration**.

| **Scenario** | **`T` is set (`*testing.T`)** | **`T` is NOT set** |
|--------------|----------------|----------------|
| Request **method** is incorrect | ❌ Test **fails (`t.Error()`)** | ⚠️ Logs error, returns **configured response** |
| Request **path** is incorrect | ❌ Test **fails** | ⚠️ Logs error, returns **configured response** |
| **Missing/Incorrect headers** | ❌ Test **fails** | ⚠️ Logs error, returns **configured response** |
| **Missing query parameters** | ❌ Test **fails** | ⚠️ Logs error, returns **configured response** |
| **Payload does not match expected** | ❌ Test **fails** | ⚠️ Logs error, returns **configured response** |

### Important: If `T` is not set, validation errors will NOT fail the test but will be logged.

 If you want strict test failures, **ensure `T` is set** in `ServerMockConfig`.
If `T` is not set, the test **will not fail**, but errors will be logged.

---

##  Configuring Mock API Behavior

### 2- Customizing Expected Requests

```go
mock := GoAPIpretender.NewConfiguredMockServer(GoAPIpretender.ServerMockConfig{
	Path: "/api/data",
	Method: "GET",
	Headers: map[string]string{"Authorization": "Bearer token123"},
	Parameters: map[string]string{"id": "42"},
	Payload: []byte(`{"key":"value"}`),
	T: t,
})
```

This ensures **only** GET requests with **valid headers, parameters, and payloads** are accepted.

### 3- Customizing Responses

```go
	mock.SetResponseStatus(http.StatusOK).
	SetResponseHeader(map[string]string{"Content-Type": "application/json"}).
	SetResponseBody([]byte(`{"message": "ok"}`))
```

The server now returns `200 OK` with a **JSON response**.

---

## Initializing Return Values Without Validations

In some cases, you may want to **return a response** without enforcing validations on method, path, headers, or payload. This can be useful in tests where users need to mock different responses from their services. This can be achieved by simply setting the desired response values without specifying expected request conditions.

```go
mock := GoAPIpretender.NewDefaultMockServer().
	SetResponseStatus(http.StatusOK).
	SetResponseHeader(map[string]string{"Content-Type": "application/json"}).
	SetResponseBody([]byte(`{"message": "unvalidated response"}`))
```

This allows the mock server to always return the specified response, regardless of the request details.

## Mock API Lifecycle

### Start the Server

```go
url := mock.Start()
```

### Stop the Server

```go
mock.Stop()
```

Always call `Stop()` after the test to **clean up resources**.

---

##  Builder Methods

| **Method** | **Description** | **Example** |
|------------|---------------|-------------|
| `NewDefaultMockServer()`| Creates a mock server with default settings  | `mock := GoAPIpretender.NewDefaultMockServer()`
| `SetCustomHandler(handler http.HandlerFunc)` | Sets a custom handler func, if provided all other validations and return values will be ignored | `mock.SetCustomHandler(func(w http.ResponseWriter, r *http.Request) {})` |
| `SetMethod(method string)` | Sets expected HTTP method | `mock.SetMethod("POST")` |
| `SetPath(path string)` | Sets expected request path | `mock.SetPath("/api/test")` |
| `SetT(t *testing.T)` | Attaches a test instance for failure logging | `mock.SetT(t)` |
| `SetPayload(payload []byte)` | Sets expected request body | `mock.SetPayload([]byte(`{"key":"value"}`))` |
| `SetResponseStatus(status int)` | Sets response HTTP status code | `mock.SetResponseStatus(201)` |
| `SetResponseHeader(headers map[string]string)` | Sets response headers | `mock.SetResponseHeader(map[string]string{"Content-Type": "application/json"})` |
| `SetResponseBody(body []byte)` | Sets response body | `mock.SetResponseBody([]byte(`{"message":"ok"}`))` |

---

## Why Use GoAPIpretender?

- **Zero external dependencies** (uses Go’s built-in `httptest.Server`)

- **Fully configurable** (method, path, headers, query params, payloads, responses)

- **Fast & lightweight** (ideal for unit tests)

- **Ensures API contract compliance**

---

##  License

This project is licensed under the [MIT License](LICENSE).

---


---

### Need Help?
**GitHub:** [GoAPIpretender Repo](https://github.com/csrar/GoAPIpretender)

**Happy Testing!** 

