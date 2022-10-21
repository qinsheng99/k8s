package main

import (
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/route"
	"log"
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	route.Route(r)

	err = r.Run(":8088")
	if err != nil {
		log.Fatal(err)
	}

}
