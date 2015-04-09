package gohelix

import "fmt"

type HelixManager struct {
	zkAddress string

	conn *Connection
}

type (
	ExternalViewChangeListener   func(externalViews []*Record, context *Context)
	LiveInstanceChangeListener   func(liveInstances []*Record, context *Context)
	CurrentStateChangeListener   func(instance string, currentState []*Record, context *Context)
	IdealStateChangeListener     func(idealState []*Record, context *Context)
	InstanceConfigChangeListener func(configs []*Record, context *Context)
	MessageListener              func(instance string, messages []*Record, context *Context)
)

func NewHelixManager(zkAddress string) *HelixManager {
	return &HelixManager{
		zkAddress: zkAddress,
	}
}

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
		externalViewResourceMap: map[string]bool{},
		idealStateResourceMap:   map[string]bool{},
		externalViewChanged:     make(chan string, 100),
		liveInstanceChanged:     make(chan string, 100),
		currentStateChanged:     make(chan string, 100),
		idealStateChanged:       make(chan string, 100),
		instanceConfigChanged:   make(chan string, 100),
	}
}

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

func (m *HelixManager) AddLiveInstanceChangeListener(clusterID string, listener LiveInstanceChangeListener) {
	if m.conn == nil {
		m.conn = NewConnection(m.zkAddress)
	}

	kb := KeyBuilder{clusterID}

	kb.liveInstances()
}

func (m *HelixManager) AddCurrentStateChangeListener(clusterID string, instance string, sessionID string, listener CurrentStateChangeListener) {

}

func (m *HelixManager) AddIdealStateChangeListener(clusterID string, listener IdealStateChangeListener) {

}

func (m *HelixManager) AddMessageListener(instance string, listener MessageListener) {

}
