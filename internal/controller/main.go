package controller

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

type Worker struct {
	kubeClient      kubernetes.Interface
	queue           workqueue.RateLimitingInterface
	informer        cache.SharedIndexInformer
	serverStartTime time.Time
	imageChan       chan Images
}

type Event struct {
	Key    string
	Action string
}
type Images struct {
	ID          string
	Images      []string
	PullSecrets []corev1.LocalObjectReference
}

func Start(clientset kubernetes.Interface, imageChan chan Images) {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	coreInformers := coreinformers.NewSharedInformerFactory(clientset, 0)

	deploymentInformer := coreInformers.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			queue.Add(Event{
				Key:    key,
				Action: "created",
			})
		},
		UpdateFunc: func(old, new interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(new)
			queue.Add(Event{
				Key:    key,
				Action: "updated",
			})
		},
		DeleteFunc: func(obj interface{}) {
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			queue.Add(Event{
				Key:    key,
				Action: "deleted",
			})
		},
	})

	w := &Worker{
		kubeClient: clientset,
		informer:   deploymentInformer,
		queue:      queue,
		imageChan:  imageChan,
	}

	stopCh := make(chan struct{})
	go w.Run(stopCh)
}

func (w *Worker) Run(stopCh <-chan struct{}) {
	defer w.queue.ShutDown()

	w.serverStartTime = time.Now().Local()

	go w.informer.Run(stopCh)

	wait.Until(w.runWorker, time.Second, stopCh)
}

func (w *Worker) runWorker() {
	for w.processNextItem() {
		// continue looping
	}
}

func (w *Worker) processNextItem() bool {
	newEvent, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(newEvent)

	err := w.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		w.queue.Forget(newEvent)
	} else if w.queue.NumRequeues(newEvent) < maxRetries {
		w.queue.AddRateLimited(newEvent)
	} else {
		w.queue.Forget(newEvent)
	}

	return true
}

func (w *Worker) processItem(newEvent Event) error {
	obj, _, err := w.informer.GetIndexer().GetByKey(newEvent.Key)
	if err != nil {
		return err
	}

	switch newEvent.Action {
	case "created":
		fallthrough
	case "updated":
		dp := obj.(*appsv1.Deployment)
		w.imageChan <- Images{
			ID:          newEvent.Key,
			Images:      getImages(*dp),
			PullSecrets: dp.Spec.Template.Spec.ImagePullSecrets,
		}
	case "deleted":
		w.imageChan <- Images{
			ID:          newEvent.Key,
			Images:      nil,
			PullSecrets: nil,
		}
	}
	return nil
}

func getImages(dp appsv1.Deployment) []string {
	images := []string{}

	for _, container := range append(dp.Spec.Template.Spec.InitContainers, dp.Spec.Template.Spec.Containers...) {
		images = append(images, container.Image)
	}

	return images
}
