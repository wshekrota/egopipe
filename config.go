package egopipe

import "encoding/json"
import "fmt"
import "io/ioutil"



func getConf() (map[string]string, error) {

	var n map[string]interface{}
	// set known defaults here
	//
	m := map[string]string{"Target": "http://127.0.0.1:9200", "Name": "egopipe", "User": "", "Password": ""}

	file, err := ioutil.ReadFile("/etc/logstash/conf.d/egopipe.conf")

	if err != nil { // soft error
		fmt.Printf("Egopipe config Get file error #%v, Defaults used. ", err)
		return m, err
	} else {

		err = json.Unmarshal(file, &n)
		if err != nil {
			return m, err
		}
		for key, val := range n {
			m[key] = val.(string)
		}

	}
	return m, nil
}

