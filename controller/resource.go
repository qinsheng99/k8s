package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"k8s-demo/client"
	"k8s-demo/tools"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
)

type Resource struct{}
type ResListStatus struct {
	ServerCreatedFlag  bool
	ServerReadyFlag    bool
	ServerInactiveFlag bool
	ServerRecycledFlag bool
	ServerErroredFlag  bool
	ServerBoundFlag    bool
	ServerCreatedTime  string
	ServerReadyTime    string
	ServerBoundTime    string
	ServerInactiveTime string
	ServerRecycledTime string
	InstanceEndpoint   string
	ErrorInfo          string
}

func (r *Resource) resource() (kind *schema.GroupVersionKind, err error, _ *unstructured.Unstructured) {
	var yamldata []byte
	yamldata, err = ioutil.ReadFile("crd-resource.yaml")
	if err != nil {
		return nil, err, nil
	}
	obj := &unstructured.Unstructured{}
	_, kind, err = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(yamldata, nil, obj)
	if err != nil {
		return nil, err, nil
	}
	return kind, nil, obj
}

func (r *Resource) getResource() (schema.GroupVersionResource, error, *unstructured.Unstructured) {
	k, err, res := r.resource()
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}

	mapping, err := client.GetrestMapper().RESTMapping(k.GroupKind(), k.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err, nil
	}
	return mapping.Resource, nil, res
}

func (r *Resource) CreateResourceCrd(c *gin.Context) {
	cli := client.GetDyna()
	resource, err, res := r.getResource()
	if err != nil {
		tools.Failure(c, err)
		return
	}
	dr := cli.Resource(resource).Namespace("default")
	create, err := dr.Create(context.TODO(), res, metav1.CreateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}
	rls := r.validation(create, dr, res)
	if rls.ServerReadyFlag && !rls.ServerRecycledFlag && rls.ServerBoundFlag && !rls.ServerInactiveFlag {
		tools.Success(c, rls.InstanceEndpoint)
		return
	}

	if rls.ServerRecycledFlag {
		tools.Failure(c, fmt.Errorf("overdue"))
		return
	}

	if rls.ServerInactiveFlag || rls.ServerErroredFlag {
		//_ = dr.Delete(context.TODO(), create.GetName(), metav1.DeleteOptions{})
		label := "app=" + create.GetName()
		podList, err := client.GetClient().CoreV1().Pods(create.GetNamespace()).List(context.TODO(), metav1.ListOptions{LabelSelector: label})
		if err != nil {
			tools.Failure(c, err)
			return
		}
		var a = new(int64)
		*a = 3
		pod := podList.Items[0]
		logs := client.GetClient().CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetName(), &corev1.PodLogOptions{TailLines: a})
		stream, err := logs.Stream(context.TODO())
		if err != nil {
			tools.Failure(c, err)
			return
		}

		var buf = new(bytes.Buffer)
		_, err = io.Copy(buf, stream)
		if err != nil {
			tools.Failure(c, err)
			return
		}

		tools.Success(c, buf.String())
		return
	}
	tools.Success(c, create)
}

