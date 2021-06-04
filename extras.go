package main

import "strings"
import "regexp"

/*

  Tools for use in transform stage.
  Thinking about grok, patterns and ECS in future.

*/

// grok - based on %{syntax:semantic} format
// patterns defined will be used as syntax or regex
func grok(map_ptr *map[string]interface{}, pattern string) bool {

	var syntaxes string
	var semantics []string
	patterns := strings.Split(pattern, " ")
	for i := 0; i < len(patterns); i++ {
		syntax, semantic := parsePattern(patterns[i])
		syntaxes += syntax + " "
		semantics = append(semantics, semantic)
	}
	reg := regexp.MustCompile(syntaxes)
	results := reg.FindString((*map_ptr)["message"].(string))
	res := strings.Split(results, " ")
	for j := 0; j < len(res)-1; j++ {
		if semantics[j] != "" {
			(*map_ptr)[semantics[j]] = res[j]
		}
	}
	return len(results) > 0

}

// oniguruma - based on a regex create/load a field
// Return was the field created?
func oniguruma(newfield string, field string, map_ptr *map[string]interface{}, regx string) bool {

	if _, ok := (*map_ptr)[field]; ok {
		reg, err := regexp.Compile(regx)
		if err != nil {
			return false
		}
		resp := reg.FindStringSubmatch((*map_ptr)[field].(string))[1]
		(*map_ptr)[newfield] = string(resp)
		return true
	} else {
		// Search field did not exist
		return false
	}

}

// addTags - Add tag(s) to tags field for this doc.
// Pointer this doc and string(s) to add to tags.
func addTags(doc *map[string]interface{}, s []string) {

	if _, ok := (*doc)["tags"]; ok {
		(*doc)["tags"] = append((*doc)["tags"].([]interface{}), s)
	} else {
		(*doc)["tags"] = s
	}
}

// Legacy fields with '.'
// dotField - If not "log.file.path": value then breakout map names with '.' in them to their component.
// ie 'log.file.path' is map["log":map["file":map["path":string]]] or a string
func dotField(h map[string]interface{}, a string) interface{} {

	var r interface{}
	// If key exists
	//
	if _, ok := h[a]; ok {
		r = h[a]
		return r
	}
	maps := strings.Split(a, ".")
	// Otherwise data is submapped
	//
	for i := 0; i < len(maps); i++ {
		switch h[maps[i]].(type) {
		case string, int, int8, int16, int32, int64, bool, byte, float32, float64:
			r = h[maps[i]]
			break
		default:
			h = h[maps[i]].(map[string]interface{})
		}

	}
	return r

}

// Parse the element %{syntax:symantic} to syntax,symantic
//
func parsePattern(pat_def string) (string, string) {

	reg := regexp.MustCompile(`%{(?P<Syntax>.*):(?P<Symantic>.*)}`)
	result := reg.FindStringSubmatch(pat_def)
	if result == nil {
		return "", ""
	}
	return result[1], result[2]

}
