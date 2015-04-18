package gohelix

import (
	"fmt"
	"sync"
	"time"
)

type spectatorState uint8

const (
	spectatorConnected    spectatorState = 0
	spectatorDisConnected spectatorState = 1
)

// Spectator is a Helix role that does not participate the cluster state transition
// but only read cluster data, or listen to cluster updates
type Spectator struct {
	// HelixManager
	conn *connection

	// The cluster this spectator is specatating
	ClusterID string

	// zookeeper connection string
	zkConnStr string

	// external view change handler
	externalViewListeners         []ExternalViewChangeListener
	liveInstanceChangeListeners   []LiveInstanceChangeListener
	currentStateChangeListeners   map[string][]CurrentStateChangeListener
	idealStateChangeListeners     []IdealStateChangeListener
	instanceConfigChangeListeners []InstanceConfigChangeListener
	controllerMessageListeners    []ControllerMessageListener
	messageListener               MessageListener

	// locks for modifying the internal data structures
	currentStateChangeListenersLock sync.Mutex

	// stop the spectator
	stop chan bool

	// keybuilder
	keys KeyBuilder

	// resources the external view is tracking. It is a map from the resource name to the
	// current state of the resource: true means it is active, false means the resource is inactive/deleted
	externalViewResourceMap map[string]bool
	idealStateResourceMap   map[string]bool
	instanceConfigMap       map[string]bool

	// channel to notify that external view has changes
	// this include new resource added, resource removed, and resource changed. Whenever there is
	// a change, we will retrieve the current snapshot of the external view and invoke the listener
	// the value in the channel is the current externalview node that has been updated. If the value
	// is empty string "", it means the root is changed.
	externalViewChanged       chan string
	liveInstanceChanged       chan string
	currentStateChanged       chan string
	idealStateChanged         chan string
	instanceConfigChanged     chan string
	controllerMessagesChanged chan string

	// control channels for stopping watches
	stopCurrentStateWatch map[string]chan interface{}

	// context of the specator, accessible from the ExternalViewChangeListener
	context *Context

	state spectatorState
}

// Connect the spectator. When connected, the spectator is able to listen to Helix cluster
// changes and handle listener updates.
func (s *Spectator) Connect() error {
	if s.conn != nil && s.conn.IsConnected() {
		return nil
	}

	s.conn = newConnection(s.zkConnStr)
	if err := s.conn.Connect(); err != nil {
		return err
	}

	if ok, err := s.conn.IsClusterSetup(s.ClusterID); !ok || err != nil {
		return ErrClusterNotSetup
	}

	// start the event loop for spectator
	s.loop()

	s.state = spectatorConnected
	return nil
}

// Disconnect will disconnect the spectator from zookeeper, and also stop all listeners
func (s *Spectator) Disconnect() {
	if s.state == spectatorDisConnected {
		return
	}

	// wait for graceful shutdown of the external view listener
	if s.state != spectatorDisConnected {
		s.stop <- true
		close(s.stop)
	}

	for s.state != spectatorDisConnected {
		time.Sleep(100 * time.Millisecond)
	}

	s.state = spectatorDisConnected
}

// IsConnected test if the spectator is connected
func (s *Spectator) IsConnected() bool {
	return s.state == spectatorConnected
}

// SetContext set the context that can be used within the listeners
func (s *Spectator) SetContext(context *Context) {
	s.context = context
}

// AddExternalViewChangeListener add a listener to external view changes.
func (s *Spectator) AddExternalViewChangeListener(listener ExternalViewChangeListener) {
	s.externalViewListeners = append(s.externalViewListeners, listener)
}

// AddLiveInstanceChangeListener add a listener to live instance changes.
func (s *Spectator) AddLiveInstanceChangeListener(listener LiveInstanceChangeListener) {
	s.liveInstanceChangeListeners = append(s.liveInstanceChangeListeners, listener)
}

