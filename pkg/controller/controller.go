package contorller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	etJobv1alpha1 "github.com/kube-queue/et-operator-extension/pkg/et-operator/apis/et/v1alpha1"
	etjobversioned "github.com/kube-queue/et-operator-extension/pkg/et-operator/client/clientset/versioned"
	etjobinformers "github.com/kube-queue/et-operator-extension/pkg/et-operator/client/informers/externalversions/et/v1alpha1"
	etJobOfficalv1 "github.com/AliyunContainerService/et-operator/pkg/controllers/api/v1"
	"github.com/kube-queue/api/pkg/apis/scheduling/v1alpha1"
	queueversioned "github.com/kube-queue/api/pkg/client/clientset/versioned"
	queueInformers "github.com/kube-queue/api/pkg/client/informers/externalversions/scheduling/v1alpha1"
)

const (
	// MaxRetries is the number of times a queue item will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a queue item is going to be requeued:
	//
	// 1-10 retry times: 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s,
	// 11-20 retry times: 5.1s, 10.2s, 20.4s, 41s, 82s, 164s, 328s, 656s(11min), 1312s(21min), 2624s(43min)
	MaxRetries = 15
	// Suspend is a flag annotation for etJob to use the queueunit crd
	Suspend = "scheduling.x-k8s.io/suspend"
)

const (
	ConsumerRefKind       = "TrainingJob"
	ConsumerRefAPIVersion = "kai.alibabacloud.com/v1alpha1"
	// QuNameSuffix is the suffix of the queue unit name when create a new one.
	// In this way, different types of jobs with the same name will create different queue unit name.
	QuNameSuffix = "-pytorch-qu"
	Queuing      = "Queuing"
)

type ETExtensionController struct {
	k8sClient          *kubernetes.Clientset
	queueInformer      queueInformers.QueueUnitInformer
	queueClient        *queueversioned.Clientset
	etJobInformer etjobinformers.TrainingJobInformer
	etJobClient   *etjobversioned.Clientset
	workqueue          workqueue.RateLimitingInterface
}

func NewETExtensionController(
	k8sClient *kubernetes.Clientset,
	queueInformer queueInformers.QueueUnitInformer,
	queueClient *queueversioned.Clientset,
	etJobInformer etjobinformers.TrainingJobInformer,
	etJobClient *etjobversioned.Clientset) *ETExtensionController {
	return &ETExtensionController{
		k8sClient:          k8sClient,
		queueInformer:      queueInformer,
		queueClient:        queueClient,
		etJobInformer: etJobInformer,
		etJobClient:   etJobClient,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "QueueUnit"),
	}
}

func (pc *ETExtensionController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer pc.workqueue.ShutDown()

	klog.Info("Start ETExtensionController Run function")
	if !cache.WaitForCacheSync(stopCh, pc.queueInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(pc.runWorker, time.Second, stopCh)
	}
	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

func (pc *ETExtensionController) runWorker() {
	for pc.processNextWorkItem() {
	}
}

func (pc *ETExtensionController) processNextWorkItem() bool {
	obj, shutdown := pc.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer pc.workqueue.Done.
	err := func(obj interface{}) error {
		defer pc.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			pc.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		err := pc.syncHandler(key)
		pc.handleErr(err, key)

		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}

	return true
}

func (pc *ETExtensionController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return err
	}
	// Get queueunit from cache
	queueUnit, err := pc.queueInformer.Lister().QueueUnits(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Errorf("failed to find qu by:%v/%v, maybe qu has been deleted", namespace, name)
			// If can't get queueunit, return nil, handleErr function will forget key from workqueue
			return nil
		}
		runtime.HandleError(fmt.Errorf("failed to get queueunit by: %s/%s", namespace, name))

		return err
	}
	klog.Infof("Get informer from add/update event,queueUnit:%v/%v", queueUnit.Namespace, queueUnit.Name)

	if queueUnit.Status.Phase == v1alpha1.Dequeued {
		klog.Infof("QueueUnit %v/%v has dequeued", queueUnit.Namespace, queueUnit.Name)
		err = pc.DeleteQueueAnnotationInetJob(queueUnit)
		if errors.IsNotFound(err) {
			// If can't find etJob for queueunit, return err, handleErr function will requeue key MaxRetries times
			return err
		}
	}

	return nil
}

