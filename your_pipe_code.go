package main

//import "log"
import "encoding/json"

/*

   stage 2
   Where your filter code runs. The doc object is the h map

*/

// yourPipeCode objective is for a user to code the pipe transform stage in golang.
// That code would exist here and a ref to that completed map goes back in channel.
//
func yourPipeCode(h map[string]interface{}, c chan *map[string]interface{}) {

	// h is the hash representing your docs (which are a collection of fields)
	// keys are fieldnames
	// value is interface{} and must be asserted

	// Access log.file.path value
	//
	p := dotField(h, "log.file.path").(string)
	
	if p == "/var/log/gitlab/gitaly/current" {
		json.Unmarshal([]byte(h["message"].(string)), &h)
		addTags(&h, []string{"DecodedJsonToFields"})
		oniguruma("justtime", "time", &h, `(T[0-9:Z]+$)`)
	}
	if p == "/var/log/syslog" {
		grok(&h, `%{(?P<a>\S+):field1} %{(?P<a>\S+):field2} %{(?P<a>\S+):} %{(?P<a>\S+):field4} %{(?P<a>\S+):field5}`)
	}

	c <- &h // Although you write code here this line is required
}
