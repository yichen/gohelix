package gohelix

import (
	"fmt"
)

// KeyBuilder geenrate a Zookeeper path
type KeyBuilder struct {
	ClusterID string
}

func (k *KeyBuilder) cluster() string {
	return fmt.Sprintf("/%s", k.ClusterID)
}

func (k *KeyBuilder) clusterConfig() string {
	return fmt.Sprintf("/%s/CONFIGS/CLUSTER/%s", k.ClusterID, k.ClusterID)
}

func (k *KeyBuilder) externalView() string {
	return fmt.Sprintf("/%s/EXTERNALVIEW", k.ClusterID)
}

func (k *KeyBuilder) externalViewForResource(resource string) string {
	return fmt.Sprintf("/%s/EXTERNALVIEW/%s", k.ClusterID, resource)
}

func (k *KeyBuilder) propertyStore() string {
	return fmt.Sprintf("/%s/PROPERTYSTORE", k.ClusterID)
}

func (k *KeyBuilder) controller() string {
	return fmt.Sprintf("/%s/CONTROLLER", k.ClusterID)
}

func (k *KeyBuilder) controllerErrors() string {
	return fmt.Sprintf("/%s/CONTROLLER/ERRORS", k.ClusterID)
}

func (k *KeyBuilder) controllerHistory() string {
	return fmt.Sprintf("/%s/CONTROLLER/HISTORY", k.ClusterID)
}

func (k *KeyBuilder) controllerMessages() string {
	return fmt.Sprintf("/%s/CONTROLLER/MESSAGES", k.ClusterID)
}

func (k *KeyBuilder) controllerStatusUpdates() string {
	return fmt.Sprintf("/%s/CONTROLLER/STATUSUPDATES", k.ClusterID)
}

func (k *KeyBuilder) idealStates() string {
	return fmt.Sprintf("/%s/IDEALSTATES", k.ClusterID)
}

func (k *KeyBuilder) idealStateForResource(resource string) string {
	return fmt.Sprintf("/%s/IDEALSTATES/%s", k.ClusterID, resource)
}

func (k *KeyBuilder) resourceConfig(resource string) string {
	return fmt.Sprintf("/%s/CONFIGS/RESOURCE/%s", k.ClusterID, resource)
}

func (k *KeyBuilder) participantConfigs() string {
	return fmt.Sprintf("/%s/CONFIGS/PARTICIPANT", k.ClusterID)
}

func (k *KeyBuilder) participantConfig(participantID string) string {
	return fmt.Sprintf("/%s/CONFIGS/PARTICIPANT/%s", k.ClusterID, participantID)
}

func (k *KeyBuilder) liveInstances() string {
	return fmt.Sprintf("/%s/LIVEINSTANCES", k.ClusterID)
}

func (k *KeyBuilder) instances() string {
	return fmt.Sprintf("/%s/INSTANCES", k.ClusterID)
}

func (k *KeyBuilder) instance(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s", k.ClusterID, participantID)
}

func (k *KeyBuilder) liveInstance(partipantID string) string {
	return fmt.Sprintf("/%s/LIVEINSTANCES/%s", k.ClusterID, partipantID)
}

func (k *KeyBuilder) currentStates(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/CURRENTSTATES", k.ClusterID, participantID)
}

func (k *KeyBuilder) currentStatesForSession(participantID string, sessionID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/CURRENTSTATES/%s", k.ClusterID, participantID, sessionID)
}

func (k *KeyBuilder) currentStateForResource(participantID string, sessionID string, resourceID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/CURRENTSTATES/%s/%s", k.ClusterID, participantID, sessionID, resourceID)
}

func (k *KeyBuilder) errorsR(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/ERRORS", k.ClusterID, participantID)
}

func (k *KeyBuilder) errors(participantID string, sessionID string, resourceID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/ERRORS/%s/%s", k.ClusterID, participantID, sessionID, resourceID)
}

func (k *KeyBuilder) healthReport(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/HEALTHREPORT", k.ClusterID, participantID)
}

func (k *KeyBuilder) statusUpdates(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/STATUSUPDATES", k.ClusterID, participantID)
}

func (k *KeyBuilder) stateModels() string {
	return fmt.Sprintf("/%s/STATEMODELDEFS", k.ClusterID)
}

func (k *KeyBuilder) stateModel(resourceID string) string {
	return fmt.Sprintf("/%s/STATEMODELDEFS/%s", k.ClusterID, resourceID)
}

func (k *KeyBuilder) messages(participantID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/MESSAGES", k.ClusterID, participantID)
}

func (k *KeyBuilder) message(participantID string, messageID string) string {
	return fmt.Sprintf("/%s/INSTANCES/%s/MESSAGES/%s", k.ClusterID, participantID, messageID)
}
