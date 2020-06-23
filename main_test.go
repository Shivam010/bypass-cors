package main

import (
	"bytes"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// args - required arguments for any request to serve
type args struct {
	w *httptest.ResponseRecorder
	r *http.Request
	// a dummy server
	srv *http.Server
}

// resChecker - response checker for checking test response
type resChecker struct {
	code int
	body string
	// set noBodyCheck to true if do not want to check body
	noBodyCheck bool
	// list header keys which should be present in response
	headers []string
}

// defineTest - defines a single unit test
type defineTest struct {
	name   string
	args   *args
	resChr *resChecker
	setup  func(*args, *resChecker)
}

func changeStdOut(s string) *os.File {
	tmp := os.Stdout
	l, _ := os.Create("./bin/logs_" + s)
	os.Stdout = l
	return tmp
}

func resetStdOut(tmp *os.File) {
	os.Stdout = tmp
}

// NOTE: Test_RootRequest should be changed after the error for Root
// request is successfully replaced with the documentation or landing
// page. Follow Issue: Shivam010/bypass-cors#3
func Test_RootRequest(t *testing.T) {
	tmp := changeStdOut(t.Name())
	defer resetStdOut(tmp)
	test := defineTest{
		name: "Root Request",
		args: &args{
			w: httptest.NewRecorder(),
			r: nil,
		},
		resChr: &resChecker{},
		setup: func(ar *args, rc *resChecker) {
			ar.r, _ = http.NewRequest("GET", "/", &bytes.Buffer{})
			rc.code = http.StatusPreconditionFailed
			rc.body = `{"error":{"Code":412,"Message":"URL not provided","Detail":{"method":"GET","requestedURL":"/"}}}` + "\n"
		},
	}
	t.Run(test.name, func(t *testing.T) {
		ha := &handler{}
		test.setup(test.args, test.resChr)
		ha.ServeHTTP(test.args.w, test.args.r)
		if test.args.w.Code != test.resChr.code {
			t.Fatalf("Status code mismatched got: %v, want: %v", test.args.w.Code, test.resChr.code)
		}
		if !test.resChr.noBodyCheck && !cmp.Equal(test.args.w.Body.String(), test.resChr.body) {
			t.Fatalf("Body mismatched got: %s, want: %s", test.args.w.Body.String(), test.resChr.body)
		}
	})
}

func Test_Success(t *testing.T) {
	tmp := changeStdOut(t.Name())
	defer resetStdOut(tmp)
	tests := []defineTest{
		{
			name: "GET-Request",
			args: &args{
				w:   httptest.NewRecorder(),
				r:   nil,
				srv: &http.Server{Addr: ":8181", Handler: http.NotFoundHandler()},
			},
			resChr: &resChecker{
				headers: []string{
					VaryHeader, QuoteHeader,
					AllowOrigin, AllowCredentials,
				},
			},
			setup: func(ar *args, rc *resChecker) {
				ar.srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprintf(w, "Success")
				})
				go ar.srv.ListenAndServe()
				ar.r, _ = http.NewRequest("GET", "/localhost"+ar.srv.Addr, &bytes.Buffer{})
				rc.code = http.StatusOK
				rc.body = fmt.Sprintln("Success")
			},
		},
		// TODO: add test for pre-flight request
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setup(tt.args, tt.resChr)
			defer tt.args.srv.Shutdown(nil)

			ha := &handler{}
			ha.ServeHTTP(tt.args.w, tt.args.r)

			if tt.args.w.Code != tt.resChr.code {
				t.Fatalf("Status code mismatched got: %v, want: %v", tt.args.w.Code, tt.resChr.code)
			}
			if !tt.resChr.noBodyCheck && !cmp.Equal(tt.args.w.Body.String(), tt.resChr.body) {
				t.Fatalf("Body mismatched got: %s, want: %s", tt.args.w.Body.String(), tt.resChr.body)
			}
		})
	}
}

func Test_OtherRequests(t *testing.T) {
	tmp := changeStdOut(t.Name())
	defer resetStdOut(tmp)
	tests := []defineTest{
		{
			name: "Can not Process",
			args: &args{
				w: httptest.NewRecorder(),
				r: nil,
			},
			resChr: &resChecker{},
			setup: func(ar *args, rc *resChecker) {
				ar.r, _ = http.NewRequest("GET", "/invalid-request", &bytes.Buffer{})
				rc.code = http.StatusUnprocessableEntity
				// Note: the error message `Get http://invalid-request: dial tcp: lookup invalid-request: no such host`
				// 	varies from environment to environment and hence, omitting the check
				rc.noBodyCheck = true
				//rc.body = `{"error":{"Code":422,"Message":"Get http://invalid-request: dial tcp: lookup invalid-request: no such host","Detail":{"body":"","method":"GET","requestedURL":"http://invalid-request","response":null}}}` + "\n"
			},
		},
		{
			name: "Invalid Request",
			args: &args{
				w: httptest.NewRecorder(),
				r: nil,
			},
			resChr: &resChecker{},
			setup: func(ar *args, rc *resChecker) {
				ar.r, _ = http.NewRequest("GET", "", &bytes.Buffer{})
				ar.r.URL.Path = "%invalid%"
				rc.code = http.StatusPreconditionFailed
				rc.body = `{"error":{"Code":412,"Message":"parse http://invalid%: invalid URL escape \"%\"","Detail":{"method":"GET","requestedURL":"http://invalid%"}}}` + "\n"
			},
		},
		{
			name: "URL not Provided",
			args: &args{
				w: httptest.NewRecorder(),
				r: nil,
			},
			resChr: &resChecker{},
			setup: func(ar *args, rc *resChecker) {
				ar.r, _ = http.NewRequest("GET", "", &bytes.Buffer{})
				rc.code = http.StatusPreconditionFailed
				rc.body = `{"error":{"Code":412,"Message":"URL not provided","Detail":{"method":"GET","requestedURL":""}}}` + "\n"
			},
		},
		{
			name: "Invalid Method",
			args: &args{
				w: httptest.NewRecorder(),
				r: nil,
			},
			resChr: &resChecker{},
			setup: func(ar *args, rc *resChecker) {
				ar.r, _ = http.NewRequest("GET", "/localhost", &bytes.Buffer{})
				ar.r.Method += "/"
				rc.code = http.StatusPreconditionFailed
				rc.body = `{"error":{"Code":412,"Message":"net/http: invalid method \"GET/\"","Detail":{"body":"","method":"GET/","requestedURL":"http://localhost"}}}` + "\n"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := &handler{}
			tt.setup(tt.args, tt.resChr)
			ha.ServeHTTP(tt.args.w, tt.args.r)
			if tt.args.w.Code != tt.resChr.code {
				t.Fatalf("Status code mismatched got: %v, want: %v", tt.args.w.Code, tt.resChr.code)
			}
			if !tt.resChr.noBodyCheck && !cmp.Equal(tt.args.w.Body.String(), tt.resChr.body) {
				t.Fatalf("Body mismatched got: %s, want: %s", tt.args.w.Body.String(), tt.resChr.body)
			}
		})
	}
}
