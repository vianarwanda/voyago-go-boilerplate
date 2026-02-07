package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// HTTPTestHelper provides utilities for HTTP testing
type HTTPTestHelper struct {
	App *fiber.App
	T   *testing.T
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(app *fiber.App, t *testing.T) *HTTPTestHelper {
	return &HTTPTestHelper{
		App: app,
		T:   t,
	}
}

// POST makes a POST request to the given path with JSON body
func (h *HTTPTestHelper) POST(path string, body interface{}) *httptest.ResponseRecorder {
	h.T.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			h.T.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest("POST", path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.App.Test(req, -1)
	if err != nil {
		h.T.Fatalf("Failed to execute request: %v", err)
	}

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			recorder.Header().Add(key, value)
		}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	recorder.Body.Write(bodyBytes)

	return recorder
}

// GET makes a GET request to the given path
func (h *HTTPTestHelper) GET(path string) *httptest.ResponseRecorder {
	h.T.Helper()

	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.App.Test(req, -1)
	if err != nil {
		h.T.Fatalf("Failed to execute request: %v", err)
	}

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			recorder.Header().Add(key, value)
		}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	recorder.Body.Write(bodyBytes)

	return recorder
}

// AssertJSONResponse asserts the response status and decodes JSON body
func (h *HTTPTestHelper) AssertJSONResponse(resp *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	h.T.Helper()

	assert.Equal(h.T, expectedStatus, resp.Code, "Unexpected status code")

	if target != nil {
		err := json.NewDecoder(resp.Body).Decode(target)
		assert.NoError(h.T, err, "Failed to decode JSON response")
	}
}

// AssertErrorResponse asserts error response structure
func (h *HTTPTestHelper) AssertErrorResponse(resp *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	h.T.Helper()

	assert.Equal(h.T, expectedStatus, resp.Code, "Unexpected status code")

	var errResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	assert.NoError(h.T, err, "Failed to decode error response")

	// Assert standard error response fields exist
	assert.Contains(h.T, errResp, "status", "Error response should contain 'status'")
	assert.Contains(h.T, errResp, "message", "Error response should contain 'message'")

	return errResp
}
