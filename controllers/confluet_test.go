package controllers

import (
	"fmt"
	"testing"
)

type KafkaClusterTest struct {
	TopicName   string
	Partitions  string
	Cluster     string
	Environment string
}

func TestCreateTopic(t *testing.T) {

	ccloudT := NewConfluentApi("kubber2", "production")

	//confluent kafka topic create users --partitions 3  --cluster lkc-57wnz2
	if environments, err := ccloudT.GetEnvironments(); err == nil {
		if ccloudT.SetEnvironment(environments) {
			if clusterId, err := ccloudT.GetKafkaCluster(); err == nil {

				cTopic := CreationTopic{
					Tenant:     "TIM",
					Namespace:  "PAGAMENTOS",
					Partitions: fmt.Sprint(6),
					ClusterId:  clusterId,
					TopicName:  "CADASTRO-DE-PAGAMENTOS",
				}

				status, _ := ccloudT.NewTopic(cTopic)

				if !status {
					t.Fail()
				}
			}
		}
	}
}
