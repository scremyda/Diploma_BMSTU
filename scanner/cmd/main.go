package main

import "log"

func main() {
	_, err := ReadConf(ReadArgs())
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}

}
