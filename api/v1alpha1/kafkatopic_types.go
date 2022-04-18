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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KafkaTopicSpec defines the desired state of KafkaTopic
type KafkaTopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TopicName  string `json:"topic-name"`
	Partitions int32  `json:"partitions"`
	//DeleteRetentionMS               int64  `json:"delete-retentin-ms,omitempty"`
	//MaxMessageBytes                 int64  `json:"max-message-bytes,omitempty"`
	//MaxCompactionLagMS              int64  `json:"max-compaction-lag-ms,omitempty"`
	//MessageTimestampDifferenceMaxMS int64  `json:"message-timestamp-difference-max-ms,omitempty"`
	//MinCompactionLagMS              int64  `json:"min-compaction-lag-ms,omitempty"`
	//MinInsyncReplicas               int64  `json:"min-insync-replicas,omitempty"`
	//RetentionBytes                  int64  `json:"retention-bytes,omitempty`
	//RetentionMS                     int64  `json:"retention-ms,omitempty"`
	//SegmentBytes                    int64  `json:"segment-bytes,omitempty"`
	//SegmentMS                       int64  `json:"segment-ms,omitempty"`
}

// KafkaTopicStatus defines the observed state of KafkaTopic
type KafkaTopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaTopic is the Schema for the kafkatopics API
type KafkaTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaTopicSpec   `json:"spec,omitempty"`
	Status KafkaTopicStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaTopicList contains a list of KafkaTopic
type KafkaTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaTopic{}, &KafkaTopicList{})
}
