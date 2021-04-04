package main

import "encoding/json"
import "log"
import "io/ioutil"



// Manage config file for pipe setup.
// Where is elasticsearch, core name for index and who authenticates?
func getConf() (map[string]string, error) {

	var n map[string]interface{}
	
	// set known defaults here .. insecure
	//
	m := map[string]string{
			"Target": "http://127.0.0.1:9200",
			"Name": "egopipe",
			"User": "",
			"Password": ""}

    // Read config file from install directory
    //
	file, err := ioutil.ReadFile(PipeDir + "/ego/" + ConfigName)

	if err != nil { // soft error
		log.Printf("Egopipe config Get file error #%v, Defaults used. ", err)
		return m, err

	} else {

        // json to map
        //
		err = json.Unmarshal(file, &n)
		if err != nil {
			return m, err
		}

		// map overlay defaults
		//
		for key, val := range n {
			m[key] = val.(string)
		}

	}
	return m, nil
}

