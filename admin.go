package gohelix

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrClusterNotSetup means the helix data structure in zookeeper /{CLUSTER_NAME}
	// is not correct or does not exist
	ErrClusterNotSetup = errors.New("cluster not setup")

	// ErrNodeAlreadyExists the zookeeper node exists when it is not expected to
	ErrNodeAlreadyExists = errors.New("node already exists in cluster")

	// ErrNodeNotExist the zookeeper node does not exist when it is expected to
	ErrNodeNotExist = errors.New("node does not exist in config for cluster")

	// ErrInstanceNotExist the instance of a cluster does not exist when it is expected to
	ErrInstanceNotExist = errors.New("node does not exist in instances for cluster")

	// ErrStateModelDefNotExist the state model definition is expected to exist in zookeeper
	ErrStateModelDefNotExist = errors.New("state model not exist in cluster")

	// ErrResourceExists the resource already exists in cluster and cannot be added again
	ErrResourceExists = errors.New("resource already exists in cluster")

	// ErrResourceNotExists the resource does not exists and cannot be removed
	ErrResourceNotExists = errors.New("resource not exists in cluster")
)

// Admin handles the administration task for the Helix cluster. Many of the operations
// are mirroring the implementions documented at
// http://helix.apache.org/0.7.0-incubating-docs/Quickstart.html
type Admin struct {
	ZkSvr string
}

// AddCluster add a cluster to Helix. As a result, a znode will be created in zookeeper
// root named after the cluster name, and corresponding data structures are populated
// under this znode.
func (adm Admin) AddCluster(cluster string) bool {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return false
	}
	defer conn.Disconnect()

	kb := KeyBuilder{cluster}
	// c = "/<cluster>"
	c := kb.cluster()

	// check if cluster already exists
	exists, err := conn.Exists(c)
	must(err)
	if exists {
		return false
	}

	conn.CreateEmptyNode(c)

	// PROPERTYSTORE is an empty node
	propertyStore := fmt.Sprintf("/%s/PROPERTYSTORE", cluster)
	conn.CreateEmptyNode(propertyStore)

	// STATEMODELDEFS has 6 children
	stateModelDefs := fmt.Sprintf("/%s/STATEMODELDEFS", cluster)
	conn.CreateEmptyNode(stateModelDefs)
	conn.CreateRecordWithData(stateModelDefs+"/LeaderStandby", HelixDefaultNodes["LeaderStandby"])
	conn.CreateRecordWithData(stateModelDefs+"/MasterSlave", HelixDefaultNodes["MasterSlave"])
	conn.CreateRecordWithData(stateModelDefs+"/OnlineOffline", HelixDefaultNodes["OnlineOffline"])
	conn.CreateRecordWithData(stateModelDefs+"/STORAGE_DEFAULT_SM_SCHEMATA", HelixDefaultNodes["STORAGE_DEFAULT_SM_SCHEMATA"])
	conn.CreateRecordWithData(stateModelDefs+"/SchedulerTaskQueue", HelixDefaultNodes["SchedulerTaskQueue"])
	conn.CreateRecordWithData(stateModelDefs+"/Task", HelixDefaultNodes["Task"])

	// INSTANCES is initailly an empty node
	instances := fmt.Sprintf("/%s/INSTANCES", cluster)
	conn.CreateEmptyNode(instances)

	// CONFIGS has 3 children: CLUSTER, RESOURCE, PARTICIPANT
	configs := fmt.Sprintf("/%s/CONFIGS", cluster)
	conn.CreateEmptyNode(configs)
	conn.CreateEmptyNode(configs + "/PARTICIPANT")
	conn.CreateEmptyNode(configs + "/RESOURCE")
	conn.CreateEmptyNode(configs + "/CLUSTER")

	clusterNode := NewRecord(cluster)
	conn.CreateRecordWithPath(configs+"/CLUSTER/"+cluster, clusterNode)

	// empty ideal states
	idealStates := fmt.Sprintf("/%s/IDEALSTATES", cluster)
	conn.CreateEmptyNode(idealStates)

	// empty external view
	externalView := fmt.Sprintf("/%s/EXTERNALVIEW", cluster)
	conn.CreateEmptyNode(externalView)

	// empty live instances
	liveInstances := fmt.Sprintf("/%s/LIVEINSTANCES", cluster)
	conn.CreateEmptyNode(liveInstances)

	// CONTROLLER has four childrens: [ERRORS, HISTORY, MESSAGES, STATUSUPDATES]
	controller := fmt.Sprintf("/%s/CONTROLLER", cluster)
	conn.CreateEmptyNode(controller)
	conn.CreateEmptyNode(controller + "/ERRORS")
	conn.CreateEmptyNode(controller + "/HISTORY")
	conn.CreateEmptyNode(controller + "/MESSAGES")
	conn.CreateEmptyNode(controller + "/STATUSUPDATES")

	return true
}

