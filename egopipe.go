package egopipe

import "io/ioutil"
import "bufio"
import "net/http"
import "log"
import "os"
import "encoding/json"
//import "bytes"
import "fmt"
import "strings"
import "time"
//import "encoding/base64"
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

