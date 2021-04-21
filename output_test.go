package main

import "testing"
import "net/http"
import "fmt"

// Will only function properly if a container is functional for elastic at 172.17.0.2
func TestOutput(t *testing.T) {

	client := &http.Client{}
	p := make(map[string]string)
	p["Target"] = "http://172.17.0.2:9200"
	p["Name"] = "egopipe"
	m := make(map[string]interface{})
	m["@timestamp"] = "1776-07-04T08:58:21.976Z"
	m["message"] = "this is a test"
	r := make(chan Result)
	go output(client, p, &m, r)
	got := <-r
	if got.Error != nil {
	   fmt.Println(got.Error)
	}
	   fmt.Println(got.Message)
}