// AddCurrentStateChangeListener add a listener to current state changes of the specified instance.
func (s *Spectator) AddCurrentStateChangeListener(instance string, listener CurrentStateChangeListener) {
	s.currentStateChangeListenersLock.Lock()
	if s.currentStateChangeListeners[instance] == nil {
		s.currentStateChangeListeners[instance] = []CurrentStateChangeListener{}
	}

	s.currentStateChangeListeners[instance] = append(s.currentStateChangeListeners[instance], listener)
	s.currentStateChangeListenersLock.Unlock()

	// if we are adding new listeners when the specator is already connected, we need
	// to kick of the listener in the event loop
	if len(s.currentStateChangeListeners[instance]) == 1 && s.IsConnected() {
		s.watchCurrentStateForInstance(instance)
	}
}

// AddIdealStateChangeListener add a listener to the cluster ideal state changes
func (s *Spectator) AddIdealStateChangeListener(listener IdealStateChangeListener) {
	s.idealStateChangeListeners = append(s.idealStateChangeListeners, listener)
}

// AddInstanceConfigChangeListener add a listener to instance config changes
func (s *Spectator) AddInstanceConfigChangeListener(listener InstanceConfigChangeListener) {
	s.instanceConfigChangeListeners = append(s.instanceConfigChangeListeners, listener)
}

// AddControllerMessageListener add a listener to controller messages
func (s *Spectator) AddControllerMessageListener(listener ControllerMessageListener) {
	s.controllerMessageListeners = append(s.controllerMessageListeners, listener)
}

func (s *Spectator) watchExternalViewResource(resource string) {
	go func() {
		for {
			// block and wait for the next update for the resource
			// when the update happens, unblock, and also send the resource
			// to the channel
			_, events, err := s.conn.GetW(s.keys.externalViewForResource(resource))
			<-events
			s.externalViewChanged <- resource
			must(err)
		}
	}()
}

func (s *Spectator) watchIdealStateResource(resource string) {
	go func() {
		for {
			// block and wait for the next update for the resource
			// when the update happens, unblock, and also send the resource
			// to the channel
			_, events, err := s.conn.GetW(s.keys.idealStateForResource(resource))
			<-events
			s.idealStateChanged <- resource
			must(err)
		}
	}()
}

// GetControllerMessages retrieves controller messages from zookeeper
func (s *Spectator) GetControllerMessages() []*Record {
	result := []*Record{}
	messages, err := s.conn.Children(s.keys.controllerMessages())

	if err != nil {
		return result
	}

	for _, m := range messages {
		record, err := s.conn.GetRecordFromPath(s.keys.controllerMessage(m))
		if err != nil {
			result = append(result, record)
		}
	}

	return result
}

// GetLiveInstances retrieve a copy of the current live instances.
func (s *Spectator) GetLiveInstances() []*Record {
	liveInstances := []*Record{}
	instances, err := s.conn.Children(s.keys.liveInstances())
	if err != nil {
		fmt.Println("Error in GetLiveInstances: " + err.Error())
		return nil
	}

	for _, participantID := range instances {
		r, err := s.conn.GetRecordFromPath(s.keys.liveInstance(participantID))
		if err != nil {
			fmt.Println("Error in get live instance for " + participantID)
			continue
		}

		liveInstances = append(liveInstances, r)
	}

	return liveInstances
}

// GetExternalView retrieves a copy of the external views
func (s *Spectator) GetExternalView() []*Record {
	result := []*Record{}

	for k, v := range s.externalViewResourceMap {
		if v == false {
			continue
		}

		record, err := s.conn.GetRecordFromPath(s.keys.externalViewForResource(k))

		if err == nil {
			result = append(result, record)
			continue
		}
	}

	return result
}

// GetIdealState retrieves a copy of the ideal state
func (s *Spectator) GetIdealState() []*Record {
	result := []*Record{}

	for k, v := range s.idealStateResourceMap {
		if v == false {
			continue
		}

		record, err := s.conn.GetRecordFromPath(s.keys.idealStateForResource(k))

		if err == nil {
			result = append(result, record)
			continue
		}
	}
	return result
}

// GetCurrentState retrieves a copy of the current state for specified instance
func (s *Spectator) GetCurrentState(instance string) []*Record {
	result := []*Record{}

	resources, err := s.conn.Children(s.keys.instance(instance))
	must(err)

	for _, r := range resources {
		record, err := s.conn.GetRecordFromPath(s.keys.currentStateForResource(instance, s.conn.GetSessionID(), r))
		if err == nil {
			result = append(result, record)
		}
	}

	return result
}

