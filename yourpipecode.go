package egopipe



/*

   stage 2
   Where your filter code runs. The doc object is the h map

*/

func yourpipecode(h map[string]interface{}, c chan *map[string]interface{}) {

	// h is the hash representing your docs
	// keys are fields
	// value is interface{} and must be asserted

	//	  h["test"] = 31415    // example field add
	//    delete(h,"key")      // delete field
	//    _, found := h["key"] // true or false does this field exist?

	//	    idx := strings.IndexRune(h["message"].(string),'{')   // json convert of message
	//	    if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }

	c <- &h // Although you write code here this line is required
}
