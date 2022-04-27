package services

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
)

type Confluent struct {
	ClusterName string
	Environment string
	Log         logr.Logger
}

func NewConfluentApi(ClusterName string, Environment string, Log logr.Logger) *Confluent {
	return &Confluent{
		ClusterName: ClusterName,
		Environment: Environment,
		Log:         Log,
	}
}

func (c *Confluent) GetEnvironments() (*[]messagesv1alpha1.Environment, error) {

	c.Log.Info("func::GetEnvironments, Getting confluent cloud environments")

	cmd := exec.Command("/bin/confluent", "environment", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		c.Log.Error(err, "func::GetEnvironments, Error to retrive clonfluent cloud environments")
		return &[]messagesv1alpha1.Environment{}, err
	} else {
		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		environments := []messagesv1alpha1.Environment{}

		json.Unmarshal([]byte(message), &environments)

		c.Log.Info("func::GetEnvironments, Success to get confluent cloud environments")

		return &environments, nil
	}
}

func (c *Confluent) SelectEnvironment(envs []messagesv1alpha1.Environment) (string, bool) {
	c.Log.Info("func::SelectEnvironment, Selecting confluent cloud environment")

	var envId string

	for i := 0; i < len(envs); i++ {
		if envs[i].Name == c.Environment {
			envId = envs[i].Id
			break
		}
	}

	return envId, c.SetEnvironment(envId)
}

func (c *Confluent) SetEnvironment(environmentId string) bool {
	c.Log.Info("func::setEnvironment, Setting confluent cloud environment")

	cmd := exec.Command("/bin/confluent", "environment", "use", environmentId)

	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

func (c *Confluent) GetKafkaCluster() (string, error) {
	c.Log.Info("func::GetKafkaCluster, Getting confluent cloud kafka clusters")

	cmd := exec.Command("/bin/confluent", "kafka", "cluster", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		c.Log.Error(err, "func::GetKafkaCluster, Error to retrive confluent cloud kafka clusters")
		return string(""), err
	} else {

		var clusterId string

		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		clusters := []messagesv1alpha1.Cluster{}

		json.Unmarshal([]byte(message), &clusters)

		for i := 0; i < len(clusters); i++ {
			if c.ClusterName == clusters[i].Name {
				clusterId = clusters[i].Id
				break
			}
		}

		c.Log.Info("func::GetKafkaCluster, Success to get confluent cloud kafka cluster")
		return clusterId, nil
	}
}

func (c *Confluent) GetSechemaRegistry(environmentId string) (string, error) {

	cmd := exec.Command("/bin/confluent", "schema-registry", "cluster", "describe", "--environment", environmentId, "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		return string(""), err
	} else {
		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		registry := messagesv1alpha1.SchemaRegistry{}

		json.Unmarshal([]byte(message), &registry)

		return registry.EndpointUrl, nil
	}
}

func (c *Confluent) NewTopic(topic *messagesv1alpha1.Topic) (*messagesv1alpha1.TopicReference, error) {

	c.Log.Info("Confluent::NewTopic:: Trying to create kafka topic")

	topicName := c.topicNameGenerator(topic.Tenant, topic.Namespace, topic.Topic)

	c.Log.Info("Confluent::NewTopic:: TopicName " + topicName)

	isOk, err := c.existsTopicName(topic.KCReference.ClusterId, topicName)

	c.Log.Info("Confluent::NewTopic:: The ttopic was found? " + strconv.FormatBool(!isOk))

	if !isOk && err == nil {

		c.Log.Info("Confluent::NewTopic:: Run the command to create kafka topic")

		c.Log.Info("Confluent::NewTopic:: topicName " + topicName)
		c.Log.Info("Confluent::NewTopic:: partitions " + topic.Partitions)
		c.Log.Info("Confluent::NewTopic:: Run the command to create kafka topic " + topic.KCReference.ClusterId)

		cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", topicName, "--partitions", topic.Partitions, "--cluster", topic.KCReference.ClusterId)

		cmdOutput := &bytes.Buffer{}
		cmd.Stdout = cmdOutput

		if err := cmd.Run(); err != nil {
			c.Log.Error(err, "Confluent::NewTopic:: Error to run the command to create kafka topic")
			return &messagesv1alpha1.TopicReference{}, err
		} else {

			topicReference := messagesv1alpha1.TopicReference{}
			topicReference.Topic = topicName

			if schemaRegistry, srError := c.GetSechemaRegistry(topic.KCReference.EnvironmentId); srError == nil {
				c.Log.Info("Selecting the EndpointURL")
				topicReference.EndpointUrl = schemaRegistry
			}

			return &topicReference, nil
		}
	} else {
		return &messagesv1alpha1.TopicReference{}, err
	}
}

func (c *Confluent) existsTopicName(clusterId string, topicName string) (bool, error) {

	cmd := exec.Command("/bin/confluent", "kafka", "topic", "list", "--cluster", clusterId)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		//c.Log.Error(err, "Falied to verify the existence of topic")
		return false, err
	} else {

		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		return regexp.MatchString("\\b"+topicName+"\\b", message)
	}
}

func (c *Confluent) getOutput(outs []byte) (string, bool) {
	if len(outs) > 0 {
		return string(outs), true
	}
	return string(""), false
}

func (c *Confluent) topicNameGenerator(tenant string, namespace string, topicName string) string {
	separator := "-"
	if tenant != string("") {
		return tenant + separator + namespace + separator + topicName
	}
	return namespace + separator + topicName
}
