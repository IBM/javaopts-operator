/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cpev1alpha1 "github.com/IBM/knative-quarkus-bench/api/v1alpha1"
)

var log = logf.Log.WithName("jvm opts operator")

// JavaOptsReconciler reconciles a JavaOpts object
type JavaOptsReconciler struct {
	*kubernetes.Clientset
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cpe.benchmark.io,resources=javaopts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpe.benchmark.io,resources=javaopts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cpe.benchmark.io,resources=javaopts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JavaOpts object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *JavaOptsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling JavaOpts")
	// TODO(user): your logic here
	instance := &cpev1alpha1.JavaOpts{}

	if instance.Status.CompletionTime != nil {
		return ctrl.Result{}, nil
	}

	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			//Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue

			return ctrl.Result{}, nil
		}

		reqLogger.Info(fmt.Sprintf("Cannot get #%v", err))
		//Error reading the object - requeue the request
		return ctrl.Result{}, err
	}

	//Check the existence of configmap
	namespace := instance.Namespace
	cmName := instance.Spec.ConfigMapName

	c, err := r.Clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})

	if err != nil {
		reqLogger.Info(fmt.Sprintf("The error of getting configmaps, err:%v", err))
	}
	reqLogger.Info(fmt.Sprintf("The status of getting configmap %s", c))

	if err != nil && errors.IsNotFound(err) {
		//If it doesn't exist, create a configmap
		cm := r.createNewConfigMap(instance)
		c, err = r.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metav1.CreateOptions{})

		if err != nil {
			reqLogger.Info(fmt.Sprintf("The error of creating configmaps #%v", err))
		} else {
			reqLogger.Info(fmt.Sprintf("Create new configMap #%s, err=%v", cmName, err))
			//update benchmark deployment
			r.updateDeployment(instance)
			reqLogger.Info(fmt.Sprintf("Update serverless benchmark deployment %s", instance.Spec.RevName))
		}

	} else if err != nil {
		reqLogger.Info(fmt.Sprintf("Cannot get configMap #%v", err))
	} else {
		//If it exists, update the configmap
		cm := r.createNewConfigMap(instance)

		c, err = r.Clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
		reqLogger.Info(fmt.Sprintf("Update configmap #%s, err=%v", cmName, err))

		//update benchmark deployment
		r.updateDeployment(instance)
		reqLogger.Info(fmt.Sprintf("Update serverless benchmark deployment %s", instance.Spec.RevName))
	}

	//create driver job
	continueFlag := false
	found, err := r.Clientset.BatchV1().Jobs(namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
	reqLogger.Info(fmt.Sprintf("Getting job #%s", found))
	if err != nil && errors.IsNotFound(err) {
		job := r.createDriverJob(instance)
		_, err := r.Clientset.BatchV1().Jobs(instance.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})

		r.updateCondition(instance, found.Status)
		continueFlag = true
		if err != nil {
			reqLogger.Info(fmt.Sprintf("The error of creating job %v", err))
			return ctrl.Result{}, err
		}
	} else if err != nil {
		reqLogger.Info(fmt.Sprintf("Cannot get Job %v", err))
		return ctrl.Result{}, err
	} else if err == nil {
		//if it's found, check the completion status and change the status of JavaOpts
		r.updateCondition(instance, found.Status)
	}

	if instance.Status.CompletionTime != nil {
		continueFlag = false
	}

	//update javaOpts status if created job become completed
	for continueFlag {
		found, err = r.Clientset.BatchV1().Jobs(instance.Namespace).Get(context.TODO(), req.Name, metav1.GetOptions{})
		cond := found.Status.Conditions
		for i := 0; i < len(cond); i++ {
			status := cond[i].Status
			tp := cond[i].Type

			if tp == batchv1.JobComplete && status == corev1.ConditionTrue {
				r.updateCondition(instance, found.Status)
				continueFlag = false
				break
			} else if tp == batchv1.JobFailed {
				r.updateCondition(instance, found.Status)
				continueFlag = false
				break
			} else if tp == batchv1.JobSuspended {
				r.updateCondition(instance, found.Status)
				continueFlag = false
				break
			}
		}

		time.Sleep(time.Second * 5)
	}

	//reqLogger.Info(fmt.Sprintf("JavaOpts Condition %v", instance.Status))

	return ctrl.Result{}, nil
}