func (pc *ETExtensionController) deleteQueueUnitWhenJobNotFound(key string) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return
	}

	err = pc.deleteQueueUnitInstance(namespace, name)
	if err != nil {
		// TODO: need to add retry policy for case that delete queueunit failed
		klog.Errorf("Delete queueunit error: %v/%v %v", namespace, name, err.Error())
		return
	}
	klog.Warningf("Delete queueunit %v/%v because can't find related etJob ", namespace, name)
}

func (pc *ETExtensionController) handleErr(err error, key string) {
	if err == nil {
		pc.workqueue.Forget(key)
		return
	}

	if pc.workqueue.NumRequeues(key) < MaxRetries {
		pc.workqueue.AddRateLimited(key)
		klog.Infof("We will requeue %v %d times,because:%v, has retried %d times", key, MaxRetries, err, pc.workqueue.NumRequeues(key)+1)
		return
	}

	runtime.HandleError(err)
	klog.Infof("Dropping queueunit %q out of the workqueue: %v", key, err)
	pc.workqueue.Forget(key)
	// If still can't find job after retry, delete queueunit
	pc.deleteQueueUnitWhenJobNotFound(key)
}

func (pc *ETExtensionController) enqueueQueueUnit(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	pc.workqueue.AddRateLimited(key)
}

func (pc *ETExtensionController) AddQueueUnit(obj interface{}) {
	qu := obj.(*v1alpha1.QueueUnit)
	klog.Infof("Add queueunit:%v/%v", qu.Namespace, qu.Name)
	pc.enqueueQueueUnit(qu)
}

func (pc *ETExtensionController) UpdateQueueUnit(oldObj, newObj interface{}) {
	oldQu := oldObj.(*v1alpha1.QueueUnit)
	newQu := newObj.(*v1alpha1.QueueUnit)
	if oldQu.ResourceVersion == newQu.ResourceVersion {
		return
	}
	pc.enqueueQueueUnit(newQu)
}

func (pc *ETExtensionController) DeleteQueueUnit(obj interface{}) {
	qu := obj.(*v1alpha1.QueueUnit)
	klog.Infof("QueueUnit deleted:%v/%v", qu.Namespace, qu.Name)
}

