package main

import "os"
import "log"


// setLog postures the setup to locate the logs in logstash pipe directory.
// Set permissions properly
func setLog() error {

	de := os.Mkdir(LogsHere, 0644)
	if de != nil { // error would be file exists thats ok
	}
	file, err := os.OpenFile(LogsHere + LogName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetPrefix(LogsHere)
	log.SetOutput(file)
	return err
}

