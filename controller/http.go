package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	"k8s-demo/tools"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

type Server struct {
	Pod
	Secret
	Deployments
	NameSpace
	Project
	Resource
}

// NotParallel  非并行job
func (s *Server) NotParallel(c *gin.Context) {
	cli := client.GetClient()
	job := &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
		Spec: v1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test-job-container",
							Image:   "busybox",
							Command: []string{"sh", "-c"},
							Args:    []string{`echo hello;sleep 50;echo job`},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	j, err := cli.BatchV1().Jobs("default").Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, j)
}

func (s *Server) parallel(cli *kubernetes.Clientset, w http.ResponseWriter) {
	i := int32(0)
	t := int32(1800)
	job := &v1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
		Spec: v1.JobSpec{
			BackoffLimit:            &i,
			TTLSecondsAfterFinished: &t,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "test-job",
							Image:   "busybox",
							Command: []string{"sh", "-c"},
							Args:    []string{`echo hello;sleep 30;echo job`},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	_, err := cli.BatchV1().Jobs("default").Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return
	}
}

func (s *Server) DeleteJob(c *gin.Context) {
	cli := client.GetClient()
	name := c.Query("name")
	namespace := c.DefaultQuery("namespace", "default")
	err := cli.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "成功")
}
