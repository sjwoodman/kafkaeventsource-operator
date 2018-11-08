package kafkaeventsource

import (
	"context"
	"github.com/knative/eventing-sources/pkg/controller/sinks"
	"k8s.io/client-go/rest"
	"log"

	sourcesv1alpha1 "github.com/rh-event-flow-incubator/kafkaeventsource-operator/pkg/apis/sources/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KafkaEventSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKafkaEventSource{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kafkaeventsource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KafkaEventSource
	err = c.Watch(&source.Kind{Type: &sourcesv1alpha1.KafkaEventSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KafkaEventSource
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sourcesv1alpha1.KafkaEventSource{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKafkaEventSource{}

// ReconcileKafkaEventSource reconciles a KafkaEventSource object
type ReconcileKafkaEventSource struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	dynamicClient dynamic.Interface
	scheme        *runtime.Scheme
}

func (r *ReconcileKafkaEventSource) InjectConfig(c *rest.Config) error {
	var err error
	r.dynamicClient, err = dynamic.NewForConfig(c)
	return err
}

// Reconcile reads that state of the cluster for a KafkaEventSource object and makes changes based on the state read
// and what is in the KafkaEventSource.Spec
func (r *ReconcileKafkaEventSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling KafkaEventSource %s/%s\n", request.Namespace, request.Name)

	// Fetch the KafkaEventSource
	kafkaEventSource := &sourcesv1alpha1.KafkaEventSource{}

	err := r.client.Get(context.TODO(), request.NamespacedName, kafkaEventSource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Resolve the SinkURI
	//todo: how to update status - r.client.Update fails because TransitionTime not set
	kafkaEventSource.Status.InitializeConditions()
	sinkURI, err := sinks.GetSinkURI(r.dynamicClient, kafkaEventSource.Spec.Sink, kafkaEventSource.Namespace)
	if err != nil {
		kafkaEventSource.Status.MarkNoSink("NotFound", "")
	}
	kafkaEventSource.Status.MarkSink(sinkURI)

	// Create a new deployment for this EventSource
	dep := deploymentForKafka(kafkaEventSource)

	// Set KafkaEventSource kafkaEventSource as the owner and controller
	if err := controllerutil.SetControllerReference(kafkaEventSource, dep, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating a new Pod %s/%s\n", dep.Namespace, dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	//todo: Check to see if it needs updating

	// Deployment already exists and de- don't requeue
	log.Printf("Skip reconcile: Deployment %s/%s already exists", found.Namespace, found.Name)
	return reconcile.Result{}, nil
}

func deploymentForKafka(kes *sourcesv1alpha1.KafkaEventSource) *appsv1.Deployment {
	labels := map[string]string{
		"app": kes.Name,
	}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kes.Name,
			Namespace: kes.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "sjwoodman/kafkaeventsource:latest",
						Name:  "kafkaeventsource",
						Env: []corev1.EnvVar{
							{
								Name:  "KAFKA_BOOTSTRAP_SERVERS",
								Value: kes.Spec.Bootstrap,
							},
							{
								Name:  "KAFKA_TOPIC",
								Value: kes.Spec.Topic,
							},
							{
								Name:  "TARGET",
								Value: kes.Status.SinkURI,
							},
						},
					}},
				},
			},
		},
	}
	return dep
}
