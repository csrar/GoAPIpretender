package GoAPIpretender

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
)

const (
	parameter = "parameter"
	header    = "header"
)

type T interface {
	Error(args ...any)
	Errorf(format string, args ...any)
}

type ServerMockConfig struct {
	Path           string
	Method         string
	ContentType    string
	Payload        []byte
	Parameters     map[string]string
	Headers        map[string]string
	ResponseStatus int
	ResponseHeader map[string]string
	ResponseBody   []byte
	T              T
}

type ServerMock struct {
	config ServerMockConfig
	server *httptest.Server
}

func (s *ServerMock) checkMethod(method string) error {
	var err error
	if s.config.Method != "" && method != s.config.Method {
		err = fmt.Errorf("invalid method, got: '%s' expected: '%s'", method, s.config.Method)
	}
	return err
}

func (s *ServerMock) checkPath(path string) error {
	var err error
	if s.config.Path != "" && path != s.config.Path {
		err = fmt.Errorf("invalid path, got: '%s' expected: '%s'", path, s.config.Path)
	}
	return err
}

func (s *ServerMock) checkValues(values map[string]string, param string, validators map[string]string) error {
	var errs []error
	for key, value := range validators {
		requestParameter := values[key]
		if requestParameter != value {
			errs = append(errs, fmt.Errorf("invalid %s %s: expected '%s', got '%s'", key, param, value, requestParameter))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errorsToStrings(errs), "\n"))
	}
	return nil
}

func (s *ServerMock) checkPayload(r *http.Request) error {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Failed to read request body: %v", err)
	}
	if s.config.Payload == nil && len(payload) > 0 {
		return fmt.Errorf("Unexpected payload received, got: '%s', but none was expected", string(payload))
	}
	if len(s.config.Payload) > 0 && len(payload) == 0 {
		return fmt.Errorf("Expected a request payload, but none was received")
	}
	return s.comparePayloads(r, payload)
}

func (s *ServerMock) comparePayloads(r *http.Request, payload []byte) error {
	var expectedPayload any
	var actualPayload any
	if !s.isJSONrequest(r) {
		if string(payload) != string(s.config.Payload) {
			return fmt.Errorf("Unexpected payload received, got: '%s', expected: '%s'", string(payload), string(s.config.Payload))
		}
		return nil
	}

	if err := json.Unmarshal(s.config.Payload, &expectedPayload); err != nil {
		return fmt.Errorf("Invalid expected JSON: '%s'", string(s.config.Payload))
	}

	if err := json.Unmarshal(payload, &actualPayload); err != nil {
		return fmt.Errorf("Invalid received JSON: '%s'", string(payload))
	}

	if !reflect.DeepEqual(expectedPayload, actualPayload) {
		return fmt.Errorf("Mismatched JSON payload, got: '%s', expected: '%s'", string(payload), string(s.config.Payload))
	}
	return nil
}

func (s *ServerMock) isJSONrequest(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "application/json")

}

func (s *ServerMock) newHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errs []string
		if err := s.checkMethod(r.Method); err != nil {
			errs = append(errs, err.Error())
		}
		if err := s.checkPath(r.URL.Path); err != nil {
			errs = append(errs, err.Error())
		}
		if err := s.checkValues(getHeaders(r), header, s.config.Headers); err != nil {
			errs = append(errs, err.Error())
		}
		if err := s.checkValues(getParameters(r), parameter, s.config.Parameters); err != nil {
			errs = append(errs, err.Error())
		}
		if err := s.checkPayload(r); err != nil {
			errs = append(errs, err.Error())
		}
		if len(errs) > 0 {
			err := joinErrors(errs)
			if s.config.T != nil {
				s.config.T.Errorf("GoAPIpretender: %v", err)
			} else {
				log.Printf("GoAPIpretender: %v", err)
			}
		}
		s.returnResponse(w)
	})
}

func (s *ServerMock) returnResponse(w http.ResponseWriter) {
	if s.config.ResponseHeader != nil {
		for key, value := range s.config.ResponseHeader {
			w.Header().Set(key, value)
		}
	}
	status := s.config.ResponseStatus
	if status == 0 {
		status = http.StatusOK
	}
	if status >= http.StatusContinue && status <= http.StatusNetworkAuthenticationRequired {
		w.WriteHeader(status)
	}
	if s.config.ResponseBody != nil {
		_, _ = w.Write(s.config.ResponseBody)
	}
}

func (s *ServerMock) Start() string {
	if s.server != nil {
		return s.server.URL
	}
	s.server = httptest.NewServer(s.newHandler())
	return s.server.URL
}

func (s *ServerMock) Server() *httptest.Server {
	return s.server
}

func (s *ServerMock) Stop() {
	if s.server == nil {
		log.Println("GoAPIpretender: warning - server already stopped")
		return
	}
	s.server.Close()
	s.server = nil
	log.Println("GoAPIpretender: server stopped successfully")
}

func NewConfiguredMockServer(config ServerMockConfig) *ServerMock {
	return &ServerMock{
		config: config,
	}
}

func NewDefaultMockServer() *ServerMock {
	return &ServerMock{}
}

func joinErrors(errs []string) string {
	return strings.Join(errs, "\n")
}

func getHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, val := range r.Header {
		headers[key] = strings.Join(val, ", ")
	}
	return headers
}

func getParameters(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, val := range r.URL.Query() {
		params[key] = strings.Join(val, ", ")
	}
	return params
}

func (s *ServerMock) SetMethod(method string) *ServerMock {
	s.config.Method = method
	return s
}

func (s *ServerMock) SetPath(path string) *ServerMock {
	s.config.Path = path
	return s
}

func (s *ServerMock) SetT(t T) *ServerMock {
	s.config.T = t
	return s
}

func (s *ServerMock) SetPayload(payload []byte) *ServerMock {
	s.config.Payload = payload
	return s
}

func (s *ServerMock) SetResponseStatus(status int) *ServerMock {
	s.config.ResponseStatus = status
	return s
}
func (s *ServerMock) SetResponseHeader(headers map[string]string) *ServerMock {
	s.config.ResponseHeader = headers
	return s
}
func (s *ServerMock) SetResponseBody(body []byte) *ServerMock {
	s.config.ResponseBody = body
	return s
}

func errorsToStrings(errs []error) []string {
	strs := make([]string, len(errs))
	for i, err := range errs {
		strs[i] = err.Error()
	}
	return strs
}
