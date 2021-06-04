package main

import "net/http"
import "encoding/json"
import "fmt"
import "bytes"
import "encoding/base64"
import "io/ioutil"
import "strings"

/*

   stage 3
   Output to index in Elastic.
   Runs as a goroutine to concurrently manage the security and output of the doc.

*/

type Result struct {
	Message string
	Error   error
}

// Function: output manages the docs output to an index.
// Passed: client structure, config map, ref to stage2 map, channel to return struct
func output(client *http.Client, c map[string]string, hp *map[string]interface{}, r chan Result) {

	var s Result

	jbuf, err := json.Marshal(hp)
	if err != nil {
		s.Message = fmt.Sprintf("Egopipe input Marshal error: %v", err)
		s.Error = err
	} else {
		// reformat date for indexname
		//
		dateof := strings.SplitN((*hp)["@timestamp"].(string), "T", 2)[0]
		dateof = strings.Replace(dateof, "-", ".", 2)
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
