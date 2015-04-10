package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/yichen/gohelix"
)

var (
	lastLiveInstances map[string]gohelix.Record
	mutex             sync.Mutex
	manager           *gohelix.HelixManager
	tracer            *gohelix.Spectator
)

func trace(zk string, cluster string, verboseLevel int) {
	manager = gohelix.NewHelixManager(zk)
	tracer = manager.NewSpectator(cluster)

	context := gohelix.NewContext()
	context.Set("VerboseLevel", verboseLevel)
	tracer.SetContext(context)

	tracer.AddExternalViewChangeListener(externalViewChangeListener)
	tracer.AddLiveInstanceChangeListener(liveInstanceChangeListener)
	tracer.AddIdealStateChangeListener(idealStateChangeListener)

	tracer.Connect()
	defer tracer.Disconnect()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func getVerboseLevel(context *gohelix.Context) int {
	if vl := context.Get("VerboseLevel"); vl == nil {
		return 0
	} else {
		return vl.(int)
	}
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

	for k, _ := range before {
		if _, ok := after[k]; !ok {
			removed = append(removed, k)
		}
	}

	for k, _ := range after {
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
		log.Printf("[onExternalViewChange] number of resource groups: %d\n", len(ev))
	}
}

func idealStateChangeListener(is []*gohelix.Record, context *gohelix.Context) {

	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.Printf("[onIdealStateChange] number of resource groups: %d\n", len(is))
	}
}

func currentStateChangeListener(instance string, currentState []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	switch verboseLevel {
	case 0:
		log.Printf("[onStateChange] instance:%s\n", instance)
	}
}

func liveInstanceChangeListener(liveInstances []*gohelix.Record, context *gohelix.Context) {
	verboseLevel := getVerboseLevel(context)

	currentLiveInstances := getMapFromRecords(liveInstances)
	added, removed := diffRecords(lastLiveInstances, currentLiveInstances)

	// for newlly added instances, start watching them for CurrentStateChange
	for _, i := range added {
		tracer.AddCurrentStateChangeListener(i, currentStateChangeListener)
	}

	// save a copy of the current live instances map
	mutex.Lock()
	lastLiveInstances = currentLiveInstances
	mutex.Unlock()

	switch verboseLevel {
	case 0:
		log.Printf("[onLiveInstancesChange] number of live instances is %d. OFFLINE -> ONLINE: %d, ONLINE -> OFFLINE: %d\n", len(liveInstances), len(added), len(removed))
	}
}
