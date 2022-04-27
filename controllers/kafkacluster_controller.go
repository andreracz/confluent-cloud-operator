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
	"encoding/json"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
	"github.com/kubbee/confluent-cloud-operator/services"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaClusterReconciler reconciles a KafkaCluster object
type KafkaClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

type KafkaClusterSecret struct {
	ClusterId     string `json:"clusterId"`
	EnvironmentId string `json:"environmentId"`
	Tenant        string `json:"tenant"`
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
	//get the logs
	log := ctrllog.FromContext(ctx)

	start := time.Now()

	//log := r.Log.WithValues("KafkaCluster", req.NamespacedName)

	kafkacluster := &messagesv1alpha1.KafkaCluster{}

	if err := r.Get(ctx, req.NamespacedName, kafkacluster); err != nil {
		if k8sErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KafkaCluster resource not found. Ignoring since object must be deleted")

			// Kubernets Event Stream
			//r.Recorder.Event(kafkacluster, corev1.EventTypeWarning, "ResourceNotFound", "Error resource KafkaCluster was not found.")

			return buildResult(false, start), nil
		}

		log.Error(err, "Failed to get KafkaCluster")

		// Kubernets Event Stream
		//r.Recorder.Event(kafkacluster, corev1.EventTypeWarning, "Error", "Error to get KafkaCluster. "+err.Error())

		return buildResult(false, start), err

	} else {
		// Kubernets Event Stream
		//r.Recorder.Event(kafkacluster, corev1.EventTypeWarning, "StatusConnection", "Connecting in the Confluent Cloud.")

		log.Info("Trying to connect in confluent cloud.")

		// Instancing Confluent Cloud API
		ccloud := services.NewConfluentGateway(kafkacluster.Spec.ClusterName, kafkacluster.Spec.Environment, log)

		if kcReference, gcrError := ccloud.GetClusterReference(); gcrError != nil {
			log.Error(gcrError, "Error to get Clueter Reference")
			// build and return the error
			return buildResult(false, start), gcrError
		} else {

			kafkaClusterSecret := KafkaClusterSecret{
				ClusterId:     kcReference.ClusterId,
				EnvironmentId: kcReference.EnvironmentId,
				Tenant:        kafkacluster.Spec.Tenant,
			}

			secret := r.specKafkaClusterSecret(kafkaClusterSecret, kafkacluster)

			s, _ := json.Marshal(secret)
			log.Info(string(s))

			found := &corev1.Secret{}
			err = r.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)

			if err != nil && k8sErrors.IsNotFound(err) { //Create secret

				err = r.Create(ctx, secret)

				if err != nil {
					// Error creating secret. Wait until it is fixed.
					return buildResult(false, start), err
				}

				log.Info("Secret created", "Name", "Namespace", secret.Name, secret.Namespace)
				return buildResult(true, start), nil
			}

			j, _ := json.Marshal(kafkaClusterSecret)
			log.Info(string(j))

			return buildResult(true, start), nil
		}
	}
}

// specKafkaClusterSecret returns the struct Secret to creation
func (r *KafkaClusterReconciler) specKafkaClusterSecret(kcs KafkaClusterSecret, kc *messagesv1alpha1.KafkaCluster) *corev1.Secret {

	var labels = make(map[string]string)
	labels["name"] = kc.Name
	labels["owner"] = "kafkacluster-controller"

	var immutable bool = false

	// create and return secret object.
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kc.Name,
			Namespace: kc.Namespace,
			Labels:    labels,
		},
		Type:      "KafkaCluster/kafkacluster-controller",
		Data:      map[string][]byte{"tenant": []byte(kcs.Tenant), "clusterId": []byte(kcs.ClusterId), "environmentId": []byte(kcs.EnvironmentId)},
		Immutable: &immutable,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaCluster{}).
		Complete(r)
}
