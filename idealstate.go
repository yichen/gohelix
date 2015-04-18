package gohelix

import (
	"strconv"
	"strings"
)

type IdealState struct {
	record Record
}

// public enum IdealStateProperty {
//   NUM_PARTITIONS,
//   STATE_MODEL_DEF_REF,
//   STATE_MODEL_FACTORY_NAME,
//   REPLICAS,
//   @Deprecated
//   IDEAL_STATE_MODE,
//   REBALANCE_MODE,
//   REBALANCE_TIMER_PERIOD,
//   MAX_PARTITIONS_PER_INSTANCE,
//   INSTANCE_GROUP_TAG,
//   REBALANCER_CLASS_NAME,
//   REBALANCER_CONFIG_NAME,
//   HELIX_ENABLED
// }

// var rebalanceMode = map[string]empty{
// 	"FULL_AUTO":    empty{},
// 	"SEMI_AUTO":    empty{},
// 	"CUSTOMIZED":   empty{},
// 	"USER_DEFINED": empty{},
// }

func NewIdealState(resource string) *IdealState {
	r := NewRecord(resource)
	is := IdealState{
		record: *r,
	}
	return &is
}

func (is *IdealState) SetNumPartitions(numPartitions int) {
	is.record.SetSimpleField("NUM_PARTITIONS", strconv.Itoa(numPartitions))
}

func (is *IdealState) SetStateModelDefRef(stateModel string) {
	is.record.SetSimpleField("STATE_MODEL_DEF_REF", stateModel)
}

func (is *IdealState) SetRebalanceMode(rebalance string) {
	is.record.SetSimpleField("REBALANCE_MODE", strings.ToUpper(rebalance))
}

func (is *IdealState) SetReplicas(replicas int) {
	is.record.SetSimpleField("REPLICAS", strconv.Itoa(replicas))
}

func (is *IdealState) Save(conn *connection, cluster string) {
	keys := KeyBuilder{cluster}
	path := keys.idealStates() + "/" + is.record.ID
	conn.CreateRecordWithPath(path, &is.record)
}
