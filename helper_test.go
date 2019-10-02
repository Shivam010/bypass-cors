package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestReturn(t *testing.T) {
	val := &ValuerStruct{Code: http.StatusNotFound, Resp: "hello CORs!"}

	rec := httptest.NewRecorder()
	Return(rec, val)
	resp := rec.Result()

	expectedContentType := "application/json; charset=utf-8"

	if resp.Header.Get("Content-Type") != expectedContentType {
		t.Errorf("Incorrect header \"Content-Type\".\nExpected: %q\n  Actual: %q\n", expectedContentType, resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Incorrect HTTP status code.\nExpected: %v\n  Actual: %q\n", http.StatusNotFound, resp.StatusCode)
	}
}

func TestDefaultHeaders(t *testing.T) {
	expectedOrigin := "origin-header"
	expectedHeaders := map[string]string{
		"Vary":  "Origin",
		"quote": "Be Happy :)",
		"Access-Control-Allow-Origin":      expectedOrigin,
		"Access-Control-Allow-Credentials": "true",
	}

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", expectedOrigin)

	rec := httptest.NewRecorder()
	defaultHeaders(rec, req)
	resp := rec.Result()

	for header, value := range expectedHeaders {
		if resp.Header.Get(header) != value {
			t.Errorf("Incorrect value for header %q.\nExpected: %q\n  Actual: %q", header, value, resp.Header.Get(header))
		}
	}
}

func TestHeadersForPreflight(t *testing.T) {
	expectedRequestMethod := "post"
	expectedRequestHeaders := "X-PINGOTHER, Content-Type"
	expectedVary := []string{"Authentication", "Access-Control-Request-Method", "Access-Control-Request-Headers"}

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Methods": "POST",
		"Access-Control-Allow-Headers": expectedRequestHeaders,
	}

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Access-Control-Request-Method", expectedRequestMethod)
	req.Header.Set("Access-Control-Request-Headers", expectedRequestHeaders)

	rec := httptest.NewRecorder()
	rec.Header().Set("Vary", "Authentication")
	headersForPreflight(rec, req)
	resp := rec.Result()

	actualVary := resp.Header["Vary"]
	if !reflect.DeepEqual(actualVary, expectedVary) {
		t.Errorf("Incorrect value for header %q.\nExpected: %q\n  Actual: %q", "Vary", expectedVary, actualVary)
	}

	for header, value := range expectedHeaders {
		if resp.Header.Get(header) != value {
			t.Errorf("Incorrect value for header %q.\nExpected: %q\n  Actual: %q", header, value, resp.Header.Get(header))
		}
	}
}

func TestAddHeaders(t *testing.T) {
	tests := []struct {
		Method          string
		RequestHeaders  map[string]string
		Origin          string
		Preflight       bool
		Vary            []string
		ResponseHeaders map[string]string
	}{
		{
			"GET",
			map[string]string{},
			"origin-header-value-get",
			false,
			[]string{"Origin"},
			map[string]string{
				"Access-Control-Allow-Origin": "origin-header-value-get",
			},
		},
		{
			"OPTIONS",
			map[string]string{
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "Content-Type",
			},
			"origin-header-value-options",
			true,
			[]string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
			map[string]string{
				"Access-Control-Allow-Origin": "origin-header-value-options",
			},
		},
	}

	constantHeaders := map[string]string{
		"quote": "Be Happy :)",
		"Access-Control-Allow-Credentials": "true",
	}

	for _, test := range tests {
		t.Run(test.Method, func(t *testing.T) {
			req := httptest.NewRequest(test.Method, "/", nil)
			for h, v := range test.RequestHeaders {
				req.Header.Set(h, v)
			}

			rec := httptest.NewRecorder()
			isPreflight := addHeaders(rec, req)
			if test.Preflight != isPreflight {
				t.Errorf("Return value incorrect.  Expected %v, actual %v.", test.Preflight, isPreflight)
			}
			resp := rec.Result()

			for header, value := range constantHeaders {
				if resp.Header.Get(header) != value {
					t.Errorf("Incorrect value for header %q.\nExpected: %q\n  Actual: %q", header, value, resp.Header.Get(header))
				}
			}

			actualVary := resp.Header["Vary"]
			if !reflect.DeepEqual(actualVary, test.Vary) {
				t.Errorf("Incorrect value for header %q.\nExpected: %q\n  Actual: %q", "Vary", test.Vary, actualVary)
			}
		})
	}
}
