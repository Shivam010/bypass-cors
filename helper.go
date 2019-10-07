package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Valuer provides an interface for Return method to write response
type Valuer interface {
	// StatusCode must returns a HTTP Status Code for the response
	StatusCode() int
	// Value must returns the value of the object to write response
	Value() interface{}
}

// Return returns and adds the corresponding Valuer to the ResponseWriter
func Return(w http.ResponseWriter, res Valuer) {
	fmt.Printf("Served with: %d-%v \n", res.StatusCode(), http.StatusText(res.StatusCode()))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(res.StatusCode())
	_, _ = fmt.Fprintln(w, res.Value())
}

// ValuerStruct is a custom wrapper for any struct to implement Valuer
type ValuerStruct struct {
	Code int
	Resp interface{}
}

func (w *ValuerStruct) StatusCode() int {
	return w.Code
}

func (w *ValuerStruct) Value() interface{} {
	return w.Resp
}

// Error is a custom error object that also provides relevant information and
// implement Valuer interface
type Error struct {
	Code    int
	Message string
	Detail  map[string]interface{}
}

func (e *Error) Error() string {
	s, _ := json.Marshal(map[string]interface{}{"error": e})
	return string(s)
}

func (e *Error) Value() interface{} {
	return e.Error()
}

func (e *Error) StatusCode() int {
	return e.Code
}

const (
	// headers
	VaryHeader   = "Vary"
	OriginHeader = "Origin"
	QuoteHeader  = "quote"
	// Access Control headers
	AllowOrigin      = "Access-Control-Allow-Origin"
	AllowMethods     = "Access-Control-Allow-Methods"
	AllowHeaders     = "Access-Control-Allow-Headers"
	AllowCredentials = "Access-Control-Allow-Credentials"
	// Access control request headers
	RequestMethod  = "Access-Control-Request-Method"
	RequestHeaders = "Access-Control-Request-Headers"
)

// defaultHeaders handles a general request and add/set corresponding headers
func defaultHeaders(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	origin := r.Header.Get(OriginHeader)

	// Adding Vary header - for http cache
	headers.Add(VaryHeader, OriginHeader)

	// quote
	headers.Set(QuoteHeader, "Be Happy :)")

	// Allowing only the requester - can be set to "*" too
	headers.Set(AllowOrigin, origin)
	// Always allowing credentials - just for the sake of proxy request
	headers.Set(AllowCredentials, "true")
}

// headersForPreflight handles the pre-flight cors request and add/set the
// corresponding headers
func headersForPreflight(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	reqMethod := r.Header.Get(RequestMethod)
	reqHeaders := r.Header.Get(RequestHeaders)

	// Vary header - for http cache
	headers.Add(VaryHeader, RequestMethod)
	headers.Add(VaryHeader, RequestHeaders)

	// Allowing the requested method
	headers.Set(AllowMethods, strings.ToUpper(reqMethod))
	// Allowing the requested headers
	headers.Set(AllowHeaders, reqHeaders)
}

// addHeaders handles request and set headers accordingly. It returns true if
// request is pre-flight with some Access-Control-Request-Method else false.
func addHeaders(w http.ResponseWriter, r *http.Request) bool {

	defaultHeaders(w, r)

	if r.Method == http.MethodOptions && r.Header.Get(RequestMethod) != "" {
		headersForPreflight(w, r)
		Return(w, &ValuerStruct{Code: http.StatusOK})
		return true
	}

	return false
}

// getRequestURL returns the requested URL to bypass-cors
func getRequestURL(w http.ResponseWriter, r *http.Request) *url.URL {

	if r.URL.Path == "" || r.URL.Path == "/" {
		serveLandingPage(w, r)
		return nil
	}

	p := r.URL.Path[1:]
	if !strings.HasPrefix(p, "http") {
		p = "http://" + p
	}

	URL, err := url.ParseRequestURI(p)
	if err != nil {
		fmt.Println("Invalid Request:", err)
		Return(w, &Error{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": p,
			},
		})
		return nil
	}

	return URL
}

func serveLandingPage(w http.ResponseWriter, r *http.Request)  {
	b, _ := ioutil.ReadFile("public/index.html")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
	return
}
