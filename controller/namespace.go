package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/tools"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NameSpace struct {
}

func (n *NameSpace) GetNamespace(c *gin.Context) {
	cli := client.GetClient()
	list, err := cli.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}
	var result [][]string
	for _, v := range list.Items {
		result = append(result, []string{v.Namespace, v.Name})
	}
	tools.Success(c, result)
	return
}

func (n *NameSpace) CreateNamespace(c *gin.Context) {
	namespace := c.Query("namespace")
	cli := client.GetClient()

	newnamespace := &corev1.Namespace{}
	newnamespace.TypeMeta = metav1.TypeMeta{
		Kind:       "Namespace",
		APIVersion: "v1",
	}
	newnamespace.ObjectMeta = metav1.ObjectMeta{
		Name: namespace,
		Labels: map[string]string{
			"name": namespace,
		},
	}

	name, err := cli.CoreV1().Namespaces().Create(context.TODO(), newnamespace, metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, name)
}

func (n *NameSpace) DeleteNamespace(c *gin.Context) {
	namespace := c.Query("namespace")
	cli := client.GetClient()

	err := cli.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "成功")
}
