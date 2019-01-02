package main

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	appslisters "k8s.io/client-go/listers/apps/v1beta2"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	clientset "k8s.io/rules/pkg/client/clientset/versioned"
	samplescheme "k8s.io/rules/pkg/client/clientset/versioned/scheme"
	informers "k8s.io/rules/pkg/client/informers/externalversions"
	listers "k8s.io/rules/pkg/client/listers/crd/v1"
	ruleV1 "k8s.io/rules/pkg/apis/crd/v1"
)

const controllerAgentName = "sample-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	//ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	//MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Rule synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	replicasetsLister appslisters.ReplicaSetLister
	replicasetsSynced cache.InformerSynced
	statefulsetsLister appslisters.StatefulSetLister
	statefulsetsSynced cache.InformerSynced
	daemonsetsLister appslisters.DaemonSetLister
	daemonsetsSynced cache.InformerSynced
	rulesLister        listers.RuleLister
	rulesSynced        cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewController(kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sampleInformerFactory informers.SharedInformerFactory) *Controller {


	ruleInformer := sampleInformerFactory.Crd().V1().Rules()
	deploymentInformer := kubeInformerFactory.Apps().V1beta2().Deployments();
	replicasetInformer := kubeInformerFactory.Apps().V1beta2().ReplicaSets();
	statefulsetInformer := kubeInformerFactory.Apps().V1beta2().StatefulSets();
	daemonsetInformer := kubeInformerFactory.Apps().V1beta2().DaemonSets();

	samplescheme.AddToScheme(scheme.Scheme)
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		replicasetsLister: replicasetInformer.Lister(),
		replicasetsSynced: replicasetInformer.Informer().HasSynced,
		statefulsetsLister: statefulsetInformer.Lister(),
		statefulsetsSynced: statefulsetInformer.Informer().HasSynced,
		daemonsetsLister: daemonsetInformer.Lister(),
		daemonsetsSynced: daemonsetInformer.Informer().HasSynced,
		rulesLister:        ruleInformer.Lister(),
		rulesSynced:        ruleInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Foos"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	ruleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFoo(new)
		},
		DeleteFunc:controller.enqueueFoo,
	})

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error{
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Foo controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh,  c.rulesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	num:=0
	for c.processNextWorkItem() {
		glog.Info("WorkQueue Len2:",c.workqueue.Len())
		glog.Info("Num:",num)
		num++
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	glog.Info("Obj:",obj)
	//defer c.workqueue.Done(obj)
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			c.workqueue.Forget(obj)
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	//c.workqueue.Forget(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	rule, err := c.rulesLister.Rules(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	switch rule.Spec.OwnerType {
	case "deployment":
		deployment,err:=c.deploymentsLister.Deployments(rule.Spec.Namespace).Get(rule.Spec.OwnerName)
		if errors.IsNotFound(err) {
			//deployment, err = c.kubeclientset.AppsV1beta2().Deployments(rule.Namespace).Create(newDeployment(rule,deployment))
			glog.Info("Not Found1")
		}
		if err != nil {

			glog.Info("Err1:",err)
			return err
		}
		//ToDo
		if rule.Spec.Replicas != nil {
			_,err = c.kubeclientset.AppsV1beta2().Deployments(rule.Spec.Namespace).Update(newDeployment(rule,deployment))
		}
		if err != nil {

			glog.Info("Err1:",err)
			return err
		}

	case "replicaset":
		glog.Info("NewReplicas")
		replicaset,err := c.replicasetsLister.ReplicaSets(rule.Spec.Namespace).Get(rule.Spec.OwnerName)
		if errors.IsNotFound(err) {
			//ToDO
		}
		if err != nil {
			return err
		}
		//ToDo
		if rule.Spec.Replicas != nil {
			_,err = c.kubeclientset.AppsV1beta2().ReplicaSets(rule.Spec.Namespace).Update(newReplicaset(rule,replicaset))
		}
	case "statefulset":
		statefulset,err:=c.statefulsetsLister.StatefulSets(rule.Spec.Namespace).Get(rule.Spec.OwnerName);
		if errors.IsNotFound(err) {
			//ToDO
		}
		if err != nil {
			return err
		}
		//ToDo
		if rule.Spec.Replicas != nil {
			_,err = c.kubeclientset.AppsV1beta2().StatefulSets(rule.Spec.Namespace).Update(newStatefulset(rule,statefulset))
		}
	case "deamonset":
		daemonset,err := c.daemonsetsLister.DaemonSets(rule.Spec.Namespace).Get(rule.Spec.OwnerName);
		if errors.IsNotFound(err) {
			//ToDO
		}
		if err != nil {
			return err
		}
		//ToDo
		if rule.Spec.Replicas != nil {
			_,err = c.kubeclientset.AppsV1beta2().DaemonSets(rule.Spec.Namespace).Update(newDaemonset(rule,daemonset))
		}
	default:

	}

	c.recorder.Event(rule, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
	glog.Info("WorkQueue:",c.workqueue.Len())
}

func newDeployment(rule *ruleV1.Rule,deployment *appsv1beta2.Deployment)*appsv1beta2.Deployment{
	//replicas := *(rule.Spec.Replicas)+*(deployment.Spec.Replicas)
	//glog.Info("Replicas:",replicas)
	return &appsv1beta2.Deployment{
		ObjectMeta:metav1.ObjectMeta{
			Name:deployment.ObjectMeta.Name,
			Namespace:deployment.ObjectMeta.Namespace,
			OwnerReferences:[]metav1.OwnerReference{
				*metav1.NewControllerRef(rule,schema.GroupVersionKind{
					Group: ruleV1.SchemeGroupVersion.Group,
					Version:ruleV1.SchemeGroupVersion.Version,
					Kind: "Rule",
				}),
			},
		},
		Spec:appsv1beta2.DeploymentSpec{
			//Replicas:&replicas,
			Replicas:rule.Spec.Replicas,
			Selector:&metav1.LabelSelector{
				MatchLabels:deployment.Spec.Selector.MatchLabels,
			},
			Template:deployment.Spec.Template,
		},
	}
}

func newReplicaset(rule *ruleV1.Rule,replicaset *appsv1beta2.ReplicaSet)*appsv1beta2.ReplicaSet{
	replicas := *(rule.Spec.Replicas)+*(replicaset.Spec.Replicas)
	return &appsv1beta2.ReplicaSet{
		ObjectMeta:metav1.ObjectMeta{
			Name:replicaset.ObjectMeta.Name,
			Namespace:replicaset.ObjectMeta.Namespace,
			OwnerReferences:[]metav1.OwnerReference{
				*metav1.NewControllerRef(rule,schema.GroupVersionKind{
					Group: ruleV1.SchemeGroupVersion.Group,
					Version:ruleV1.SchemeGroupVersion.Version,
					Kind: "Rule",
				}),
			},
		},
		Spec:appsv1beta2.ReplicaSetSpec{
			Replicas:&replicas,
			Selector:&metav1.LabelSelector{
				MatchLabels:replicaset.Spec.Selector.MatchLabels,
			},
			Template:replicaset.Spec.Template,
		},
	}
}

func newStatefulset(rule *ruleV1.Rule, statefulset *appsv1beta2.StatefulSet)*appsv1beta2.StatefulSet{
	replicas := *(rule.Spec.Replicas)+*(statefulset.Spec.Replicas)
	return &appsv1beta2.StatefulSet{
		ObjectMeta:metav1.ObjectMeta{
			Name:statefulset.ObjectMeta.Name,
			Namespace:statefulset.ObjectMeta.Namespace,
			OwnerReferences:[]metav1.OwnerReference{
				*metav1.NewControllerRef(rule,schema.GroupVersionKind{
					Group: ruleV1.SchemeGroupVersion.Group,
					Version:ruleV1.SchemeGroupVersion.Version,
					Kind: "Rule",
				}),
			},
		},
		Spec:appsv1beta2.StatefulSetSpec{
			Replicas:&replicas,
			Selector:&metav1.LabelSelector{
				MatchLabels:statefulset.Spec.Selector.MatchLabels,
			},
			Template:statefulset.Spec.Template,
		},
	}
}

func newDaemonset(rule *ruleV1.Rule,daemonset *appsv1beta2.DaemonSet)*appsv1beta2.DaemonSet{
	//replicas := *(rule.Spec.Replicas)+*(daemonset.Spec.Replicas)
	return &appsv1beta2.DaemonSet{
		ObjectMeta:metav1.ObjectMeta{
			Name:daemonset.ObjectMeta.Name,
			Namespace:daemonset.ObjectMeta.Namespace,
			OwnerReferences:[]metav1.OwnerReference{
				*metav1.NewControllerRef(rule,schema.GroupVersionKind{
					Group: ruleV1.SchemeGroupVersion.Group,
					Version:ruleV1.SchemeGroupVersion.Version,
					Kind: "Rule",
				}),
			},
		},
		Spec:appsv1beta2.DaemonSetSpec{
			Selector:&metav1.LabelSelector{
				MatchLabels:daemonset.Spec.Selector.MatchLabels,
			},
			Template:daemonset.Spec.Template,
		},
	}
}