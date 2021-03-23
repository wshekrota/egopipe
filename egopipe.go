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
import "time"

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
	Error   error
}

type Metrics struct {
    Bytes int
    Docs int
    Fields map[int]int
    Elapsed time.Duration
}


func main() {

	// Read config options
	//

	p, err := getConf()

	if err != nil {
		log.Fatalf("Egopipe config Unmarshal error: %v", err)
	}

	res := setLog()
	if res != nil {
		log.Fatalf("Egopipe config setLog error: %v", err)
		os.Exit(28)
	}

	var Hash map[string]interface{}
	c := make(chan *map[string]interface{})
	r := make(chan Result)
    totals := Metrics{}
    totals.Fields = make(map[int]int)
    
	// Read json from stdin passed from null logstash pipe
	//

	reader := bufio.NewReader(os.Stdin)

	for {
	    // Read from pipe
	    //
		slice, _ := (*reader).ReadBytes('\n')

		if err := json.Unmarshal(slice, &Hash); err != nil {
			fmt.Printf("Egopipe input Unmarshal error: %v", err)
		}
        now := time.Now().UTC()

		// Save datestamp for output stage
		//
		ds := strings.SplitN(Hash["@timestamp"].(string), "T", 2)[0]
		ds = strings.Replace(ds, "-", ".", 2)

		go yourpipecode(Hash, c) // stage 2
		log.Println(Hash["message"])
		pstg2map := <-c // return pointer to internal map

		go output(ds, p, pstg2map, r) // stage 3
		resp := <- r
		log.Println("response from output", resp.Message, err)
        if err != nil {
           os.Exit(3)
        }

        totals.Elapsed = time.Since(now)
        totals.Docs++
        totals.Bytes+=len(resp.Message)
        nfields := len(*pstg2map)
    	(totals.Fields)[nfields]++

        log.Println("metrics:",totals)
	}

}


func setLog() error {

    de := os.Mkdir("/var/log/logstash/egopipe", 0644)
    if de != nil {  // error would be file exists thats ok
    }
    file, err := os.OpenFile("/var/log/logstash/egopipe/egopipe.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    log.SetPrefix("/var/log/logstash/egopipe")
    log.SetOutput(file)
    return err
}


type Config struct {
	Target string
	Name   string
}


func getConf() (*Config, error) {

	conf := &Config{Target: "http://127.0.0.1:9200", Name: "egopipe"}

	file, err := ioutil.ReadFile("/etc/logstash/conf.d/egopipe.conf")

	if err != nil { // soft error
		fmt.Printf("Egopipe config Get file error #%v, Defaults used. ", err)
		return conf, err
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

//	    idx := strings.IndexRune(h["message"].(string),'{')   // json convert of message
//	    if idx>0 { json.Unmarshal([]byte((h["message"].(string))[idx:]),&h) }

	c <- &h // Although you write code here this line is required
}


/*

   stage 3
   Output to index in Elastic

*/

func output(dateof string, c *Config, hp *map[string]interface{}, r chan Result) {

	var s Result
	jbuf, err := json.Marshal(hp)
	if err != nil {
		s.Message = fmt.Sprintf("Egopipe input Marshal error: %v", err)
		s.Error = err
	} else {
		responseBody := bytes.NewBuffer(jbuf)
		url := fmt.Sprintf("%s/log-%s-%s/_doc/", c.Target, c.Name, dateof)
		resp, err := http.Post(url, "application/json", responseBody)
		if err != nil {
			s.Message = fmt.Sprintf("Egopipe output POST error #s writing Elastic index. error: %v", resp.Status, err)
			s.Error = err
		} else {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			s.Message = string(body)
			s.Error = err
		}
	}
	r <- s
}
