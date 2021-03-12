package main

import "io/ioutil"
import "bufio"
import "net/http"
import "log"
import "os"
import "encoding/json"
import "bytes"

func main() {

	// Read json from stdin passed from null logstash pipe
	//

	var Hash map[string]interface{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if err := json.Unmarshal([]byte(scanner.Text()), &Hash); err != nil {
			break
		}
		log.Println("stage1 field message=", Hash["message"])

		c := make(chan *map[string]interface{})
		go yourpipecode(Hash, c)     // stage 2
		pstg2map := <-c
//		log.Println("stage2", (*pstg2map)["test"])
		s := make(chan string)
		r := make(chan string)
		go output(pstg2map, s, r)    // stage 3
		log.Println("stage3 index write request field json=", <-s)
		log.Println("response from POST", <-r)

	}

}


func yourpipecode(h map[string]interface{}, c chan *map[string]interface{}) {

    // h is the hash representing your docs
	h["test"] = 31415    // example field add
	
	c <- &h  // Although you write code here this line is required 
}


func output(hp *map[string]interface{}, s chan string, r chan string) {

	jbuf, _ := json.Marshal(*hp)
	s <- string(jbuf)
	responseBody := bytes.NewBuffer(jbuf)
	resp, err := http.Post("http://192.168.1.43:9200/log-tester-2021.03.08/_doc/", "application/json", responseBody)
	if err != nil {
		log.Println("HTTP POST error writing Elastic index. error=", err)
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		r <- string(body)
	}
}
