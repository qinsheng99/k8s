package main

import (
	"k8s-demo/client"
	"log"
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}
}
