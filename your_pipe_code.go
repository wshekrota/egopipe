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

    //	  	h["test"] = 31415    // example field add
	//    	delete(h,"key")      // delete field

	//      See README.md for more code suggestions.	



	// Remove these lines when you code your solution it was a testcase.
	json.Unmarshal([]byte(h["message"].(string)),&h)

//    if _, ok := h["tags"]; ok {   // if you checked for existence of "tags"
    
    // Access log.file.path value
    //
    x := h["log"].(map[string]interface{})  // assert value is map
    y := x["file"].(map[string]interface{}) // assert value is map
    z := y["path"].(string)					// finally assert value is string
    if z == "/var/log/gitlab/gitaly/current" {
         h["tags"] = append(h["tags"].([]interface{}),"DecodedJsonToFields")
    }

	c <- &h // Although you write code here this line is required
}
