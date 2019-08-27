package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Err struct {
	Code    int
	Message string
	Detail  map[string]interface{}
}

func (e *Err) Error() string {
	s, _ := json.Marshal(map[string]interface{}{"error": e})
	return string(s)
}

func (e *Err) Value() interface{} {
	return e.Error()
}

type Response interface {
	StatusCode() int
	Value() interface{}
}

func (e *Err) StatusCode() int {
	return e.Code
}

type WithStatusCode struct {
	Code     int
	Response interface{}
}

func (w *WithStatusCode) Value() interface{} {
	return w.Response
}

func (w *WithStatusCode) StatusCode() int {
	return w.Code
}

func Return(w http.ResponseWriter, res Response) {
	fmt.Printf("Served with: %v \n", http.StatusText(res.StatusCode()))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(res.StatusCode())
	_, _ = fmt.Fprintln(w, res.Value())
}

type handler struct{}

func (*handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer fmt.Println()

	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))

	if r.URL.Path == "" || r.URL.Path == "/" {
		fmt.Printf("Root Request: %s \n", r.Method)
		Return(w, &Err{
			Code:    http.StatusPreconditionFailed,
			Message: "URL not provided",
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": r.URL.String(),
			},
		})
		return
	}

	requestedURL := r.URL.Path[1:]
	if !strings.HasPrefix(requestedURL, "http") {
		requestedURL = "http://" + requestedURL
	}

	p, err := url.ParseRequestURI(requestedURL)
	if err != nil {
		fmt.Printf("Invalid Request: %s - %s \n", r.Method, requestedURL)
		Return(w, &Err{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": requestedURL,
			},
		})
		return
	}

	fmt.Printf("Proxy Request Over: %s - %s \n", r.Method, requestedURL)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Invalid Body")
		Return(w, &Err{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": requestedURL,
			},
		})
		return
	}

	req, err := http.NewRequest(r.Method, p.String(), r.Body)
	if err != nil {
		fmt.Println("Request cannot be created:", err)
		Return(w, &Err{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"body":         b,
				"method":       r.Method,
				"requestedURL": requestedURL,
			},
		})
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Request Failed:", err)
		Return(w, &Err{
			Code:    http.StatusUnprocessableEntity,
			Message: "unknown host",
			Detail: map[string]interface{}{
				"body":         b,
				"method":       r.Method,
				"requestedURL": requestedURL,
				"response":     res,
			},
		})
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Return(w, &Err{
			Code:    res.StatusCode,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"body":         b,
				"requestedURL": requestedURL,

				"response":     res,
				"responseCode": res.StatusCode,
			},
		})
		return
	}
	Return(w, &WithStatusCode{res.StatusCode, string(body)})
}

func main() {
	if err := http.ListenAndServe(":8080", &handler{}); err != nil {
		log.Println("\n\nPanic", err)
	}
}