func (r *Resource) validation(code *unstructured.Unstructured, dr dynamic.ResourceInterface, object *unstructured.Unstructured) ResListStatus {
	var err error
	var num int
try:
	rls := ResListStatus{}
	code, err = dr.Get(context.TODO(), code.GetName(), metav1.GetOptions{})
	if err != nil {
		num++
		if num >= 10 {
			err = dr.Delete(context.TODO(), code.GetName(), metav1.DeleteOptions{})
			rls.ServerErroredFlag = true
		}
		goto try
	} else {
		if object.GetAPIVersion() == code.GetAPIVersion() {
			metadata, ok := tools.ParsingMap(code.Object, "metadata")
			if !ok {
				rls.ServerErroredFlag = true
				return rls
			}

			name, nameok := tools.ParsingMapStr(metadata, "name")
			if !nameok {
				rls.ServerErroredFlag = true
				return rls
			}

			if name != object.GetName() {
				rls.ServerErroredFlag = true
				return rls
			}

			status, statusok := tools.ParsingMap(code.Object, "statsus")
			if !statusok {
				return rls
			}

			conditions, conditionok := tools.ParsingMapSlice(status, "condition")
			if !conditionok {
				return rls
			}

			for _, condition := range conditions {
				cond := condition.(map[string]interface{})
				typ, typeok := tools.ParsingMapStr(cond, "type")
				if !typeok {
					continue
				}

				switch typ {
				case "ServerCreated": //means the code server has been accepted by the system.
					create, createok := tools.ParsingMapStr(cond, "status")
					if createok && create == "True" {
						rls.ServerCreatedFlag = true
					}
					lastTransitionTime, createtimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if createtimeok {
						rls.ServerCreatedTime = lastTransitionTime
					}
				case "ServerReady": //means the code server has been ready for usage.
					ready, readyok := tools.ParsingMapStr(cond, "status")
					if readyok && ready == "True" {
						rls.ServerReadyFlag = true
					}
					lastTransitionTime, timeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if timeok {
						rls.ServerReadyTime = lastTransitionTime
					}
					message, messageok := tools.ParsingMap(cond, "message")
					if messageok {
						instanceEndpoint, pointok := tools.ParsingMapStr(message, "instanceEndpoint")
						if pointok {
							rls.InstanceEndpoint = instanceEndpoint
						}
					}
				case "ServerBound": //means the code server has been bound to user.
					bound, boundok := tools.ParsingMapStr(cond, "status")
					if boundok && bound == "True" {
						rls.ServerBoundFlag = true
					}
					lastTransitionTime, boundtimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if boundtimeok {
						rls.ServerBoundTime = lastTransitionTime
					}
				case "ServerRecycled": //means the code server has been recycled totally.
					recycled, recycleok := tools.ParsingMapStr(cond, "status")
					if recycleok && recycled == "True" {
						rls.ServerRecycledFlag = true
					}
					lastTransitionTime, recycletimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if recycletimeok {
						rls.ServerRecycledTime = lastTransitionTime
					}
				case "ServerInactive": //means the code server will be marked inactive if `InactiveAfterSeconds` elapsed
					inactive, inactiveok := tools.ParsingMapStr(cond, "status")
					if inactiveok && inactive == "True" {
						rls.ServerInactiveFlag = true
					}
					lastTransitionTime, inactivetimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if inactivetimeok {
						rls.ServerInactiveTime = lastTransitionTime
					}
				case "ServerErrored": //means failed to reconcile code server.
					bound, errorok := tools.ParsingMapStr(cond, "status")
					if errorok && bound == "True" {
						rls.ServerBoundFlag = true
					}
					lastTransitionTime, errortimeok := tools.ParsingMapStr(cond, "lastTransitionTime")
					if errortimeok {
						rls.ServerBoundTime = lastTransitionTime
					}
				}
			}
		}
	}
	return rls
}

func (r *Resource) Get(c *gin.Context) {
	cli := client.GetDyna()
	resource, err, _ := r.getResource()
	if err != nil {
		tools.Failure(c, err)
		return
	}
	name := c.Query("name")
	if len(name) == 0 {
		unstructuredList, err := cli.Resource(resource).
			Namespace("default").
			List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			tools.Failure(c, err)
			return
		}
		if unstructuredList == nil {
			tools.Failure(c, errors.New("data is nil"))
			return
		}
		l := *unstructuredList
		var res []interface{}
		for _, v := range l.Items {
			res = append(res, map[string]string{
				"name":      v.GetName(),
				"version":   v.GetAPIVersion(),
				"kind":      v.GetKind(),
				"namespace": v.GetNamespace(),
			})
		}

		tools.Success(c, res)
		return
	}
	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	delete(get.Object, "metadata")

	tools.Success(c, get.Object)
}

func (r *Resource) GetCrd() (schema.GroupVersionResource, error, *unstructured.Unstructured) {
	return r.getResource()
}

func (r *Resource) UpdateResourceCrd(c *gin.Context) {
	cli := client.GetDyna()
	resource, err, _ := r.getResource()
	if err != nil {
		tools.Failure(c, err)
		return
	}
	get, err := cli.Resource(resource).Namespace("default").Get(context.TODO(), c.Query("name"), metav1.GetOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	if sp, ok := get.Object["spec"]; ok {
		if spc, ok := sp.(map[string]interface{}); ok {
			spc["add"] = true
			spc["recycleAfterSeconds"] = 100
		}
	}

	_, err = cli.Resource(resource).Namespace("default").Update(context.TODO(), get, metav1.UpdateOptions{})
	if err != nil {
		tools.Failure(c, err)
		return
	}

	tools.Success(c, "success")

}
