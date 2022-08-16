package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var (
	k8sConfig *rest.Config
	k8sClient *kubernetes.Clientset
)

func Init() (err error) {
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", "~/.kube/config")
	if err != nil {
		log.Println(err)
		return
	}

	k8sClient, err = kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Println(err)
		return
	}
	return nil
}

func GetClient() *kubernetes.Clientset {
	return k8sClient
}
