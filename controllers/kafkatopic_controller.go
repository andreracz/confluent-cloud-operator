/*
Copyright 2022 Luiz H. de Sousa Ribas.

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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
)

// KafkaTopicReconciler reconciles a KafkaTopic object
type KafkaTopicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkatopics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaTopic object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaTopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Duration of the Reconcile execution
	start := time.Now()

	_ = log.FromContext(ctx)

	log := r.Log.WithValues("KafkaTopic", req.NamespacedName)

	kafktopic := &messagesv1alpha1.KafkaTopic{}

	err := r.Get(ctx, req.NamespacedName, kafktopic)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KafkaTopic resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KafkaTopic")
		return ctrl.Result{}, err
	}

	ccloudT := NewConfluentApi("default", "default")

	//confluent kafka topic create users --partitions 3  --cluster lkc-57wnz2
	if environments, eErr := ccloudT.GetEnvironments(); eErr == nil {
		//
		log.Info("Get the Confluent Cloud Environments", environments)
		//
		if ccloudT.SetEnvironment(environments) {
			//
			log.Info("The environment was choosed")
			//
			if clusterId, cErr := ccloudT.GetKafkaCluster(); cErr == nil {
				//
				log.Info("Creating Topic on the Confluent Cloud")

				cTopic := CreationTopic{
					Tenant:     "",
					Namespace:  req.NamespacedName.String(),
					Partitions: fmt.Sprint(kafktopic.Spec.Partitions),
					ClusterId:  clusterId,
					TopicName:  kafktopic.Spec.TopicName,
				}

				status, tErr := ccloudT.NewTopic(cTopic)

				if !status {
					log.Info("The TopicName was created")

					//
					return buildResult(status, time.Since(start)), nil
				} else {
					log.Error(tErr, "Error to create the Topicname")

					//
					return buildResult(status, time.Since(start)), tErr
				}
			} else {
				log.Error(cErr, "Error to create the Topicname")

				//
				return buildResult(false, time.Since(start)), cErr
			}
		}
	} else {
		log.Error(eErr, "Error to create the Topicname")

		//
		return buildResult(false, time.Since(start)), eErr
	}

	return ctrl.Result{}, err
}

func buildResult(requeue bool, requeueAfter time.Duration) ctrl.Result {
	return ctrl.Result{
		Requeue:      requeue,
		RequeueAfter: requeueAfter,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaTopic{}).
		Complete(r)
}
