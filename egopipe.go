// Wanting simplicity in elastic data ingestion I created this so go could be used instead of plugins.
package main

import "io/ioutil"
import "io"
import "bufio"
import "net/http"
import "log"
import "os"
import "encoding/json"
import "strings"
import "time"
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

type Metrics struct {
	Bytes   int
	Docs    int
	Fields  map[int]int
	Elapsed time.Duration
}

type Ret struct {
	slice []byte
	err   error
}

// The main contols the pipe in an ETL fashion. SSL and authentication is also handled for elastic host.
// Read stdin, amend the doc, output to index
// Delivers end to end statistics in log so you can assess response time.
func main() {

	// Read config options
	//

	p := getConf()

	err := setLog()
	if err != nil {
		log.Fatalf("Egopipe config setLog error: %v", err)
		os.Exit(28)
	}

	c := make(chan *map[string]interface{})
	r := make(chan Result)
	totals := Metrics{}
	totals.Fields = make(map[int]int)
	var client *http.Client

	// Read cert in
	// is it a secure transaction?
	//
	if Secure := strings.HasPrefix(p["Target"], "https"); Secure {
		caCert, err := ioutil.ReadFile(PIPE_DIR + "/ego/cert.pem")
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

		// or it is unsecure transaction ie http
		//
	} else {
		client = &http.Client{}
	}

	// Read json from stdin passed from null logstash pipe
	//

	log.Println("Start pipe.")
	lineFromPipe := make(chan Ret, 1)  // buffered
	reader := bufio.NewScanner(os.Stdin)
	var retData Ret

PipeLoop:
	for {
		go func(x chan Ret) {
			reader.Scan()
			x <- Ret{reader.Bytes(), reader.Err()}
		}(lineFromPipe)

		if retData.err == io.EOF {
			break PipeLoop
		}

		// Alternative - pipe end w/ timeout or EOF
		//
		select {
		case retData = <-lineFromPipe:

			log.Println("New line.", string(retData.slice), retData.err)
			now := time.Now().UTC()
			Hash := map[string]interface{}{}

			if err := json.Unmarshal(retData.slice, &Hash); err != nil {
				log.Printf("Egopipe input Unmarshal error: %v", err)
				continue PipeLoop
			}

			go yourPipeCode(Hash, c) // stage 2

			// return pointer to internal map
			// if channel not returned will block here
			//
			pstg2map := <-c

			go output(client, p, pstg2map, r) // stage 3
			resp := <-r

			log.Println("response from output", resp.Message, resp.Error)
			if resp.Error != nil {
				break PipeLoop
			}

			totals.Elapsed = time.Since(now)
			totals.Docs++
			totals.Bytes += len(resp.Message)
			(totals.Fields)[len(*pstg2map)]++

			log.Println("metrics:", totals)
		case <-time.After(1 * time.Second):
			break PipeLoop
		}
	}
	log.Printf("Pipe exit.")

}
