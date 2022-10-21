package controller

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"k8s-demo/client"
	"k8s-demo/tools"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strconv"
	"strings"
	"time"
)

type Pod struct {
}

func (Pod) getPodConf(secretName, namespace string) *corev1.Pod {
	newPod := &corev1.Pod{}
	newPod.TypeMeta = metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	name := "test" + strconv.Itoa(int(time.Now().Unix()))
	newPod.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	newPod.Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "test-server",
				Image: "ccr.ccs.tencentyun.com/kugo/demo:v5",
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
				ImagePullPolicy: corev1.PullAlways,
				Resources:       corev1.ResourceRequirements{},
				//LivenessProbe: &corev1.Probe{
				//	ProbeHandler: corev1.ProbeHandler{
				//		HTTPGet: &corev1.HTTPGetAction{
				//			Path: "callback/" + namespace + "/" + name,
				//			Port: intstr.IntOrString{
				//				Type:   intstr.Int,
				//				IntVal: 8080,
				//			},
				//		},
				//	},
				//	InitialDelaySeconds: 5,  //Pod容器启动多少时间后开始检测
				//	PeriodSeconds:       10, //探测间隔时间
				//	TimeoutSeconds:      3,  //超时时间
				//},
				//Lifecycle: &corev1.Lifecycle{
				//	PostStart: &corev1.LifecycleHandler{
				//		HTTPGet: &corev1.HTTPGetAction{
				//			Path: "callback/" + namespace + "/" + name,
				//			Port: intstr.IntOrString{
				//				Type:   intstr.Int,
				//				IntVal: 8080,
				//			},
				//		},
				//	},
				//},
			},
		},
		RestartPolicy: corev1.RestartPolicyNever,
	}
	return newPod
}

func (p *Pod) CreatePod(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")
	cli := client.GetClient()

	pods, err := cli.CoreV1().Pods(namespace).Create(context.TODO(), p.getPodConf(secretName, namespace), metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	go p.watch(pods.Name, pods.Namespace)

	tools.Success(c, pods)
}

func (p *Pod) UpdatePod(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	secretName := c.Query("secret")
	cli := client.GetClient()

	pods, err := cli.CoreV1().Pods(namespace).Update(context.TODO(), p.getPodConf(secretName, namespace), metav1.UpdateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, pods)
}

func (p *Pod) PodStatus(c *gin.Context) {
	cli := client.GetClient()
	name := c.Query("name")
	namespace := c.DefaultQuery("namespace", "default")

	pod, err := cli.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, pod.Status.Phase)
}

func (p *Pod) int64ptr(i int64) *int64 {
	return &i
}
func (p *Pod) PodLog(c *gin.Context) {
	cli := client.GetClient()
	name := c.Query("name")
	namespace := c.DefaultQuery("namespace", "default")

	podlog := cli.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{})
	stream, err := podlog.Stream(context.TODO())
	if err != nil {
		tools.Failure(c, err)
		return
	}
	defer stream.Close()
	var b = new(bytes.Buffer)
	_, err = io.Copy(b, stream)
	if err != nil {
		tools.Failure(c, err)
		return
	}
	l := strings.Split(b.String(), "\n")
	tools.Success(c, l[:len(l)-1])
}

func (p *Pod) DeletePod(c *gin.Context) {
	cli := client.GetClient()
	name := c.Query("name")
	namespace := c.DefaultQuery("namespace", "default")

	err := cli.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "成功")
}

func (p *Pod) GetPod(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	name := c.Query("name")
	cli := client.GetClient()
	var label string
	deploy := c.Query("deployment")
	if len(deploy) > 0 {
		label = "app=" + deploy
	}
	if len(name) == 0 {
		list, err := cli.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: label})
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

	pod, err := cli.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, pod)
}

func (p *Pod) watch(name, namespace string) {
	cli := client.GetClient()

	watch, err := cli.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err)
		return
	}

	defer watch.Stop()

	for {
		select {
		case w := <-watch.ResultChan():
			log.Println(name + "---" + string(w.Type))
		}
	}
}