// SetConfig set the configuration values for the cluster, defined by the config scope
func (adm Admin) SetConfig(cluster string, scope string, properties map[string]string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	switch strings.ToUpper(scope) {
	case "CLUSTER":
		if allow, ok := properties["allowParticipantAutoJoin"]; ok {
			keys := KeyBuilder{cluster}
			path := keys.clusterConfig()

			if strings.ToLower(allow) == "true" {
				conn.UpdateSimpleField(path, "allowParticipantAutoJoin", "true")
			}
		}
	case "CONSTRAINT":
	case "PARTICIPANT":
	case "PARTITION":
	case "RESOURCE":
	}

	return nil
}

// GetConfig obtains the configuration value of a property, defined by a config scope
func (adm Admin) GetConfig(cluster string, scope string, keys []string) map[string]interface{} {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return nil
	}
	defer conn.Disconnect()

	result := make(map[string]interface{})

	switch scope {
	case "CLUSTER":
		kb := KeyBuilder{cluster}
		path := kb.clusterConfig()

		for _, k := range keys {
			result[k] = conn.GetSimpleFieldValueByKey(path, k)
		}
	case "CONSTRAINT":
	case "PARTICIPANT":
	case "PARTITION":
	case "RESOURCE":
	}

	return result
}

// DropCluster removes a helix cluster from zookeeper. This will remove the
// znode named after the cluster name from the zookeeper root.
func (adm Admin) DropCluster(cluster string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	kb := KeyBuilder{cluster}
	c := kb.cluster()

	return conn.DeleteTree(c)
}

// AddNode is the internal implementation corresponding to command
// ./helix-admin.sh --zkSvr <ZookeeperServerAddress> --addNode <clusterName instanceId>
// node is in the form of host_port
func (adm Admin) AddNode(cluster string, node string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	if ok, err := conn.IsClusterSetup(cluster); ok == false || err != nil {
		return ErrClusterNotSetup
	}

	// check if node already exists under /<cluster>/CONFIGS/PARTICIPANT/<NODE>
	keys := KeyBuilder{cluster}
	path := keys.participantConfig(node)
	exists, err := conn.Exists(path)
	must(err)
	if exists {
		return ErrNodeAlreadyExists
	}

	// create new node for the participant
	parts := strings.Split(node, "_")
	n := NewRecord(node)
	n.SetSimpleField("HELIX_HOST", parts[0])
	n.SetSimpleField("HELIX_PORT", parts[1])

	conn.CreateRecordWithPath(path, n)
	conn.CreateEmptyNode(keys.instance(node))
	conn.CreateEmptyNode(keys.messages(node))
	conn.CreateEmptyNode(keys.currentStates(node))
	conn.CreateEmptyNode(keys.errorsR(node))
	conn.CreateEmptyNode(keys.statusUpdates(node))

	return nil
}

// DropNode removes a node from a cluster. The corresponding znodes
// in zookeeper will be removed.
func (adm Admin) DropNode(cluster string, node string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	// check if node already exists under /<cluster>/CONFIGS/PARTICIPANT/<node>
	keys := KeyBuilder{cluster}
	if exists, err := conn.Exists(keys.participantConfig(node)); !exists || err != nil {
		return ErrNodeNotExist
	}

	// check if node exist under instance: /<cluster>/INSTANCES/<node>
	if exists, err := conn.Exists(keys.instance(node)); !exists || err != nil {
		return ErrInstanceNotExist
	}

	// delete /<cluster>/CONFIGS/PARTICIPANT/<node>
	conn.DeleteTree(keys.participantConfig(node))

	// delete /<cluster>/INSTANCES/<node>
	conn.DeleteTree(keys.instance(node))

	return nil
}

// AddResource implements the helix-admin.sh --addResource
// # helix-admin.sh --zkSvr <zk_address> --addResource <clustername> <resourceName> <numPartitions> <StateModelName>
// ./helix-admin.sh --zkSvr localhost:2199 --addResource MYCLUSTER myDB 6 MasterSlave
func (adm Admin) AddResource(cluster string, resource string, partitions int, stateModel string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}

	// make sure the state model def exists
	if exists, err := conn.Exists(keys.stateModel(stateModel)); !exists || err != nil {
		return ErrStateModelDefNotExist
	}

	// make sure the path for the ideal state does not exit
	isPath := keys.idealStates() + "/" + resource
	if exists, err := conn.Exists(isPath); exists || err != nil {
		if exists {
			return ErrResourceExists
		}
		return err
	}

	// create the idealstate for the resource
	is := NewIdealState(resource)
	is.SetNumPartitions(partitions)
	is.SetReplicas(0)
	is.SetRebalanceMode("SEMI_AUTO")
	is.SetStateModelDefRef(stateModel)
	// save the ideal state in zookeeper
	is.Save(conn, cluster)

	return nil
}

// DropResource removes the specified resource from the cluster.
func (adm Admin) DropResource(cluster string, resource string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}

	// make sure the path for the ideal state does not exit
	conn.DeleteTree(keys.idealStates() + "/" + resource)
	conn.DeleteTree(keys.resourceConfig(resource))

	return nil
}

