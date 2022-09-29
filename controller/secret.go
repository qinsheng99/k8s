package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/tools"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Secret struct {
}

func (Secret) getSecret(secretName, namespace string) *corev1.Secret {
	newsecret := &corev1.Secret{}
	newsecret.Type = corev1.SecretTypeOpaque
	newsecret.TypeMeta = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Secret",
	}

	newsecret.ObjectMeta = metav1.ObjectMeta{
		Name:      secretName,
		Namespace: namespace,
	}

	newsecret.Data = map[string][]byte{
		"db-user":     []byte("zjm="),
		"db-password": []byte("zjm123=="),
	}
	return newsecret
}

func (s *Secret) CreateSecret(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")

	secret, err := cli.CoreV1().Secrets(namespace).Create(context.TODO(), s.getSecret(secretName, namespace), metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}
	tools.Success(c, secret)
}

func (s *Secret) DeleteSecret(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")

	err := cli.CoreV1().Secrets(namespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "成功")
}
