package v1alpha1

import (
	"github.com/knative/pkg/apis/duck"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

const (
	// KubernetesEventSourceConditionReady has status True when the source is ready to send events.
	KafkaEventSourceConditionReady = duckv1alpha1.ConditionReady
)

// Check that KubernetesEventSource can be validated and can be defaulted.
var _ runtime.Object = (*KafkaEventSource)(nil)

// Check that KubernetesEventSource implements the Conditions duck type.
var _ = duck.VerifyType(&KafkaEventSource{}, &duckv1alpha1.Conditions{})

var kafkaEventSourceCondSet = duckv1alpha1.NewLivingConditionSet()

// GetCondition returns the condition currently associated with the given type, or nil.
func (s *KafkaEventSourceStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return kafkaEventSourceCondSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *KafkaEventSourceStatus) IsReady() bool {
	return kafkaEventSourceCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *KafkaEventSourceStatus) InitializeConditions() {
	kafkaEventSourceCondSet.Manage(s).InitializeConditions()
}

// MarkReady sets the condition that the ContainerSource owned by
// the source has Ready status True.
func (s *KafkaEventSourceStatus) MarkReady() {
	kafkaEventSourceCondSet.Manage(s).MarkTrue(KafkaEventSourceConditionReady)
}

// MarkUnready sets the condition that the ContainerSource owned by
// the source does not have Ready status True.
func (s *KafkaEventSourceStatus) MarkUnready(reason, messageFormat string, messageA ...interface{}) {
	kafkaEventSourceCondSet.Manage(s).MarkFalse(KafkaEventSourceConditionReady, reason, messageFormat, messageA...)
}

// MarkSink sets the condition that the source has a sink configured.
func (s *KafkaEventSourceStatus) MarkSink(uri string) {
	s.SinkURI = uri
	if len(uri) > 0 {
		kafkaEventSourceCondSet.Manage(s).MarkTrue(KafkaEventSourceConditionReady)
	} else {
		kafkaEventSourceCondSet.Manage(s).MarkUnknown(KafkaEventSourceConditionReady, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *KafkaEventSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	kafkaEventSourceCondSet.Manage(s).MarkFalse(KafkaEventSourceConditionReady, reason, messageFormat, messageA...)
}

// KafkaEventSourceStatus defines the observed state of KafkaEventSource
type KafkaEventSourceStatus struct {

	// Conditions holds the state of a source at a point in time.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions duckv1alpha1.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +optional
	SinkURI string `json:"sinkUri,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaEventSource is the Schema for the kafkaeventsources API
// +k8s:openapi-gen=true
type KafkaEventSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaEventSourceSpec   `json:"spec,omitempty"`
	Status KafkaEventSourceStatus `json:"status,omitempty"`
}

// KafkaEventSourceSpec defines the desired state of KafkaEventSource
type KafkaEventSourceSpec struct {

	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Bootstrap string `json:"bootstrap"`
	Topic     string `json:"topic"`
	//+optional
	ConsumerGroupID string `json:"consumerGroupID"`
	//+optional
	Net KafkaEventSourceNet `json:"net,omitempty"`
	//+optional
	Consumer KafkaEventSourceConsumer `json:"consumer,omitempty"`

	//+optional
	ChannelBufferSize int64 `json:"channelBufferSize,omitempty"`

	//+optional
	Group KafkaEventSourceGroup `json:"group,omitempty"`

	// Sink is a reference to an object that will resolve to a domain name to use as the sink.
	// +optional
	Sink *corev1.ObjectReference `json:"sink,omitempty"`

	//+optional
	Replicas int64 `json:"replicas,omitempty"`
}

//KafkaEventSourceConsumer defines consumer related properties
type KafkaEventSourceConsumer struct {
	// +optional
	MaxWaitTime int64 `json:"maxWaitTime,omitempty"`
	// +optional
	MaxProcessingTime int64 `json:"maxProcessingTime,omitempty"`
	// +optional
	Offsets KafkaEventSourceOffsets `json:"offsets,omitempty"`
}

//KafkaEventSourceOffsets offsets information
type KafkaEventSourceOffsets struct {
	//+optional
	CommitInterval int64 `json:"commitInterval,omitempty"`
	//+optional
	InitialOffset string `json:"initial,omitempty"`
	//+optional
	Retention int64 `json:"retention,omitempty"`
	//+optional
	Retry KafkaEventSourceRetry `json:"retry,omitempty"`
}

//KafkaEventSourceRetry retry information
type KafkaEventSourceRetry struct {
	//+optional
	Max int64 `json:"max,omitempty"`
}

//KafkaEventSourceGroup group information
type KafkaEventSourceGroup struct {
	//+optional
	PartitionStrategy string `json:"partitionStrategy,omitempty"`
	//+optional
	Session KafkaEventSourceSession `json:"session,omitempty"`
}

//KafkaEventSourceSession session information
type KafkaEventSourceSession struct {
	//+optional
	Timeout int64 `json:"timeout,omitempty"`
}

// KafkaEventSourceNet defines network related properties
type KafkaEventSourceNet struct {
	MaxOpenRequests int64 `json:"maxOpenRequests"`
	KeepAlive       int64 `json:"keepAlive"`

	//+optional
	Sasl KafkaEventSourceSpecSasl `json:"sasl,omitempty"`
}

// KafkaEventSourceSpecSasl defines whether or not and how to use Sasl authentication
type KafkaEventSourceSpecSasl struct {
	Enable    bool   `json:"enable"`
	Handshake bool   `json:"handshake"`
	User      string `json:"user"`
	Password  string `json:"password"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaEventSourceList contains a list of KafkaEventSource
type KafkaEventSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaEventSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaEventSource{}, &KafkaEventSourceList{})
}
