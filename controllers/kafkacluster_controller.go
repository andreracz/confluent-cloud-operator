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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// KafkaClusterReconciler reconciles a KafkaCluster object
type KafkaClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	recorder record.EventRecorder
}

type KafkaClusterSecret struct {
	ClusterId     string
	EnvironmentId string
	Tenant        string
}

//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=messages.kubbee.tech,resources=kafkaclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KafkaCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KafkaClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	start := time.Now()

	_ = log.FromContext(ctx)

	log := r.Log.WithValues("KafkaCluster", req.NamespacedName)

	kafkacluster := &messagesv1alpha1.KafkaCluster{}

	if err := r.Get(ctx, req.NamespacedName, kafkacluster); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KafkaCluster resource not found. Ignoring since object must be deleted")

			// Kubernets Event Stream
			r.recorder.Event(kafkacluster, corev1.EventTypeWarning, "ResourceNotFound", "Error resource KafkaCluster was not found.")

			return buildResult(false, start), nil
		}

		log.Error(err, "Failed to get KafkaCluster")

		// Kubernets Event Stream
		r.recorder.Event(kafkacluster, corev1.EventTypeWarning, "Error", "Error to get KafkaCluster. "+err.Error())

		return buildResult(false, start), err

	} else {
		// Kubernets Event Stream
		r.recorder.Event(kafkacluster, corev1.EventTypeWarning, "StatusConnection", "Connection in the Confluent Cloud.")

		// Instancing Confluent Cloud API
		ccloudT := NewConfluentApi(kafkacluster.Spec.ClusterName, kafkacluster.Spec.Environment)

		//Get the cloud confluent environments
		environments, eErr := ccloudT.GetEnvironments()

		if eErr == nil {
			// Kubernets Event Stream
			r.recorder.Event(kafkacluster, corev1.EventTypeNormal, "Successfuly", "Selecting the Confluent Cloud Environment")

			//Selecting the Cloud Confluent Environment
			envId := ccloudT.SelectEnvironment(environments)

			//Selecting the Kafka Cluster
			if clusterId, cErr := ccloudT.GetKafkaCluster(); cErr == nil {

				// Kubernets Event Stream
				r.recorder.Event(kafkacluster, corev1.EventTypeNormal, "Successfuly", "Selecting the Kafka Cluster.")

				// Instancing kafkaClusterSecret struct
				kafkaClusterSecret := KafkaClusterSecret{
					ClusterId:     clusterId,
					EnvironmentId: envId,
					Tenant:        kafkacluster.Spec.Tenant,
				}

				// Generating the k8s Secret Resource
				if secret, skcsError := r.specKafkaClusterSecret(kafkaClusterSecret, kafkacluster); skcsError == nil {

					// Creating the k8s Secret Resource
					r.Create(ctx, secret)

					// Kubernets Event Stream
					r.recorder.Event(kafkacluster, corev1.EventTypeNormal, "Successfuly", "The secret with cluster information was created.")

					//
					return buildResult(true, start), nil
				} else {
					log.Error(cErr, "Error to create secret resource")
					//
					return buildResult(false, start), skcsError
				}

			} else {
				log.Error(cErr, "Error to create the Topicname")
				//
				return buildResult(false, start), cErr
			}
		}
		//
		return buildResult(false, start), nil
	}
}

// specKafkaClusterSecret returns the struct Secret to creation
func (r *KafkaClusterReconciler) specKafkaClusterSecret(kcs KafkaClusterSecret, kc *messagesv1alpha1.KafkaCluster) (*corev1.Secret, error) {

	data := make(map[string][]byte)

	data["tenant"] = []byte(string(kcs.Tenant))
	data["clusterId"] = []byte(string(kcs.ClusterId))
	data["environmentId"] = []byte(string(kcs.EnvironmentId))

	//
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kc.Name,
			Namespace: kc.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	return secret, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaCluster{}).
		Complete(r)
}