// GetInstanceConfigs retrieves a copy of instance configs from zookeeper
func (s *Spectator) GetInstanceConfigs() []*Record {
	result := []*Record{}

	configs, err := s.conn.Children(s.keys.participantConfigs())
	must(err)

	for _, i := range configs {
		record, err := s.conn.GetRecordFromPath(s.keys.participantConfig(i))
		if err == nil {
			result = append(result, record)
		}
	}

	return result
}

func (s *Spectator) watchCurrentStates() {
	for k := range s.currentStateChangeListeners {
		s.watchCurrentStateForInstance(k)
	}
}

func (s *Spectator) watchCurrentStateForInstance(instance string) {
	sessions, err := s.conn.Children(s.keys.currentStates(instance))
	must(err)

	// TODO: only have one session?
	if len(sessions) > 0 {
		resources, err := s.conn.Children(s.keys.currentStatesForSession(instance, sessions[0]))
		must(err)

		for _, r := range resources {
			s.watchCurrentStateOfInstanceForResource(instance, r, sessions[0])
		}
	}
}

func (s *Spectator) watchCurrentStateOfInstanceForResource(instance string, resource string, sessionID string) {

	watchPath := s.keys.currentStateForResource(instance, sessionID, resource)
	if _, ok := s.stopCurrentStateWatch[watchPath]; !ok {
		s.stopCurrentStateWatch[watchPath] = make(chan interface{})
	}

	// check if the session are ever expired. If so, remove the watcher
	go func() {

		c := time.Tick(10 * time.Second)
		for now := range c {
			if ok, err := s.conn.Exists(watchPath); !ok || err != nil {
				s.stopCurrentStateWatch[watchPath] <- now
				return
			}
		}
	}()

	go func() {
		for {
			_, events, err := s.conn.GetW(watchPath)
			must(err)
			select {
			case <-events:
				s.currentStateChanged <- instance
				continue
			case <-s.stopCurrentStateWatch[watchPath]:
				delete(s.stopCurrentStateWatch, watchPath)
				return
			}
		}
	}()
}

