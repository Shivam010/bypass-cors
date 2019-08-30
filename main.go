package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type handler struct{}

func (*handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer fmt.Println()

	fmt.Printf("Proxy Request Over: %s - %s \n", r.Method, r.URL.String())

	pfr := addHeaders(w, r)
	if pfr { // pre-flight request
		return
	}

	URL := getRequestURL(w, r)
	if URL == nil { // invalid URL
		return
	}

	// extract request body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Invalid Body:", err)
		Return(w, &Error{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": URL.String(),
			},
		})
		return
	}

	// create proxy request
	req, err := http.NewRequest(r.Method, URL.String(), bytes.NewReader(b))
	if err != nil {
		fmt.Println("Request cannot be created:", err)
		Return(w, &Error{
			Code:    http.StatusPreconditionFailed,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"body":         b,
				"method":       r.Method,
				"requestedURL": URL.String(),
			},
		})
		return
	}

	fmt.Println("UserClient --> bypass-cors -->", req.URL.Host)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Request Failed:", err)
		Return(w, &Error{
			Code:    http.StatusUnprocessableEntity,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"body":         b,
				"method":       r.Method,
				"requestedURL": URL.String(),
				"response":     res,
			},
		})
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed to read:", err)
		Return(w, &Error{
			Code:    res.StatusCode,
			Message: err.Error(),
			Detail: map[string]interface{}{
				"method":       r.Method,
				"requestedURL": URL.String(),
				"body":         b,
				"response":     res,
				"responseCode": res.StatusCode,
			},
		})
		return
	}

	Return(w, &ValuerStruct{res.StatusCode, string(body)})
}

func main() {
	var PORT string
	if PORT = os.Getenv("PORT"); PORT == "" {
		flag.StringVar(&PORT, "p", "8080", "PORT at which the server will run")
	}
	flag.Parse()

	fmt.Printf("\nRunning Proxy ByPass Cors Server at port = %v...\n\n", PORT)
	if err := http.ListenAndServe(":"+PORT, &handler{}); err != nil {
		log.Println("\n\nPanic", err)
	}
}