func (pc *ETExtensionController) createQueueUnitInstance(etJob *etJobv1alpha1.TrainingJob) error {
	// 1. try to get annotation scheduling.x-k8s.io/suspend
	_, ok := etJob.Annotations[Suspend]
	if !ok {
		klog.Infof("etJob %v/%v is not scheduled by kube-queue", etJob.Namespace, etJob.Name)
		return nil
	}

	// 2. annotation has been found and try to get queueunit from cache
	qu, err := pc.queueInformer.Lister().QueueUnits(etJob.Namespace).Get(etJob.Name + QuNameSuffix)
	if err != nil {
		if errors.IsNotFound(err) {
			// 2.1 there is no specified queueunit in k8s
			klog.Infof("Creating queueunit for etJob %v/%v", etJob.Namespace, etJob.Name)
			// 2.2 generate a new queueunit
			quMeta := pc.generateQueueUnitInstance(etJob)
			// 2.3 create the queueunit
			qu, err = pc.queueClient.SchedulingV1alpha1().QueueUnits(quMeta.Namespace).Create(context.TODO(), quMeta, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			klog.Infof("Created queueunit %v/%v successfully", qu.Namespace, qu.Name)
			return nil
		} else {
			return err
		}
	}

	if qu.Spec.ConsumerRef.Kind == ConsumerRefKind {
		klog.Infof("It already has a queueunit %v/%v for etJob %v/%v",
			qu.Namespace, qu.Name, etJob.Namespace, etJob.Name)
	} else {
		klog.Warningf("There is an exception queueunit:%v/%v for etJob in k8s, please check it", qu.Namespace, qu.Name)
	}

	return nil
}

func (pc *ETExtensionController) generateQueueUnitInstance(etJob *etJobv1alpha1.TrainingJob) *v1alpha1.QueueUnit {
	// 1. build ObjectReference from corresponding Job CR
	objectReference := pc.generateObjectReference(etJob)
	// 2. get priorityClassName and priority from one of etJob roles
	var priorityClassName string
	var priority *int32

	priorityClassName = etJob.Spec.ETReplicaSpecs.Worker.Template.Spec.PriorityClassName
	priority = etJob.Spec.ETReplicaSpecs.Worker.Template.Spec.Priority

	// If there is a related priorityClassInstance in K8s, we use priorityClass's value instead of etJob.Spec.PyTorchReplicaSpecs[role].Template.Spec.Priority
	if priorityClassName != "" {
		priorityClassInstance, err := pc.k8sClient.SchedulingV1().PriorityClasses().Get(priorityClassName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("Can not find priority class name %v in k8s, we will ignore it", priorityClassName)
			} else {
				klog.Errorf("Can not get PriorityClass %v from k8s for etJob:%v/%v, err:%v", priorityClassName, etJob.Namespace, etJob.Name, err)
			}
		} else {
			priority = &priorityClassInstance.Value
		}
	}

	// 3. calculate the total resources of this pytorch instance
	resources := pc.calculateTotalResources(etJob)
	// 4. build QueueUnit
	return &v1alpha1.QueueUnit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      etJob.Name + QuNameSuffix,
			Namespace: etJob.Namespace,
		},
		Spec: v1alpha1.QueueUnitSpec{
			ConsumerRef:       objectReference,
			Priority:          priority,
			PriorityClassName: priorityClassName,
			Resource:          resources,
		},
		Status: v1alpha1.QueueUnitStatus{
			Phase:   v1alpha1.Enqueued,
			Message: "the queueunit is enqueued after created",
		},
	}
}

func (pc *ETExtensionController) generateObjectReference(etJob *etJobv1alpha1.TrainingJob) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: ConsumerRefAPIVersion,
		Kind:       ConsumerRefKind,
		Namespace:  etJob.Namespace,
		Name:       etJob.Name,
	}
}

func (pc *ETExtensionController) calculateTotalResources(etJob *etJobv1alpha1.TrainingJob) corev1.ResourceList {
	totalResources := corev1.ResourceList{}
	numbers := len(etJob.Status.CurrentWorkers)
	// calculate the total resource request
	count := int(*etJob.Spec.ETReplicaSpecs.Worker.Replicas)
	containers := etJob.Spec.ETReplicaSpecs.Worker.Template.Spec.Containers
	for i := 0;i < numbers;i++ {
		for _, container := range containers {
			// calculate the resource request of pods first (the pod count is decided by replicas's number)
			resources := container.Resources.Requests
			for resourceType := range resources {
				quantity := resources[resourceType]
				// scale the quantity by count
				replicaQuantity := resource.Quantity{}
				for i := 1; i <= count; i++ {
					replicaQuantity.Add(quantity)
				}
				// check if the resourceType is in totalResources
				if totalQuantity, ok := totalResources[resourceType]; !ok {
					// not in: set this replicaQuantity
					totalResources[resourceType] = replicaQuantity
				} else {
					// in: append this replicaQuantity and update
					totalQuantity.Add(replicaQuantity)
					totalResources[resourceType] = totalQuantity
				}
			}
		}
	}
	return totalResources
}

