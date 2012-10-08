package main

import (
	"testing"
	"bytes"
	"fmt"
	"net/http"
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
)

type MockCredentialsVerifierActivityTypesGetter struct {
}

func (db *MockCredentialsVerifierActivityTypesGetter) GetActivityTypesForUser(user_id int64) []ActivityType {
	return []ActivityType{ 
		{ Id: 1, Name: "Foobar" },
		{ Id: 2, Name: "Quux" },
	}
}

func (db *MockCredentialsVerifierActivityTypesGetter) VerifyCredentials(username, password string) (user_id int64, authenticated bool) {
	return 42, (username == "foo" && password == "bar")
}

func TestAuthenticateHandler(t *testing.T) {
	testdata := []struct{
		auth_data string
		http_code int
		authenticated_result bool
	} {
		{ "username=foo&password=bar", http.StatusOK, true },
		{ "username=invalid&password=invalid", http.StatusOK, false },
	}

	for _, td := range testdata {
		req, _ := http.NewRequest("POST", "http://localhost/auth",  bytes.NewBufferString(td.auth_data))
		req.Header["Content-Length"] = []string{fmt.Sprintf("%d", len(td.auth_data))}
		req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

		mock_db := &MockCredentialsVerifierActivityTypesGetter{}

		handler := &AuthenticateHandler{Db: mock_db, Store: sessions.NewCookieStore([]byte(""))}

		resp := NewMockResponseWriter()

		handler.ServeHTTP(resp, req)

		if resp.StatusCode != td.http_code {
			t.Errorf("AuthenticateHandler responded with %d (expected: %d)", resp.StatusCode, td.http_code)
		}

		data := make(map[string]interface{})
		if err := json.Unmarshal(resp.Buffer.Bytes(), &data); err != nil {
			t.Errorf("AuthenticateHandler returned invalid JSON: %v", err)
		}

		if authenticated, ok := data["authenticated"].(bool); !ok {
			t.Errorf("JSON authenticated field didn't contain bool")
		} else {
			if authenticated != td.authenticated_result {
				t.Errorf("AuthenticateHandler authenticated returns %v (expected %v)", authenticated, td.authenticated_result)
			}
		}
	}
}
