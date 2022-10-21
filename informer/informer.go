package informer

import (
	"k8s.io/client-go/kubernetes"
	"sync"
	"time"
)

func NewListen(res *kubernetes.Clientset, namespace string) ListenInter {
	return &Listen{res: res, namespace: namespace, wg: &sync.WaitGroup{}}
}

type ListenInter interface {
	ListenResource()
}

type Listen struct {
	wg        *sync.WaitGroup
	res       *kubernetes.Clientset
	namespace string
	resync    time.Duration
}

func (l *Listen) ListenResource() {
	//ifac := informers.NewSharedInformerFactory(l.res, l.resync)
	//ifac
}
