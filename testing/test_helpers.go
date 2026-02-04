package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type HTTPTestHelper struct {
	tc *TestCase
}

func NewHTTPTestHelper(tc *TestCase) *HTTPTestHelper {
	return &HTTPTestHelper{tc: tc}
}

func (h *HTTPTestHelper) Get(path string, headers ...map[string]string) (*http.Response, error) {
	return h.request("GET", path, nil, headers...)
}

func (h *HTTPTestHelper) Post(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return h.request("POST", path, body, headers...)
}

func (h *HTTPTestHelper) Put(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return h.request("PUT", path, body, headers...)
}

func (h *HTTPTestHelper) Patch(path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	return h.request("PATCH", path, body, headers...)
}

func (h *HTTPTestHelper) Delete(path string, headers ...map[string]string) (*http.Response, error) {
	return h.request("DELETE", path, nil, headers...)
}

func (h *HTTPTestHelper) request(method, path string, body interface{}, headers ...map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, path, bodyReader)

	req.Header.Set("Content-Type", "application/json")

	if len(headers) > 0 {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	return h.tc.App.Test(req, -1)
}

func (h *HTTPTestHelper) WithAuth(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

func (h *HTTPTestHelper) WithHeaders(headers map[string]string) map[string]string {
	return headers
}

func (h *HTTPTestHelper) ParseJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (h *HTTPTestHelper) GetBody(resp *http.Response) string {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

type AssertHelper struct {
	tc *TestCase
}

func NewAssertHelper(tc *TestCase) *AssertHelper {
	return &AssertHelper{tc: tc}
}

func (a *AssertHelper) AssertStatus(resp *http.Response, expectedStatus int) {
	a.tc.Equal(expectedStatus, resp.StatusCode, "Expected status %d, got %d", expectedStatus, resp.StatusCode)
}

func (a *AssertHelper) AssertOK(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusOK)
}

func (a *AssertHelper) AssertCreated(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusCreated)
}

func (a *AssertHelper) AssertNoContent(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusNoContent)
}

func (a *AssertHelper) AssertBadRequest(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusBadRequest)
}

func (a *AssertHelper) AssertUnauthorized(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusUnauthorized)
}

func (a *AssertHelper) AssertNotFound(resp *http.Response) {
	a.AssertStatus(resp, fiber.StatusNotFound)
}

func (a *AssertHelper) AssertJSON(resp *http.Response, key string, expectedValue interface{}) {
	var data map[string]interface{}
	helper := NewHTTPTestHelper(a.tc)
	err := helper.ParseJSON(resp, &data)

	a.tc.NoError(err, "Failed to parse JSON response")
	a.tc.Equal(expectedValue, data[key], "Expected %v for key %s, got %v", expectedValue, key, data[key])
}

func (a *AssertHelper) AssertJSONContains(resp *http.Response, key string) {
	var data map[string]interface{}
	helper := NewHTTPTestHelper(a.tc)
	err := helper.ParseJSON(resp, &data)

	a.tc.NoError(err, "Failed to parse JSON response")
	a.tc.Contains(data, key, "Expected JSON to contain key: %s", key)
}

func (a *AssertHelper) AssertBodyContains(resp *http.Response, substring string) {
	helper := NewHTTPTestHelper(a.tc)
	body := helper.GetBody(resp)
	a.tc.True(strings.Contains(body, substring), "Expected body to contain: %s", substring)
}

type DatabaseHelper struct {
	tc *TestCase
}

func NewDatabaseHelper(tc *TestCase) *DatabaseHelper {
	return &DatabaseHelper{tc: tc}
}

func (d *DatabaseHelper) AssertDatabaseHas(table string, conditions map[string]interface{}) {
	db := d.tc.GetDB()
	var count int64

	query := db.Table(table)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&count)
	d.tc.True(count > 0, "Expected to find record in table %s with conditions %v", table, conditions)
}

func (d *DatabaseHelper) AssertDatabaseMissing(table string, conditions map[string]interface{}) {
	db := d.tc.GetDB()
	var count int64

	query := db.Table(table)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&count)
	d.tc.True(count == 0, "Expected NOT to find record in table %s with conditions %v", table, conditions)
}

func (d *DatabaseHelper) AssertDatabaseCount(table string, expectedCount int) {
	db := d.tc.GetDB()
	var count int64

	db.Table(table).Count(&count)
	d.tc.Equal(int64(expectedCount), count, "Expected %d records in table %s, got %d", expectedCount, table, count)
}

func (d *DatabaseHelper) Create(model interface{}) error {
	return d.tc.GetDB().Create(model).Error
}

func (d *DatabaseHelper) Find(dest interface{}, conditions ...interface{}) error {
	return d.tc.GetDB().First(dest, conditions...).Error
}

func (d *DatabaseHelper) Delete(model interface{}) error {
	return d.tc.GetDB().Delete(model).Error
}
