package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/yichen/gohelix"

	log "github.com/Sirupsen/logrus"
)

var (
	lastLiveInstances map[string]gohelix.Record
	mutex             sync.Mutex
	manager           *gohelix.HelixManager
	tracer            *gohelix.Spectator
)

func init() {
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func trace(zk string, cluster string, verboseLevel int) {
	manager = gohelix.NewHelixManager(zk)
	tracer = manager.NewSpectator(cluster)

	context := gohelix.NewContext()
	context.Set("VerboseLevel", verboseLevel)
	tracer.SetContext(context)

	tracer.AddExternalViewChangeListener(externalViewChangeListener)
	tracer.AddLiveInstanceChangeListener(liveInstanceChangeListener)
	tracer.AddIdealStateChangeListener(idealStateChangeListener)
	tracer.AddControllerMessageListener(controllerMessagesListener)
	tracer.AddInstanceConfigChangeListener(instanceConfigChangeListener)

	if err := tracer.Connect(); err != nil {
		fmt.Println("Unable to connect to zookeeper: " + err.Error())
		return
	}
	defer tracer.Disconnect()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func getVerboseLevel(context *gohelix.Context) int {
	vl := context.Get("VerboseLevel")
	if vl == nil {
		return 0
	}
	return vl.(int)
}

func getMapFromRecords(records []*gohelix.Record) map[string]gohelix.Record {
	result := map[string]gohelix.Record{}

	for _, r := range records {
		result[r.ID] = *r
	}

	return result
}

func diffRecords(before map[string]gohelix.Record, after map[string]gohelix.Record) ([]string, []string) {
	added := []string{}
	removed := []string{}

	for k := range before {
		if _, ok := after[k]; !ok {
			removed = append(removed, k)
		}
	}

	for k := range after {
		if _, ok := before[k]; !ok {
			added = append(added, k)
		}
	}

	return added, removed
}

func externalViewChangeListener(ev []*gohelix.Record, context *gohelix.Context) {

	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onExternalViewChange").Infof("number of resource groups: %d", len(ev))
	}
}

func idealStateChangeListener(is []*gohelix.Record, context *gohelix.Context) {

	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onIdealStateChange").Infof("number of resource groups: %d", len(is))
	}
}

func currentStateChangeListener(instance string, currentState []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onStateChange").Infof("instance:%s", instance)
	}
}

func liveInstanceChangeListener(liveInstances []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	currentLiveInstances := getMapFromRecords(liveInstances)
	added, removed := diffRecords(lastLiveInstances, currentLiveInstances)

	// for newlly added instances, start watching them for CurrentStateChange
	for _, i := range added {
		log.Printf("Add CurrentStateChangedListener for live instance: %s", i)
		tracer.AddCurrentStateChangeListener(i, currentStateChangeListener)
		log.Printf("Add MessageListener for live instance: %s", i)
		tracer.AddMessageListener(i, instanceMessageListener)
	}

	// save a copy of the current live instances map
	mutex.Lock()
	lastLiveInstances = currentLiveInstances
	mutex.Unlock()

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onLiveInstancesChange").Infof("number of live instances is %d. OFFLINE -> ONLINE: %d, ONLINE -> OFFLINE: %d", len(liveInstances), len(added), len(removed))
	}
}

func controllerMessagesListener(messages []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onMessage").Infof("Number of controller messages is %d", len(messages))
	}
}

func instanceMessageListener(instance string, messages []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onMessage").Infof("Instance %s received new messages.", instance)
		break
	case 1:
		msgStr := ""
		for _, m := range messages {
			msgStr += m.String()
			msgStr += "\n"
		}

		log.WithField("CALLBACK", "onMessage").Infof("Instance %s received messages: %s", instance, msgStr)
	}
}

func instanceConfigChangeListener(configs []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.WithField("CALLBACK", "onInstanceConfigChange").Infof("Number of instances configs is %d", len(configs))
	}
}