func (pc *ETExtensionController) AddETJob(obj interface{}) {
	etJob := obj.(*etJobv1alpha1.TrainingJob)
	klog.Infof("Get add etJob %v/%v event", etJob.Namespace, etJob.Name)
	err := pc.createQueueUnitInstance(etJob)
	if err != nil {
		klog.Errorf("Can't create queueunit for etJob %v/%v,err is:%v", etJob.Namespace, etJob.Name, err)
	}

	if etJob.Status.Conditions == nil {
		etJob.Status.Conditions = make([]etJobOfficalv1.JobCondition, 0)
		etJob.Status.Conditions = append(etJob.Status.Conditions, etJobOfficalv1.JobCondition{
			Type:           Queuing,
			LastUpdateTime: metav1.Now(),
		})
		_, err := pc.etJobClient.EtV1alpha1().TrainingJobs(etJob.Namespace).UpdateStatus(context.TODO(), etJob, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("update etJob failed Queuing %v/%v %v", etJob.Namespace, etJob.Name, err.Error())
		}
		klog.Infof("update etJob %v/%v status Queuing successfully", etJob.Namespace, etJob.Name)
	}
}

func (pc *ETExtensionController) UpdateETJob(_, newObj interface{}) {
	newetJob := newObj.(*etJobv1alpha1.TrainingJob)
	conditionsLen := len(newetJob.Status.Conditions)
	if conditionsLen > 0 {
		lastCondition := newetJob.Status.Conditions[conditionsLen-1]
		if lastCondition.Type == etJobOfficalv1.JobFailed || lastCondition.Type == etJobOfficalv1.JobSucceeded {
			klog.Infof("job %v/%v finished, current lastCondition.Type: [%v]", newetJob.Namespace, newetJob.Name, lastCondition.Type)
			pc.deleteQueueUnitAfterJobTerminated(newetJob)
		}
	}
}

func (pc *ETExtensionController) DeleteETJob(obj interface{}) {
	job := obj.(*etJobv1alpha1.TrainingJob)
	pc.deleteQueueUnitAfterJobTerminated(job)
}

func (pc *ETExtensionController) deleteQueueUnitAfterJobTerminated(etJob *etJobv1alpha1.TrainingJob) {
	// Get queueunit from cache
	qu, err := pc.queueInformer.Lister().QueueUnits(etJob.Namespace).Get(etJob.Name + QuNameSuffix)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Errorf("failed to get related queueunit by etJob:%v/%v when delete queueunit, "+
				"maybe qu has been deleted", etJob.Namespace, etJob.Name)
			return
		}
	}
	if qu.Spec.ConsumerRef.Name == etJob.Name && qu.Spec.ConsumerRef.Kind == ConsumerRefKind {
		err = pc.deleteQueueUnitInstance(qu.Namespace, qu.Name)
		if err != nil {
			klog.Errorf("Delete queueunit error: delete qu failed %v/%v %v", qu.Namespace, qu.Name, err)
		}
		klog.Infof("Delete queueunit %s because related etJob %v/%v terminated", qu.Name, etJob.Namespace, etJob.Name)
	}
}

func (pc *ETExtensionController) deleteQueueUnitInstance(namespace, name string) error {
	err := pc.queueClient.SchedulingV1alpha1().QueueUnits(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (pc *ETExtensionController) DeleteQueueAnnotationInetJob(qu *v1alpha1.QueueUnit) error {
	namespace := qu.Spec.ConsumerRef.Namespace
	etJobName := qu.Spec.ConsumerRef.Name
	etJob, err := pc.etJobClient.EtV1alpha1().TrainingJobs(qu.Spec.ConsumerRef.Namespace).Get(context.TODO(), etJobName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Warningf("Can not find related etJob:%v for queueunit:%v in namespace:%v", etJobName, qu.Name, namespace)
			return err
		}
		klog.Errorf("Get etJob failed %v/%v %v", namespace, etJobName, err.Error())
		return err
	}

	var annotation = map[string]string{}
	for k, v := range etJob.Annotations {
		if k != Suspend {
			annotation[k] = v
		}
	}
	etJob.SetAnnotations(annotation)

	// TODO change to patch
	_, err = pc.etJobClient.EtV1alpha1().TrainingJobs(namespace).Update(context.TODO(), etJob, metav1.UpdateOptions{})
	if err != nil {
		klog.Errorf("UpdateQueueUnit error: update etJob failed %v/%v %v", namespace, etJobName, err.Error())
		return err
	}
	klog.Infof("Update annotations for etJob %v/%v", etJob.Namespace, etJob.Name)

	return nil
}
