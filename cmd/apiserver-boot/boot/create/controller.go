package create

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/markbates/inflect"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/input"
	"sigs.k8s.io/kubebuilder/pkg/scaffold/resource"
)

// Controller scaffolds a Controller for a Resource
type Controller struct {
	input.Input

	// Resource is the Resource to make the Controller for
	Resource *resource.Resource

	// ResourcePackage is the package of the Resource
	ResourcePackage string

	// Plural is the plural lowercase of kind
	Plural string

	// Is the Group + "." + Domain for the Resource
	GroupDomain string
}

// GetInput implements input.File
func (a *Controller) GetInput() (input.Input, error) {
	// Use the k8s.io/api package for core resources
	coreGroups := map[string]string{
		"apps":                  "",
		"admissionregistration": "k8s.io",
		"apiextensions":         "k8s.io",
		"authentication":        "k8s.io",
		"autoscaling":           "",
		"batch":                 "",
		"certificates":          "k8s.io",
		"core":                  "",
		"extensions":            "",
		"metrics":               "k8s.io",
		"policy":                "",
		"rbac.authorization":    "k8s.io",
		"storage":               "k8s.io",
	}

	a.ResourcePackage, a.GroupDomain = getResourceInfo(coreGroups, a.Resource, a.Input)

	if a.Plural == "" {
		rs := inflect.NewDefaultRuleset()
		a.Plural = rs.Pluralize(strings.ToLower(a.Resource.Kind))
	}

	if a.Path == "" {
		a.Path = filepath.Join("pkg", "controller",
			strings.ToLower(a.Resource.Kind),
			strings.ToLower(a.Resource.Kind)+"_controller.go")
	}
	a.TemplateBody = controllerTemplate
	a.Input.IfExistsAction = input.Error
	return a.Input, nil
}

func getResourceInfo(coreGroups map[string]string, r *resource.Resource, in input.Input) (resourcePackage, groupDomain string) {
	resourcePath := filepath.Join("pkg", "apis", r.Group, r.Version,
		fmt.Sprintf("%s_types.go", strings.ToLower(r.Kind)))
	if _, err := os.Stat(resourcePath); os.IsNotExist(err) {
		if domain, found := coreGroups[r.Group]; found {
			resourcePackage := path.Join("k8s.io", "api")
			groupDomain = r.Group
			if domain != "" {
				groupDomain = r.Group + "." + domain
			}
			return resourcePackage, groupDomain
		}
		// TODO: need to support '--resource-pkg-path' flag for specifying resourcePath
	}
	return path.Join(in.Repo, "pkg", "apis"), r.Group + "." + in.Domain
}

var controllerTemplate = `{{ .Boilerplate }}

package {{ lower .Resource.Kind }}

import (
{{ if .Resource.CreateExampleReconcileBody }}	"context"
	"log"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	{{ .Resource.Group}}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Group}}/{{ .Resource.Version }}"
)

var log = logf.Log.WithName("{{ lower .Resource.Kind }}-controller")
{{ else }}	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	{{ .Resource.Group}}{{ .Resource.Version }} "{{ .ResourcePackage }}/{{ .Resource.Group}}/{{ .Resource.Version }}"
)
{{ end -}}

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new {{ .Resource.Kind }} Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &Reconcile{{ .Resource.Kind }}{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("{{ lower .Resource.Kind }}-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to {{ .Resource.Kind }}
	err = c.Watch(&source.Kind{Type: &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by {{ .Resource.Kind }} - change this for objects you create
	// err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{},
	// })
	// if err != nil {
	//	return err
	// }

	return nil
}

var _ reconcile.Reconciler = &Reconcile{{ .Resource.Kind }}{}

// Reconcile{{ .Resource.Kind }} reconciles a {{ .Resource.Kind }} object
type Reconcile{{ .Resource.Kind }} struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a {{ .Resource.Kind }} object and makes changes based on the state read
// and what is in the {{ .Resource.Kind }}.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
{{ if .Resource.CreateExampleReconcileBody -}}
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
{{ end -}}
// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }},verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups={{.GroupDomain}},resources={{ .Plural }}/status,verbs=get;update;patch
func (r *Reconcile{{ .Resource.Kind }}) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the {{ .Resource.Kind }} instance
	instance := &{{ .Resource.Group}}{{ .Resource.Version }}.{{ .Resource.Kind }}{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	{{ if .Resource.CreateExampleReconcileBody -}}
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-deployment",
			Namespace: {{ if .Resource.Namespaced}}instance.Namespace{{ else }}"default"{{ end }},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	{{ end -}}

	return reconcile.Result{}, nil
}
`
