package controller

import (
	"encoding/json"
	"testing"
)

func TestIsPublicOperation(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"login", `query { login(email: "a", password: "b") { session { accessToken } } }`, true},
		{"register", `mutation { register(input: {}) { user { id } } }`, true},
		{"refresh", `mutation { refreshToken(refreshToken: "x") { session { accessToken } } }`, true},
		{"introspection", `query { __schema { types { name } } }`, true},
		{"protected", `query { me { id } }`, false},
		{"protected with login in text", `query { searchYoutubeFromApi(query: "login") { totalCount } }`, false},
		{"protected with login in comment", "# login\nquery { me { id } }", false},
		{"protected aliased as login", `query { login: me { id } }`, false},
		{"public and protected together", `query { login(email: "a", password: "b") { session { accessToken } } me { id } }`, false},
		{"second operation protected", `query A { login(email: "a", password: "b") { session { accessToken } } } query B { me { id } }`, false},
		{"fragment at root", `query { ...F } fragment F on Query { me { id } }`, false},
		{"invalid", `query {`, false},
		{"empty", ``, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"query": tt.query})
			if err != nil {
				t.Fatal(err)
			}
			if got := isPublicOperation(body); got != tt.expected {
				t.Errorf("isPublicOperation(%q) = %v, expected %v", tt.query, got, tt.expected)
			}
		})
	}

	if isPublicOperation([]byte(`not json with login`)) {
		t.Error("expected a non-JSON body to require a session")
	}
}
