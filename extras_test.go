package main

import "testing"
import "strings"

// Test dotField functionality
// First submaps then key:value
func TestExtrasDot(t *testing.T) {

	// test 1
	z := map[string]interface{}{"path": "abc"}
	y := map[string]interface{}{"file": z}
	v := map[string]interface{}{"log": y}

	if "abc" != dotField(v, "log.file.path") {
		t.Error("Submap value did not compare.")
	}
	// test2
	u := map[string]interface{}{"log.file.path": "abc"}
	if "abc" != dotField(u, "log.file.path") {
		t.Error("key/value did not compare.")
	}

}

func TestExtrasGrok(t *testing.T) {

	tests := []map[string]string{
		{"message": "one two three", "pattern": `%{(?P<A>[\S]+):field1} %{(?P<B>[a-z]+):field2}`},
		{"message": "one 314159 three", "pattern": `%{(?P<A>[a-z]+):field1} %{(?P<B>[0-9]+):field2}`},
		{"message": "55.3.244.1 GET /index.html 15824 0.043", "pattern": "%{(?P<A>[0-9.]+):ip} %{(?P<B>[A-Z]+):method} %{(?P<C>[a-zA-Z0-9.-_/]+):url} %{(?P<D>[0-9]+):bytes}"}}

	// For each test
	for j := 0; j < len(tests); j++ {
		h := make(map[string]interface{})
		// Setup map to fake grok use
		h["message"] = tests[j]["message"]
		// Call grok with each test
		if !grok(&h, tests[j]["pattern"]) {
			t.Error("grok failed.")
		}
		captures := strings.Split(h["message"].(string), " ")
		patterns := strings.Split(tests[j]["pattern"], " ")
		for i := 0; i < len(patterns); i++ {
			// semantic is field
			_, semantic := parsePattern(patterns[i])
			// Skip if no semantic
			if semantic != "" {
				// Does newfield exist?
				if _, ok := h[semantic]; !ok {
					t.Error("field not created.", semantic)
				}
				// Is it's value appropriate?
				if h[semantic] != captures[i] {
					t.Error("field incorrect value.")
				}
			}
		}
	}
}

func TestExtrasOniguruma(t *testing.T) {

	tests := []map[string]string{
		{"text": "blabla one=1 two=2 three=3", "regex": `(?:two=)([0-9]+)`, "expected": "2"},
		{"text": "two=2 three=3", "regex": `(?:two=)([0-9]+)`, "expected": "2"},
		{"text": "lalala one=1 two=2 three=3", "regex": `(two=[0-9]+)`, "expected": "two=2"},
		{"text": "two=2 three=3", "regex": `(two=[0-9]+)`, "expected": "two=2"},
		{"text": "wshek/repo.git,blablabla", "regex": `([a-z/]+)(?:.git*)`, "expected": "wshek/repo"},
		{"text": "wshek/repo", "regex": `([a-zA-Z0-9_-]+)(?:/*)`, "expected": "wshek"}}

	for i, _ := range tests {
		h := make(map[string]interface{})
		h["message"] = tests[i]["text"]
		if !oniguruma("test", "message", &h, tests[i]["regex"]) {
			t.Error("Oniguruma failed.")
		}
		if h["test"] != tests[i]["expected"] {
			t.Error("Result oniguruma field incorrect.", h["test"])
		}
	}
}
