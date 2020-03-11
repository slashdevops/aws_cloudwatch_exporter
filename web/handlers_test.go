package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_Handler(t *testing.T) {
	tests := []struct{
		name string
		in *http.Request
		out *httptest.ResponseRecorder
		expectedStatus int
		expectedBody string
	}{
		{
			name:"Test Home",
			in: httptest.NewRequest(http.MethodGet, "/", nil),
			out: httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
			expectedBody: "message",
		},
	}

	for _, test := range tests{
		test := test
		t.Run(test.name, func(t *testing.T) {
			h:= NewHandlers(nil,nil)
			h.Home(test.out,test.in)
			if test.out.Code != test.expectedStatus{
				t.Logf("expected: %d\ngot: %d\n", test.expectedStatus,test.out.Code)
				t.Fail()
			}

			body:= test.out.Body.String()
			if body != test.expectedBody{
				t.Logf("expected: %v\ngot: %v\n", test.expectedBody,body)
				t.Fail()
			}
		})
	}
}
/*
func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/healthz", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned %v", status)
	}

	// 	expected := `OK`
	// 	if rr.Body.String() != expected {
	// 		t.Errorf("Handler returned %v", rr.Body.String())
	// 	}
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned %v", status)
	}
}
*/