func (r *JavaOptsReconciler) updateDeployment(instance *cpev1alpha1.JavaOpts) error {
	reqLogger := log.WithName("Update Deployment")

	depName := instance.Spec.RevName + "-deployment"
	dep, err := r.Clientset.AppsV1().Deployments(instance.Namespace).Get(context.TODO(), depName, metav1.GetOptions{})
	annots := dep.ObjectMeta.Annotations
	if annots == nil {
		annots = make(map[string]string)
	}
	annots["job-version"] = instance.Name
	newDep := dep.DeepCopy()
	newDep.ObjectMeta.Annotations = annots
	d, err := r.Clientset.AppsV1().Deployments(instance.Namespace).Update(context.TODO(), newDep, metav1.UpdateOptions{})

	if err != nil {
		reqLogger.Info(fmt.Sprintf("Cannot update %v", err))
	} else {
		reqLogger.Info(fmt.Sprintf("Update deployment %s, %v", depName, d.ObjectMeta.Annotations))
	}

	podList, err := r.Clientset.CoreV1().Pods(instance.Namespace).List(context.TODO(), metav1.ListOptions{
		//LabelSelector: "apps=benchmark",
		LabelSelector: "app=" + instance.Spec.RevName,
	})

	//delete pods to restart
	for _, pod := range podList.Items {
		err := r.Clientset.CoreV1().Pods(instance.Namespace).Delete(context.TODO(), pod.ObjectMeta.Name, metav1.DeleteOptions{})

		if err != nil {
			reqLogger.Info(fmt.Sprintf("Cannot delete %v", err))
		}
	}

	return nil
}

func (r *JavaOptsReconciler) updateCondition(instance *cpev1alpha1.JavaOpts, js batchv1.JobStatus) error {
	reqLogger := log.WithName("Update Condition")

	condition := batchv1.JobCondition{}
	jc := js.Conditions
	now := metav1.Now()

	if instance.Status.CompletionTime != nil {
		return nil
	}

	if len(jc) > 0 {
		jtype := jc[len(jc)-1].Type
		cs := jc[len(jc)-1].Status
		ttime := jc[len(jc)-1].LastTransitionTime

		condition = batchv1.JobCondition{
			Type:   jtype,
			Status: cs,
		}
		if jtype == batchv1.JobComplete && cs == corev1.ConditionTrue {
			instance.Status.CompletionTime = &ttime
		}

	} else {

		instance.Status.StartTime = &now
	}

	instance.Status.Conditions = append(instance.Status.Conditions, condition)
	reqLogger.Info(fmt.Sprintf("Current status: %v", instance.Status.Conditions))
	err := r.Client.Status().Update(context.Background(), instance)
	if err != nil {
		reqLogger.Info(fmt.Sprintf("Cannot update %v", err))
	}

	return nil
}

func (r *JavaOptsReconciler) createDriverJob(instance *cpev1alpha1.JavaOpts) *batchv1.Job {

	//command := []string{"echo", "Please input any commands."}
	image := "fedora:33"
	cmd := []string{"/bin/bash", "-c"}
	args := []string{"echo 'Please input any commands"}
	arg := args[0]
	if len(instance.Spec.Command) > 0 {
		args[0] = instance.Spec.Command
		args = append(args, "source ./log.sh")
		arg = strings.Join(args, ";")
	}

	if len(instance.Spec.DriverImage) > 0 {
		image = instance.Spec.DriverImage
	}

	envs := make([]corev1.EnvVar, 1)
	envs[0] = corev1.EnvVar{
		Name:  "LABEL",
		Value: instance.Spec.RevName,
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "driver",
							Image:   image,
							Env:     envs,
							Command: cmd,
							Args:    []string{arg},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	//Set instance as the owner and controller
	ctrl.SetControllerReference(instance, job, r.Scheme)

	return job
}

func (r *JavaOptsReconciler) createNewConfigMap(instance *cpev1alpha1.JavaOpts) *corev1.ConfigMap {

	namespace := instance.Namespace
	cmName := instance.Spec.ConfigMapName
	defaultOpts := instance.Spec.DefaultOpts

	//check s.spec and add values to default_opts
	params := []string{defaultOpts, instance.Spec.MaxHeap, instance.Spec.MinHeap, instance.Spec.GCType, instance.Spec.GCThreads, instance.Spec.Escape}
	opts := strings.Join(params, " ")
	log.Info(fmt.Sprintf("JVM options: %s", opts))
	configMapData := map[string]string{"JAVA_OPTIONS": opts}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
		},
		Data: configMapData,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *JavaOptsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpev1alpha1.JavaOpts{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
