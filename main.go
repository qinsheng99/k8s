package main

import (
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/controller"
	"log"
)

func main() {
	err := client.Init()
	if err != nil {
		log.Fatalln(err)
	}

	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	s := controller.Server{}

	r.GET("/delete-job", s.DeleteJob)

	r.GET("/get-pod", s.Pod.GetPod)
	r.GET("/create-pod", s.Pod.CreatePod)
	r.GET("/update-pod", s.Pod.UpdatePod)
	r.GET("/delete-pod", s.Pod.DeletePod)
	r.GET("/get-pod-status", s.Pod.PodStatus)
	r.GET("/get-pod-log", s.Pod.PodLog)

	r.GET("/create-secret", s.Secret.CreateSecret)
	r.GET("/delete-secret", s.Secret.DeleteSecret)

	r.GET("/get-namespace", s.NameSpace.GetNamespace)
	r.GET("/create-namespace", s.NameSpace.CreateNamespace)
	r.GET("/delete-namespace", s.NameSpace.CreateNamespace)

	r.GET("/get-deployment", s.Deployments.GetDeployments)
	r.GET("/delete-deployment", s.Deployments.DeleteDeployments)
	r.GET("/create-deployment", s.Deployments.CreateDeployments)

	r.GET("/crd-list", s.Resource.Get)
	r.GET("/get-crd", s.Resource.GetCrd)
	r.GET("/create-crd-source", s.Resource.CreateResourceCrd)
	//r.GET("/update-crd-source-status", s.Resource.UpdateResourceCrdStatus)
	r.GET("/update-crd-source", s.Resource.UpdateResourceCrd)

	//r.GET("/demo", s.Resource.Demo)

	r.GET("/not-parallel", s.NotParallel)

	err = r.Run(":8088")
	if err != nil {
		log.Fatal(err)
	}

}
