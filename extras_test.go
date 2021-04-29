package main

import "testing"
import "strings"
import "fmt"

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

	h := make(map[string]interface{})
	tests := []map[string]string{
		{"message": "one two three", "pattern": `%{(?P<A>[\S]+):field1} %{(?P<B>[a-z]+):field2}`},
		{"message": "one 314159 three", "pattern": `%{(?P<A>[a-z]+):field1} %{(?P<B>[0-9]+):field2}`},
		{"message": "55.3.244.1 GET /index.html 15824 0.043", "pattern": "%{(?P<A>[0-9.]+):ip} %{(?P<B>[A-Z]+):method} %{(?P<C>[a-zA-Z0-9.-_/]+):url} %{(?P<D>[0-9]+):bytes}"}}

	// For each test
	for j := 0; j < len(tests); j++ {
		// Setup map to fake grok use
		h["message"] = tests[j]["message"]
		// Call grok with each test
		grok(&h, tests[j]["pattern"])
		fmt.Println(h)
		captures := strings.Split(h["message"].(string), " ")
		patterns := strings.Split(tests[j]["pattern"], " ")
		for i := 0; i < len(patterns); i++ {
			// semantic is field
			_, semantic := parsePattern(patterns[i])
			if semantic != "" {
				// Does newfield exist?
				if _, ok := h[semantic]; !ok {
					t.Error("field not created.")
				}
				// Is it's value appropriate?
				if h[semantic] != captures[i] {
					t.Error("field not created.")
				}
			}
		}
	}
}
