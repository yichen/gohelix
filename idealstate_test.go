package gohelix

import "testing"

func TestIdeaState(t *testing.T) {
	t.Parallel()

	is := NewIdealState("resource")

	is.SetNumPartitions(32)
	value := is.record.GetSimpleField("NUM_PARTITIONS")
	if value == nil || value.(string) != "32" {
		t.Error("Failed to set and get NUM_PARTITIONS")
	}

	is.SetStateModelDefRef("MasterSlave")
	value = is.record.GetSimpleField("STATE_MODEL_DEF_REF")
	if value == nil || value.(string) != "MasterSlave" {
		t.Error("Failed to set/get STATE_MODEL_DEF_REF")
	}

	is.SetRebalanceMode("SEMI_AUTO")
	value = is.record.GetSimpleField("REBALANCE_MODE")
	if value == nil || value.(string) != "SEMI_AUTO" {
		t.Error("Failed to set/get REBALANCE_MODE")
	}

	is.SetReplicas(3)
	value = is.record.GetSimpleField("REPLICAS")
	if value == nil || value.(string) != "3" {
		t.Error("Failed to set/get REPLICAS")
	}
}
