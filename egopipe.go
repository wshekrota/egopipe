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

   Define a pipeline with 3 stages.
   Launch it from logstash via pipe output plugin.
   ETL accepts input/transforms/indexes
   Logstash is just an empty conduit.

 */

func main() {

    // Read config options
    //

    p, _ := getConf()

	// Read json from stdin passed from null logstash pipe
	//

	var Hash map[string]interface{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if err := json.Unmarshal([]byte(scanner.Text()), &Hash); err != nil {
			break
		}
		log.Println("xstage1 field message=", Hash["message"])
        ds := strings.SplitN(Hash["@timestamp"].(string),"T",2)[0]
        log.Println("xramone",ds)
		c := make(chan *map[string]interface{})
		go yourpipecode(Hash, c)            // stage 2
		pstg2map := <-c                     // return pointer to internal map
		log.Println("xstage2", (*pstg2map)["test"])
		s := make(chan string)
		r := make(chan string)
		log.Println("hello",ds, p.Target,p.Name, pstg2map, s, r)    // stage 3
		go output(ds, p.Target,p.Name, pstg2map, s, r)    // stage 3
		log.Println("stage3 index write request field json=", <-s)
		log.Println("response from POST", <-r)

	}

}


type Config struct {
    Target string
    Name string
}


func getConf() (*Config, error) {
 
    conf := &Config{Target:"http://127.0.0.1:9200",Name:"egopipe"}
//    fmt.Println("Defaults",conf.Target,conf.Name)
    file, err := ioutil.ReadFile("egopipe.conf")
//    fmt.Println("read",string(file))
    if err != nil {
        log.Printf("my.conf.Get err   #%v ", err)
        return nil, err
    }
    err = json.Unmarshal(file, conf)
    fmt.Println("decode",conf,err)
    if err != nil {
        log.Fatalf("Unmarshal error: %v", err)
        return nil, err
    }
    return conf, nil
}


/*

       stage 2
       Where your filter code runs. The doc object is the h map

 */


func yourpipecode(h map[string]interface{}, c chan *map[string]interface{}) {

    // h is the hash representing your docs
	h["test"] = 31415    // example field add
	
	c <- &h  // Although you write code here this line is required 
}


/*

       stage 3
       Output to index in Elastic

 */

func output(dateof string, target string, name string, hp *map[string]interface{}, s chan string, r chan string) {

		log.Println("json garbage")
	jbuf, err := json.Marshal(hp)
    s <- string(jbuf)
	responseBody := bytes.NewBuffer(jbuf)
//	url := fmt.Sprintf("%s/log-%s-%s/_doc/", target, name, dateof)
resp, err := http.Post("http://192.168.1.43:9200/log-tester-2021.03.08/_doc/", "application/json", responseBody)
//	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		log.Println("HTTP POST error writing Elastic index. error=", err)
	} else {
	    log.Println("POST response=",resp)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		r <- string(body)
	}
}
