package main


//import "strings"
import "encoding/json"



/*

   stage 2
   Where your filter code runs. The doc object is the h map

*/


// The objective of this package was a user coding the transform stage in golang.
// That code would exist here and a ref to that completed map goes back in channel.
//
func yourpipecode(h map[string]interface{}, c chan *map[string]interface{}) {

	// h is the hash representing your docs (which are a collection of fields)
	// keys are fieldnames
	// value is interface{} and must be asserted

    //	  	h["test"] = 31415    // example field add
	//    	delete(h,"key")      // delete field

	//    	_, found := h["key"] // true or false does this field exist?

			// This will decode a message field that is prefixed by non json
			//
    //	  	idx := strings.IndexRune(h["message"].(string),'{')   // json convert of message
    //	  	if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }

			// I know gitaly is complete json so this will blindly decode a non prefixed message field
			//
    		json.Unmarshal([]byte(h["message"].(string)),&h)

	c <- &h // Although you write code here this line is required
}
