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
	"strconv"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"errors"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
	"github.com/kubbee/confluent-cloud-operator/services"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaTopicReconciler reconciles a KafkaTopic object
type KafkaTopicReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

type ConnectionCredentials interface {
	Data(key string) ([]byte, bool)
}

type ClusterCredentials struct {
	data map[string][]byte
}

func (c ClusterCredentials) Data(key string) ([]byte, bool) {
	result, ok := c.data[key]
	return result, ok
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
	//get the logs
	log := ctrllog.FromContext(ctx)

	// Duration of the Reconcile execution
	start := time.Now()

	kafkatopic := &messagesv1alpha1.KafkaTopic{}

	if err := r.Get(ctx, req.NamespacedName, kafkatopic); err != nil {
		if k8sErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KafkaTopic resource not found. Ignoring since object must be deleted")

			// Kubernets Event Stream
			//r.Recorder.Event(kafkatopic, corev1.EventTypeWarning, "ResourceNotFound", "Error resource KafkaCluster was not found.")

			return buildResult(false, start), nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KafkaTopic")
		// Kubernets Event Stream
		//r.Recorder.Event(kafkatopic, corev1.EventTypeWarning, "Error", "Error to get KafkaTopic. "+err.Error())

		return buildResult(false, start), err
	} else {

		if connectionCreds, cErr := r.getSecret(ctx, req.NamespacedName.Namespace, "kafka-cluster-connection"); cErr != nil {
			//
			//r.Recorder.Event(kafkatopic, corev1.EventTypeWarning, "Error", "Error to get kafka-cluster-connection. "+err.Error())
			log.Error(err, "Failed to get Secret")
			return buildResult(false, start), err

		} else {

			cf := &corev1.ConfigMap{}
			err = r.Get(ctx, types.NamespacedName{Name: kafkatopic.Name, Namespace: kafkatopic.Namespace}, cf)

			if err != nil && k8sErrors.IsNotFound(err) { //Create ConfigMap

				log.Info("Trying to create kafka topic")

				tenant, isTenantOk := connectionCreds.Data("tenant")
				log.Info("Retrive Tenant" + string(tenant))

				clusterId, isClusterIdOk := connectionCreds.Data("clusterId")
				log.Info("Retrive clusterId" + string(clusterId))

				environmentId, isEnvironmentIdOk := connectionCreds.Data("environmentId")
				log.Info("Retrive environmentId" + string(environmentId))

				if isTenantOk && isClusterIdOk && isEnvironmentIdOk {

					log.Info("Success to get secret values")

					kcReference := messagesv1alpha1.KafkaClusterReference{
						ClusterId:     string(clusterId),
						EnvironmentId: string(environmentId),
					}

					Partitions := strconv.FormatInt(int64(kafkatopic.Spec.Partitions), 10)
					log.Info("Converting the partitions was converted to string? " + Partitions)

					topic := &messagesv1alpha1.Topic{
						Tenant:      string(tenant),
						Topic:       kafkatopic.Spec.TopicName,
						Partitions:  Partitions,
						Namespace:   req.NamespacedName.Namespace,
						KCReference: kcReference,
					}

					ccloud := services.NewConfluentGateway("", "", log)

					if topicRefence, topicError := ccloud.NewTopic(topic); topicError != nil {
						return buildResult(false, start), err
					} else {
						cfg := r.createConfigMap(*topicRefence, kafkatopic)

						err = r.Create(ctx, cfg)

						if err != nil {
							// Error creating secret. Wait until it is fixed.
							return buildResult(false, start), err
						}

						log.Info("ConfigMap created", "Name", "Namespace", kafkatopic.Name, kafkatopic.Namespace)
						return buildResult(true, start), nil
					}
				}
			}
		}

		return buildResult(false, start), nil
	}
}

func (r *KafkaTopicReconciler) getSecret(ctx context.Context, requestNamespace string, secretName string) (ConnectionCredentials, error) {
	log := ctrllog.FromContext(ctx)

	log.Info("requestNamespace --->>> " + requestNamespace)
	log.Info("secretName --->>> " + secretName)

	secret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Namespace: requestNamespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	return readCredentialsFromKubernetesSecret(secret)
}

func (r *KafkaTopicReconciler) createConfigMap(topicReference messagesv1alpha1.TopicReference, kt *messagesv1alpha1.KafkaTopic) *corev1.ConfigMap {

	var labels = make(map[string]string)
	labels["name"] = kt.Name
	labels["owner"] = "kafkacluster-controller"

	var immutable bool = false

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kt.Name,
			Namespace: kt.Namespace,
			Labels:    labels,
		},
		Data:      map[string]string{"topic.name": topicReference.Topic, "endpoint.url": topicReference.EndpointUrl},
		Immutable: &immutable,
	}
}

func readCredentialsFromKubernetesSecret(secret *corev1.Secret) (ConnectionCredentials, error) {
	if secret == nil {
		return nil, fmt.Errorf("unable to retrieve information from Kubernetes secret %s: %w", secret.Name, errors.New("nil secret"))
	}

	return ClusterCredentials{
		data: map[string][]byte{
			"tenant":        secret.Data["tenant"],
			"clusterId":     secret.Data["clusterId"],
			"environmentId": secret.Data["environmentId"],
		},
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaTopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagesv1alpha1.KafkaTopic{}).
		Complete(r)
}
