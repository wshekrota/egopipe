package main

import "testing"
import "reflect"


// For given config inputs compare the new layered map to expected result
// Assignment of read map arrays {{"x":"y"}....} or file input then layers with defaults and compared to desired results
func TestConf(t *testing.T) {

	// in is what is read ... nothing, just Target, all
	in := []map[string]string{		{}, 
									{"Target": "http://172.17.0.2:9200"}, 
									{"Target": "http://172.17.0.2:9200","Name":"testname","User":"fred","Password":"bla"} }

	desired := []map[string]string{	{"Target": "http://127.0.0.1:9200", "Name": "egopipe", "User": "", "Password": ""}, 
									{"Target": "http://172.17.0.2:9200", "Name": "egopipe", "User": "", "Password": ""}, 
									{"Target": "http://172.17.0.2:9200","Name":"testname","User":"fred","Password":"bla"} }

	for i := 0; i < len(in); i++ {
		new_map := configLayer(in[i])
		if !reflect.DeepEqual(new_map,desired[i]) {
			t.Errorf("Unexpected result testcase %d", i)
		}
	}

}
