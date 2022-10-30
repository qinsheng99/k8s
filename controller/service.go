package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"k8s-demo/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Service struct {
}

func (s *Service) Create(c *gin.Context) {
	_, _ = client.GetClient().
		CoreV1().
		Services(c.DefaultQuery("namespace", "default")).
		Create(context.TODO(), s.getService(c.Query("name"), c.Query("deployment"), 8080), metav1.CreateOptions{})
}

func (s *Service) getService(name, deployment string, port int32) *corev1.Service {
	var service = &corev1.Service{}
	service.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
	service.ObjectMeta = metav1.ObjectMeta{
		Name: name,
	}
	service.Spec = corev1.ServiceSpec{
		Selector: map[string]string{
			"app": deployment,
		},
		Ports: []corev1.ServicePort{
			{
				Name:       "http-port",
				Protocol:   corev1.ProtocolTCP,
				Port:       port,
				TargetPort: intstr.FromInt(int(port)),
			},
		},
		Type: corev1.ServiceTypeClusterIP,
	}
	return service
}
