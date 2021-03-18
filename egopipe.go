package main

import "io/ioutil"
import "bufio"
import "net/http"
import "log"
import "os"
import "encoding/json"
import "bytes"
import "fmt"
import "strings"

/*
   Author: Walt Shekrota wshekrota@icloud.com
   Name: egopipe

   Description:
   Define a pipeline with 3 stages.
   Launch it from logstash via pipe output plugin.
   ETL accepts input/transforms/indexes
   Logstash is just an empty conduit.

*/

type Result struct {
    Message string
    Error error
}

func main() {

	// Read config options
	//

	p, err := getConf()

    if err != nil {
		log.Fatalf("Egopipe config Unmarshal error: %v", err)
	}

	// Read json from stdin passed from null logstash pipe
	//

	var Hash map[string]interface{}
	c := make(chan *map[string]interface{})
	r := make(chan []byte)

	reader := bufio.NewReader(os.Stdin)

	for {
		slice, _ := (*reader).ReadBytes('\n')
		log.Println(">>line", string(slice))
		if err := json.Unmarshal(slice, &Hash); err != nil {
			log.Fatalf("Egopipe input Unmarshal error: %v", err)
		}

		ds := strings.SplitN(Hash["@timestamp"].(string), "T", 2)[0]
		ds = strings.Replace(ds, "-", ".", 2)

		go yourpipecode(Hash, c) // stage 2
		log.Println(Hash["message"])
		pstg2map := <-c // return pointer to internal map

		go output(ds, p, pstg2map, r) // stage 3
		log.Println("response from POST", string(<-r))

	}

}

type Config struct {
	Target string
	Name   string
}

func getConf() (*Config, error) {

	conf := &Config{Target: "http://127.0.0.1:9200", Name: "egopipe"}

	file, err := ioutil.ReadFile("/etc/logstash/conf.d/egopipe.conf")

	if err != nil {    // soft error
		log.Printf("Egopipe config Get file error #%v, Defaults used. ", err)
	} else {
		err = json.Unmarshal(file, conf)
		if err != nil {
			return conf, err
		}
	}
	return conf, nil
}

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

	//    idx := strings.IndexRune(h["message"].(string),'{')   // json convert of message
	//    if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }

	c <- &h // Although you write code here this line is required
}

/*

   stage 3
   Output to index in Elastic

*/

func output(dateof string, c *Config, hp *map[string]interface{}, r chan []byte) {

	jbuf, err := json.Marshal(hp)
	if err != nil {
		log.Fatalf("Egopipe input Marshal error: %v", err)
	}
	responseBody := bytes.NewBuffer(jbuf)
	url := fmt.Sprintf("%s/log-%s-%s/_doc/", c.Target, c.Name, dateof)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		log.Fatalf("Egopipe output POST error writing Elastic index. error=", err)
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		r <- body
	}
}
