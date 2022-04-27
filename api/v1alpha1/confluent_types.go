package v1alpha1

type Environment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type KafkaClusterReference struct {
	ClusterId     string `json:"clusterId"`
	EnvironmentId string `json:"environmentId"`
}

type Cluster struct {
	Availability string `json:"availability"` //"availability": "single-zone",
	Id           string `json:"id"`           //"id": "lkc-nvywwv",
	Name         string `json:"name"`         //"name": "kubber2",
	Provider     string `json:"provider"`     //"provider": "aws",
	Region       string `json:"region"`       //"region": "us-east-2",
	Status       string `json:"status"`       //"status": "UP",
	Type         string `json:"type"`         //"type": "BASIC"
}

type Topic struct {
	Tenant      string                `json:"tenant"`
	Namespace   string                `json:"namespace"`
	Topic       string                `json:"topic"`
	Partitions  string                `json:"partitions"`
	KCReference KafkaClusterReference `json:"kcReference"`
}

type TopicReference struct {
	Topic       string `json:"topic"`
	EndpointUrl string `json:"endpointUrl"`
}

type SchemaRegistry struct {
	Name                string `json:"name"`
	ClusterId           string `json:"cluster_id"`
	EndpointUrl         string `json:"endpoint_url"`
	UsedSchemas         string `json:"used_schemas"`
	AvailableSchemas    string `json:"available_schemas"`
	GlobalCompatibility string `json:"global_compatibility"`
	Mode                string `json:"mode"`
	ServiceProvider     string `json:"service_provider"`
}
