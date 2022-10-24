package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/tools"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployments struct {
}

func (d *Deployments) GetDeployments(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	name := c.Query("name")
	if len(name) == 0 {
		list, err := cli.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			tools.Failure(c, err)
			return
		}
		var result []map[string]string
		for _, item := range list.Items {
			result = append(result, map[string]string{"namespace": item.Namespace, "name": item.Name})
		}
		tools.Success(c, result)
		return
	}

	deployment, err := cli.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, deployment)
}

func (d *Deployments) DeleteDeployments(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	name := c.Query("name")
	err := cli.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "成功")
}

func (d *Deployments) int32ptr(i int32) *int32 {
	return &i
}

func (d *Deployments) getDeploymentConf(name, namespace, secretName string) *v1.Deployment {
	newdeployment := &v1.Deployment{}
	newdeployment.TypeMeta = metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "v1",
	}

	newdeployment.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	newdeployment.Spec = v1.DeploymentSpec{
		Replicas: d.int32ptr(2),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						ImagePullPolicy: corev1.PullAlways,
						Name:            "test-server",
						Image:           "ccr.ccs.tencentyun.com/kugo/demo:v5",
						Env: []corev1.EnvVar{
							{
								Name: "DB_USER",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
										Key:                  "db-user",
									},
								},
							},
							{
								Name: "DB_PWD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
										Key:                  "db-password",
									},
								},
							},
						},
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": name,
				},
			},
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": name,
			},
		},
	}

	return newdeployment
}

func (d *Deployments) CreateDeployments(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")
	name := c.Query("name")
	deployment, err := cli.AppsV1().Deployments(namespace).Create(context.TODO(), d.getDeploymentConf(name, namespace, secretName), metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, deployment)
}

func (d *Deployments) UpdateDeployments(c *gin.Context) {
	cli := client.GetClient()
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")
	name := c.Query("name")
	up := d.getDeploymentConf(name, namespace, secretName)
	up.Spec.Replicas = d.int32ptr(3)
	deployment, err := cli.AppsV1().Deployments(namespace).Update(context.TODO(), up, metav1.UpdateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, deployment)
}
