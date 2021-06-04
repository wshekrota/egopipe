package main

//import "encoding/json"

/*

   stage 2
   Where your filter code runs. The doc object is the h map

*/

// your_pipe_code objective is for a user to code the pipe transform stage in golang.
// That code would exist here and a ref to that completed map goes back in channel.
//
func yourPipeCode(h map[string]interface{}, c chan *map[string]interface{}) {

	// h is the hash representing your docs (which are a collection of fields)
	// keys are fieldnames
	// value is interface{} and must be asserted




	c <- &h // Although you write code here this line is required
}