// EnableResource enables the specified resource in the cluster
func (adm Admin) EnableResource(cluster string, resource string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}

	isPath := keys.idealStates() + "/" + resource

	if exists, err := conn.Exists(isPath); !exists || err != nil {
		if !exists {
			return ErrResourceNotExists
		}
		return err
	}

	// TODO: set the value at leaf node instead of the record level
	conn.UpdateSimpleField(isPath, "HELIX_ENABLED", "true")
	return nil
}

// DisableResource disables the specified resource in the cluster.
func (adm Admin) DisableResource(cluster string, resource string) error {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}

	isPath := keys.idealStates() + "/" + resource

	if exists, err := conn.Exists(isPath); !exists || err != nil {
		if !exists {
			return ErrResourceNotExists
		}

		return err
	}

	conn.UpdateSimpleField(isPath, "HELIX_ENABLED", "false")

	return nil
}

// Rebalance not implemented yet
func (adm Admin) Rebalance(cluster string, resource string, replicationFactor int) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		fmt.Println("Failed to connect to zookeeper.")
		return
	}
	defer conn.Disconnect()

	fmt.Println("Not implemented")
}

// ListClusterInfo shows the existing resources and instances in the glaster
func (adm Admin) ListClusterInfo(cluster string) (string, error) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return "", err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return "", ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}
	isPath := keys.idealStates()
	instancesPath := keys.instances()

	resources, err := conn.Children(isPath)
	if err != nil {
		return "", err
	}

	instances, err := conn.Children(instancesPath)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	buffer.WriteString("Existing resources in cluster " + cluster + ":\n")

	for _, r := range resources {
		buffer.WriteString("  " + r + "\n")
	}

	buffer.WriteString("\nInstances in cluster " + cluster + ":\n")
	for _, i := range instances {
		buffer.WriteString("  " + i + "\n")
	}
	return buffer.String(), nil
}

// ListClusters shows all Helix managed clusters in the connected zookeeper cluster
func (adm Admin) ListClusters() (string, error) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		fmt.Println("Failed to connect to zookeeper.")
		return "", err
	}
	defer conn.Disconnect()

	var clusters []string

	children, err := conn.Children("/")
	if err != nil {
		return "", err
	}

	for _, cluster := range children {
		if ok, err := conn.IsClusterSetup(cluster); ok && err == nil {
			clusters = append(clusters, cluster)
		}
	}

	var buffer bytes.Buffer
	buffer.WriteString("Existing clusters: \n")

	for _, cluster := range clusters {
		buffer.WriteString("  " + cluster + "\n")
	}
	return buffer.String(), nil
}

// ListResources shows a list of resources managed by the helix cluster
func (adm Admin) ListResources(cluster string) (string, error) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return "", err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return "", ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}
	isPath := keys.idealStates()
	resources, err := conn.Children(isPath)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	buffer.WriteString("Existing resources in cluster " + cluster + ":\n")

	for _, r := range resources {
		buffer.WriteString("  " + r + "\n")
	}

	return buffer.String(), nil
}

// ListInstances shows a list of instances participating the cluster.
func (adm Admin) ListInstances(cluster string) (string, error) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return "", err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return "", ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}
	isPath := keys.instances()
	instances, err := conn.Children(isPath)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Existing instances in cluster %s:\n", cluster))

	for _, r := range instances {
		buffer.WriteString("  " + r + "\n")
	}

	return buffer.String(), nil
}

// ListInstanceInfo shows detailed information of an inspace in the helix cluster
func (adm Admin) ListInstanceInfo(cluster string, instance string) (string, error) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		return "", err
	}
	defer conn.Disconnect()

	// make sure the cluster is already setup
	if ok, err := conn.IsClusterSetup(cluster); !ok || err != nil {
		return "", ErrClusterNotSetup
	}

	keys := KeyBuilder{cluster}
	instanceCfg := keys.participantConfig(instance)

	if exists, err := conn.Exists(instanceCfg); !exists || err != nil {
		if !exists {
			return "", ErrNodeNotExist
		}
		return "", err
	}

	r, err := conn.GetRecordFromPath(instanceCfg)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

// GetInstances prints out lists of instances
func (adm Admin) GetInstances(cluster string) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		fmt.Println("Failed to connect to zookeeper.")
	}
	defer conn.Disconnect()

	kb := KeyBuilder{cluster}
	instancesKey := kb.instances()

	data, err := conn.Get(instancesKey)
	must(err)

	for _, c := range data {
		fmt.Println(c)
	}

}

// DropInstance removes a participating instance from the helix cluster
func (adm Admin) DropInstance(zkSvr string, cluster string, instance string) {
	conn := newConnection(adm.ZkSvr)
	err := conn.Connect()
	if err != nil {
		fmt.Println("Failed to connect to zookeeper.")
	}
	defer conn.Disconnect()

	kb := KeyBuilder{cluster}
	instanceKey := kb.instance(instance)
	err = conn.Delete(instanceKey)
	must(err)

	fmt.Printf("/%s/%s deleted from zookeeper.\n", cluster, instance)
}
