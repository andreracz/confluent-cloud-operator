package services

import (
	"errors"
	"strconv"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
)

type ConfluentGateway struct {
	ClusterName string
	Environment string
	Log         logr.Logger
}

func NewConfluentGateway(ClusterName string, Environment string, Log logr.Logger) *ConfluentGateway {
	return &ConfluentGateway{
		ClusterName: ClusterName,
		Environment: Environment,
		Log:         Log,
	}
}

func (c *ConfluentGateway) GetClusterReference() (*messagesv1alpha1.KafkaClusterReference, error) {

	c.Log.Info("Trying to get the environments.")

	ccli := NewConfluentApi(c.ClusterName, c.Environment, c.Log)

	environments, err := ccli.GetEnvironments()

	if err != nil {
		return &messagesv1alpha1.KafkaClusterReference{}, err
	}

	environmentId, selected := ccli.SelectEnvironment(*environments)

	if selected {
		clusterId, err := ccli.GetKafkaCluster()

		if err != nil {
			return &messagesv1alpha1.KafkaClusterReference{}, err
		}

		return &messagesv1alpha1.KafkaClusterReference{
			ClusterId:     clusterId,
			EnvironmentId: environmentId,
		}, nil
	}

	return &messagesv1alpha1.KafkaClusterReference{}, nil
}

func (c *ConfluentGateway) NewTopic(topic *messagesv1alpha1.Topic) (*messagesv1alpha1.TopicReference, error) {

	ccli := NewConfluentApi(c.ClusterName, c.Environment, c.Log)

	c.Log.Info("Selecting Environment")
	c.Log.Info("Checking the Environment to select " + topic.KCReference.EnvironmentId)

	selected := ccli.SetEnvironment(topic.KCReference.EnvironmentId)

	c.Log.Info("Environment to selected? " + strconv.FormatBool(selected))

	if selected {

		c.Log.Info("Next step, create kafka topic")

		if topicReference, tErr := ccli.NewTopic(topic); tErr == nil {
			c.Log.Info("Creating New Topic")
			//
			return topicReference, nil
		} else {
			c.Log.Error(tErr, "Error to create kafka topic")
			return &messagesv1alpha1.TopicReference{}, tErr
		}

	} else {
		c.Log.Error(errors.New("error to select the environment"), "Error to select the environment")
		return &messagesv1alpha1.TopicReference{}, errors.New("error to select the environment")
	}

}
