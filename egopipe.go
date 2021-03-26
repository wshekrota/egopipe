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
import "encoding/base64"
import "crypto/tls"
import "crypto/x509"

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
	Bytes   int
	Docs    int
	Fields  map[int]int
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
    var client *http.Client

	// Read cert in
	//
    if Secure := strings.HasPrefix(p["Target"],"https"); Secure {
		caCert, err := ioutil.ReadFile("cert.pem")
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Create a HTTPS client and supply the created CA pool
		//
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
		}
    } else {
		client = &http.Client{}
    }

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

		go output(client, ds, p, pstg2map, r) // stage 3
		resp := <-r
		log.Println("response from output", resp.Message, resp.Error)
		if err != nil {
			os.Exit(3)
		}

		totals.Elapsed = time.Since(now)
		totals.Docs++
		totals.Bytes += len(resp.Message)
		(totals.Fields)[len(*pstg2map)]++

		log.Println("metrics:", totals)
	}

}

func setLog() error {

	de := os.Mkdir("/var/log/logstash/egopipe", 0644)
	if de != nil { // error would be file exists thats ok
	}
	file, err := os.OpenFile("/var/log/logstash/egopipe/egopipe.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetPrefix("/var/log/logstash/egopipe")
	log.SetOutput(file)
	return err
}

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


func output(client *http.Client, dateof string, c map[string]string, hp *map[string]interface{}, r chan Result) {

	var s Result

	jbuf, err := json.Marshal(hp)
	if err != nil {
		s.Message = fmt.Sprintf("Egopipe input Marshal error: %v", err)
		s.Error = err
	} else {
		url := fmt.Sprintf("%s/log-%s-%s/_doc/", c["Target"], c["Name"], dateof)

		// post request
		//
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jbuf))
		if err != nil {
			s.Error = err
			s.Message = "error request="
		} else {

			/*

			   Basic Auth:  encode authentication in the header so it is not exposed

			*/
			if len(c["User"]) != 0 {
				// build authentication string
				//
				as := fmt.Sprintf("%s:%s", c["User"], c["Password"])
				enc := base64.StdEncoding.EncodeToString([]byte(as))
				auth := fmt.Sprintf("%s %s", "Basic", enc)

				// add to header
				//
				req.Header.Add("Authorization", auth)
			}

			req.Header.Add("Content-Type", "application/json")

			// do it
			//
			response, err := (*client).Do(req)
			if err != nil {
				s.Error = err
				s.Message = "error do="
			} else {
				// read the response
				//
				bites, err := ioutil.ReadAll(response.Body)
				if err != nil {
					s.Error = err
					s.Message = "error read="
				} else {
					s.Message = string(bites)
					s.Error = err
					defer response.Body.Close()
				}
			}
		}
	}
	r <- s
}
