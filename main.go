package main

import (
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/controller"
	"k8s-demo/informer"
	"k8s-demo/route"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
)

var (
	resource schema.GroupVersionResource
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	s := controller.Server{}
	route.Route(r, s)

	resource, err, _ = s.GetCrd()
	if err != nil {
		log.Fatalln(err)
	}

	listen := informer.NewListen(client.GetClient(), client.GetK8sConfig(), client.GetDyna(), resource)
	go listen.ListenResource()

	err = r.Run(":8088")
	if err != nil {
		log.Fatal(err)
	}

}
