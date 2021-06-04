package main

import "encoding/json"
import "log"
import "io/ioutil"

// configLayer manages config file for pipe setup.
// Overlay read file on top of defaults resulting in at least read settings.
func configLayer(n map[string]string) map[string]string {

	// set known defaults here .. insecure
	//
	m := map[string]string{
		"Target":   "http://127.0.0.1:9200",
		"Name":     "egopipe",
		"User":     "",
		"Password": ""}

	// map overlay defaults
	//
	for key, val := range n {
		m[key] = val
	}

	return m
}

// configRead reads local json config file.
// Then it returns a decoded map.
func configRead() (map[string]string, error) {

	var m, n map[string]string // zero value

	// Read config file from install directory
	//
	file, err := ioutil.ReadFile(PIPE_DIR + "/ego/" + CONFIG_NAME)

	if err != nil { // soft error
		log.Printf("Egopipe config Get file error #%v, Defaults used. ", err)
		return m, err

	}

	// json to map
	//
	err = json.Unmarshal(file, &n)
	if err != nil {
		return m, err
	}

	return n, nil

}

// getConf call to get config info from file and layer it with defaults
func getConf() map[string]string {

	// Returns zero value or decode
	//
	read, _ := configRead()

	// Returns overlay
	return configLayer(read)

}