func (s *Spectator) watchLiveInstances() {
	errors := make(chan error)

	go func() {
		for {
			_, events, err := s.conn.ChildrenW(s.keys.liveInstances())
			if err != nil {
				errors <- err
				return
			}

			// notify the live instance update
			s.liveInstanceChanged <- ""

			// block the loop to wait for the live instance change
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
}

func (s *Spectator) watchInstanceConfig() {
	errors := make(chan error)

	go func() {
		for {
			configs, events, err := s.conn.ChildrenW(s.keys.participantConfigs())
			if err != nil {
				errors <- err
				return
			}

			// find the resources that are newly added, and create a watcher
			for _, k := range configs {
				_, ok := s.instanceConfigMap[k]
				if !ok {
					s.watchInstanceConfigForParticipant(k)
					s.instanceConfigMap[k] = true
				}
			}

			// refresh the instanceConfigMap to make sure only the currently existing resources
			// are marked as true
			for k := range s.instanceConfigMap {
				s.instanceConfigMap[k] = false
			}
			for _, k := range configs {
				s.instanceConfigMap[k] = true
			}

			// Notify an update of external view if there are new resources added.
			s.instanceConfigChanged <- ""

			// now need to block the loop to wait for the next update event
			evt := <-events
			if evt.Err != nil {
				panic(evt.Err)
				return
			}
		}
	}()
}

func (s *Spectator) watchInstanceConfigForParticipant(instance string) {
	go func() {
		for {
			// block and wait for the next update for the resource
			// when the update happens, unblock, and also send the resource
			// to the channel
			_, events, err := s.conn.GetW(s.keys.participantConfig(instance))
			<-events
			s.instanceConfigChanged <- instance
			must(err)
		}
	}()

}

func (s *Spectator) watchIdealState() {
	errors := make(chan error)

	go func() {
		for {
			resources, events, err := s.conn.ChildrenW(s.keys.idealStates())
			if err != nil {
				errors <- err
				return
			}

			// find the resources that are newly added, and create a watcher
			for _, k := range resources {
				_, ok := s.idealStateResourceMap[k]
				if !ok {
					s.watchIdealStateResource(k)
					s.idealStateResourceMap[k] = true
				}
			}

			// refresh the idealStateResourceMap to make sure only the currently existing resources
			// are marked as true
			for k := range s.idealStateResourceMap {
				s.idealStateResourceMap[k] = false
			}
			for _, k := range resources {
				s.idealStateResourceMap[k] = true
			}

			// Notify an update of external view if there are new resources added.
			s.idealStateChanged <- ""

			// now need to block the loop to wait for the next update event
			evt := <-events
			if evt.Err != nil {
				panic(evt.Err)
				return
			}
		}
	}()
}

func (s *Spectator) watchExternalView() {
	errors := make(chan error)

	go func() {
		for {
			resources, events, err := s.conn.ChildrenW(s.keys.externalView())
			if err != nil {
				errors <- err
				return
			}

			// find the resources that are newly added, and create a watcher
			for _, k := range resources {
				_, ok := s.externalViewResourceMap[k]
				if !ok {
					s.watchExternalViewResource(k)
					s.externalViewResourceMap[k] = true
				}
			}

			// refresh the externalViewResourceMap to make sure only the currently existing resources
			// are marked as true
			for k := range s.externalViewResourceMap {
				s.externalViewResourceMap[k] = false
			}
			for _, k := range resources {
				s.externalViewResourceMap[k] = true
			}

			// Notify an update of external view if there are new resources added.
			s.externalViewChanged <- ""

			// now need to block the loop to wait for the next update event
			evt := <-events
			if evt.Err != nil {
				panic(evt.Err)
				return
			}
		}
	}()
}

// watchControllerMessages only watch the changes of message list, it currently
// doesn't watch the content of the messages.
func (s *Spectator) watchControllerMessages() {
	go func() {
		_, events, err := s.conn.ChildrenW(s.keys.controllerMessages())
		if err != nil {
			panic(err)
		}

		// send the INIT update
		s.controllerMessagesChanged <- ""

		// block to wait for CALLBACK
		<-events
	}()
}

// loop is the main event loop for Spectator. Whenever an external view update happpened
// the loop will pause for a short period of time to bucket all subsequent external view
// changes so that we don't send duplicate updates too often.
func (s *Spectator) loop() {

	hasListeners := false

	if len(s.externalViewListeners) > 0 {
		hasListeners = true
		s.watchExternalView()
	}

	if len(s.liveInstanceChangeListeners) > 0 {
		hasListeners = true
		s.watchLiveInstances()
	}

	if len(s.currentStateChangeListeners) > 0 {
		hasListeners = true
		s.watchCurrentStates()
	}

	if len(s.idealStateChangeListeners) > 0 {
		hasListeners = true
		s.watchIdealState()
	}

	if len(s.controllerMessageListeners) > 0 {
		hasListeners = true
		s.watchControllerMessages()
	}

	if len(s.instanceConfigChangeListeners) > 0 {
		hasListeners = true
		s.watchInstanceConfig()
	}

	if !hasListeners {
		return
	}

	go func() {
		for {
			select {
			case <-s.stop:
				s.state = spectatorDisConnected
				return

			case <-s.liveInstanceChanged:
				li := s.GetLiveInstances()
				for _, l := range s.liveInstanceChangeListeners {
					go l(li, s.context)
				}
				continue

			case r := <-s.externalViewChanged:
				ev := s.GetExternalView()
				if s.context != nil {
					s.context.Set("trigger", r)
				}

				for _, evListener := range s.externalViewListeners {
					go evListener(ev, s.context)
				}
				continue

			case p := <-s.currentStateChanged:
				cs := s.GetCurrentState(p)
				for _, listener := range s.currentStateChangeListeners[p] {
					go listener(p, cs, s.context)
				}
				continue

			case <-s.idealStateChanged:
				is := s.GetIdealState()

				for _, isListener := range s.idealStateChangeListeners {
					go isListener(is, s.context)
				}
				continue

			case <-s.instanceConfigChanged:
				ic := s.GetInstanceConfigs()
				for _, icListener := range s.instanceConfigChangeListeners {
					go icListener(ic, s.context)
				}
				continue

			case <-s.controllerMessagesChanged:
				cm := s.GetControllerMessages()
				for _, cmListener := range s.controllerMessageListeners {
					go cmListener(cm, s.context)
				}
			}
		}
	}()
}
