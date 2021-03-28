package egopipe

import "os"
import "log"



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

