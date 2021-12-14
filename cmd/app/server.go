package app

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/sample-controller/pkg/signals"

	"github.com/kube-queue/api/pkg/apis/scheduling/v1alpha1"
	queueversioned "github.com/kube-queue/api/pkg/client/clientset/versioned"
	queueinformers "github.com/kube-queue/api/pkg/client/informers/externalversions"
	"github.com/kube-queue/et-operator-extension/cmd/app/options"
	"github.com/kube-queue/et-operator-extension/pkg/controller"
	etjobv1alpha1 "github.com/kube-queue/et-operator-extension/pkg/et-operator/apis/et/v1alpha1"
	etjobversioned "github.com/kube-queue/et-operator-extension/pkg/et-operator/client/clientset/versioned"
	etjobinformers "github.com/kube-queue/et-operator-extension/pkg/et-operator/client/informers/externalversions"
	"k8s.io/klog/v2"
)

// Run runs the server.
func Run(opt *options.ServerOption) error {
	var restConfig *rest.Config
	var err error
	// Set up signals so we handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	if restConfig, err = rest.InClusterConfig(); err != nil {
		if restConfig, err = clientcmd.BuildConfigFromFlags("", opt.KubeConfig); err != nil {
			return err
		}
	}

	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	queueClient, err := queueversioned.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	etJobClient, err := etjobversioned.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	queueInformerFactory := queueinformers.NewSharedInformerFactory(queueClient, 0)
	queueInformer := queueInformerFactory.Scheduling().V1alpha1().QueueUnits().Informer()
	etJobInformerFactory := etjobinformers.NewSharedInformerFactory(etJobClient, 0)
	etJobInformer := etJobInformerFactory.Et().V1alpha1().TrainingJobs().Informer()

	etExtensionController := contorller.NewETExtensionController(
		k8sClientSet,
		queueInformerFactory.Scheduling().V1alpha1().QueueUnits(),
		queueClient,
		etJobInformerFactory.Et().V1alpha1().TrainingJobs(),
		etJobClient)

	queueInformer.AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch qu := obj.(type) {
				case *v1alpha1.QueueUnit:
					if qu.Spec.ConsumerRef != nil &&
						qu.Spec.ConsumerRef.Kind == contorller.ConsumerRefKind &&
						qu.Spec.ConsumerRef.APIVersion == contorller.ConsumerRefAPIVersion {
						return true
					}
					return false
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    etExtensionController.AddQueueUnit,
				UpdateFunc: etExtensionController.UpdateQueueUnit,
				DeleteFunc: etExtensionController.DeleteQueueUnit,
			},
		},
	)

	etJobInformer.AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch obj.(type) {
				case *etjobv1alpha1.TrainingJob:
					return true
				default:
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    etExtensionController.AddETJob,
				UpdateFunc: etExtensionController.UpdateETJob,
				DeleteFunc: etExtensionController.DeleteETJob,
			},
		},
	)

	// start queueunit informer
	go queueInformerFactory.Start(stopCh)
	// start tfjob informer
	go etJobInformerFactory.Start(stopCh)

	err = etExtensionController.Run(2, stopCh)
	if err != nil {
		klog.Fatalf("Error running tfExtensionController", err.Error())
		return err
	}

	return nil
}
