package application

import (
	"context"
	"strings"

	applicationv1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_application")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Application Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileApplication{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("application-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Application
	err = c.Watch(&source.Kind{Type: &applicationv1alpha1.Application{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileApplication implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileApplication{}

// ReconcileApplication reconciles a Application object
type ReconcileApplication struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Application object and makes changes based on the state read
// and what is in the Application.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileApplication) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Application")

	// Fetch the Application instance
	instance := &applicationv1alpha1.Application{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	projectID := instance.Annotations["cattle.topicus.nl/projectId"]

	if projectID != "" {
		log.Info("Found ArgoCD Application", "name", instance.Name, "cattle project", projectID)
	} else {
		log.Info("Skipping ArgoCD Application", "name", instance.Name, "reason", "No 'cattle.topicus.nl/projectId' annotation")
		return reconcile.Result{}, nil
	}

	// Define a new Namespace object
	ns := newNamespaceForCR(instance)

	// Set Application instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ns, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Namespace already exists
	found := &corev1.Namespace{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ns.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", "Name", ns.Name)
		err = r.client.Create(context.TODO(), ns)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Namespace created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// todo check if uodate is nodig

	// todo update namespace
	reqLogger.Info("Updating Namespace", "Name", ns.Name)
	err = r.client.Update(context.TODO(), ns)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Namespace updated successfully - don't requeue
	return reconcile.Result{}, nil

	//reqLogger.Info("Skip reconcile: Namespace already exists", "Name", found.Name)
	// Pod already exists - don't requeue
	//return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newNamespaceForCR(cr *applicationv1alpha1.Application) *corev1.Namespace {
	projectID := cr.Annotations["cattle.topicus.nl/projectId"]
	shortProjectID := strings.Split(projectID, ":")[0]

	// FIXME alles optioneel maken, dan kan het ook zonder rancher/cattle gebruikt worden
	labels := map[string]string{
		"field.cattle.io/projectId": shortProjectID,
	}
	annotations := map[string]string{
		"field.cattle.io/projectId": projectID,
	}
	// FIXME: lijst, zodat je meerdere lables kunt opgeven
	extraLabel := cr.Annotations["argocd-namespace.topicus.nl/label"]
	if extraLabel != "" {
		extraLabelEntry := strings.Split(extraLabel, ":")
		labels[strings.TrimSpace(extraLabelEntry[0])] = strings.TrimSpace(extraLabelEntry[1])
	}

	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Spec.Destination.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.NamespaceSpec{},
	}
}
