package gohelix

import (
	"fmt"
	"sync"
)

// HelixManager manages the Helix client connections and roles
type HelixManager struct {
	zkAddress string
	conn      *connection
}

type (
	// ExternalViewChangeListener is triggered when the external view is updated
	ExternalViewChangeListener func(externalViews []*Record, context *Context)

	// LiveInstanceChangeListener is triggered when live instances of the cluster are updated
	LiveInstanceChangeListener func(liveInstances []*Record, context *Context)

	// CurrentStateChangeListener is triggered when the current state of a participant changed
	CurrentStateChangeListener func(instance string, currentState []*Record, context *Context)

	// IdealStateChangeListener is triggered when the ideal state changed
	IdealStateChangeListener func(idealState []*Record, context *Context)

	// InstanceConfigChangeListener is triggered when the instance configs are updated
	InstanceConfigChangeListener func(configs []*Record, context *Context)

	// ControllerMessageListener is triggered when the controller messages are updated
	ControllerMessageListener func(messages []*Record, context *Context)

	// MessageListener is triggered when the instance received new messages
	MessageListener func(instance string, messages []*Record, context *Context)
)

// NewHelixManager creates a new instance of HelixManager from a zookeeper connection string
func NewHelixManager(zkAddress string) *HelixManager {
	return &HelixManager{
		zkAddress: zkAddress,
	}
}

// NewSpectator creates a new Helix Spectator instance. This role handles most "read-only"
// operations of a Helix client.
func (m *HelixManager) NewSpectator(clusterID string) *Spectator {
	return &Spectator{
		ClusterID:                   clusterID,
		zkConnStr:                   m.zkAddress,
		externalViewListeners:       []ExternalViewChangeListener{},
		liveInstanceChangeListeners: []LiveInstanceChangeListener{},
		currentStateChangeListeners: map[string][]CurrentStateChangeListener{},
		idealStateChangeListeners:   []IdealStateChangeListener{},
		keys: KeyBuilder{clusterID},
		stop: make(chan bool),
		externalViewResourceMap:   map[string]bool{},
		idealStateResourceMap:     map[string]bool{},
		instanceConfigMap:         map[string]bool{},
		externalViewChanged:       make(chan string, 100),
		liveInstanceChanged:       make(chan string, 100),
		currentStateChanged:       make(chan string, 100),
		idealStateChanged:         make(chan string, 100),
		instanceConfigChanged:     make(chan string, 100),
		controllerMessagesChanged: make(chan string, 100),

		stopCurrentStateWatch: make(map[string]chan interface{}),

		currentStateChangeListenersLock: sync.Mutex{},
	}
}

// NewParticipant creates a new Helix Participant. This instance will act as a live instance
// of the Helix cluster when connected, and will participate the state model transition.
func (m *HelixManager) NewParticipant(clusterID string, host string, port string) *Participant {
	return &Participant{
		ClusterID:     clusterID,
		Host:          host,
		Port:          port,
		ParticipantID: fmt.Sprintf("%s_%s", host, port),
		zkConnStr:     m.zkAddress,
		started:       make(chan interface{}),
		stop:          make(chan bool),
		stopWatch:     make(chan bool),
		keys:          KeyBuilder{clusterID},
	}
}
