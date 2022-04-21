package controllers

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"regexp"

	"github.com/go-logr/logr"
	messagesv1alpha1 "github.com/kubbee/confluent-cloud-operator/api/v1alpha1"
)

type Confluent struct {
	ClusterName string
	Environment string
	Log         logr.Logger
}

type CreationTopic struct {
	Tenant     string
	Namespace  string
	TopicName  string
	Partitions string
	ClusterId  string
}

func NewConfluentApi(clusterName string, environment string) *Confluent {
	return &Confluent{
		ClusterName: clusterName,
		Environment: environment,
	}
}

//
func (c *Confluent) GetEnvironments() ([]messagesv1alpha1.Environment, error) {

	cmd := exec.Command("/bin/confluent", "environment", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		return []messagesv1alpha1.Environment{}, err
	} else {
		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		environments := []messagesv1alpha1.Environment{}
		json.Unmarshal([]byte(message), &environments)

		return environments, nil
	}
}

//
func (c *Confluent) SelectEnvironment(envs []messagesv1alpha1.Environment) string {
	var status string
	for i := 0; i < len(envs); i++ {
		if envs[i].Name == c.Environment {
			status = envs[i].Id
			break
		}
	}
	return status
}

//
func (c *Confluent) SetEnvironment(environmentId string) bool {

	cmd := exec.Command("/bin/confluent", "environment", "use", environmentId)

	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

//
func (c *Confluent) GetKafkaCluster() (string, error) {

	cmd := exec.Command("/bin/confluent", "kafka", "cluster", "list", "--output", "json")

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err := cmd.Run(); err != nil {
		return string(""), err
	} else {

		var id string

		output := cmdOutput.Bytes()
		message, _ := c.getOutput(output)

		clusters := []messagesv1alpha1.Cluster{}
		json.Unmarshal([]byte(message), &clusters)

		for i := 0; i < len(clusters); i++ {
			if c.ClusterName == clusters[i].Name {
				id = clusters[i].Id
				break
			}
		}

		return id, nil
	}
}

//
func (c *Confluent) NewTopic(cTopic CreationTopic) (bool, error) {

	topicName := c.topicNameGenerator(cTopic.Tenant, cTopic.Namespace, cTopic.TopicName)

	isOk, err := c.existsTopicName(cTopic.ClusterId, topicName)

	if !isOk && err == nil {

		cmd := exec.Command("/bin/confluent", "kafka", "topic", "create", topicName, "--partitions", cTopic.Partitions, "--cluster", cTopic.ClusterId)

		cmdOutput := &bytes.Buffer{}
		cmd.Stdout = cmdOutput

		if err := cmd.Run(); err != nil {
			//c.Log.Error(err, "Falied to create topic")
			return false, err
		} else {

			output := cmdOutput.Bytes()
			message, _ := c.getOutput(output)

			return len(message) > 0, nil
		}
	}
	return false, err
}

//
func (c *Confluent) existsTopicName(cluster string, topicName string) (bool, error) {

	cmd := exec.Command("/bin/confluent", "kafka", "topic", "list", "--cluster", cluster)

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

//
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
