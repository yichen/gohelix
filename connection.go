package gohelix

import (
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yichen/go-zookeeper/zk"
	"github.com/yichen/retry"
)

var (
	zkRetryOptions = retry.RetryOptions{
		"zookeeper",
		time.Millisecond * 10,
		time.Second * 1,
		1,
		0, // infinit retry
		false,
	}
)

type Connection struct {
	zkSvr       string
	zkConn      *zk.Conn
	isConnected bool
	stat        *zk.Stat
	sync.RWMutex
}

func NewConnection(zkSvr string) *Connection {
	conn := Connection{
		zkSvr: zkSvr,
	}

	return &conn
}

func (conn *Connection) Connect() error {
	zkServers := strings.Split(strings.TrimSpace(conn.zkSvr), ",")
	zkConn, _, err := zk.Connect(zkServers, 1*time.Minute)
	if err != nil {
		return err
	}

	_, _, err = zkConn.Exists("/zookeeper")
	if err != nil {
		return err
	}

	conn.isConnected = true
	conn.zkConn = zkConn

	return nil
}

func (conn *Connection) IsConnected() bool {
	if conn == nil || conn.isConnected == false {
		return false
	}

	_, _, err := conn.zkConn.Exists("/zookeeper")
	if err != nil {
		conn.isConnected = false
		return false
	}

	conn.isConnected = true
	return true
}

func (conn *Connection) GetSessionID() string {
	return strconv.FormatInt(conn.zkConn.SessionID, 10)
}

func (conn *Connection) Disconnect() {
	conn.zkConn.Close()
	conn.isConnected = false
}

func (conn *Connection) CreateEmptyNode(path string) {
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	_, err := conn.Create(path, []byte(""), flags, acl)
	must(err)
}

func (conn *Connection) CreateRecordWithData(path string, data string) {
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	_, err := conn.Create(path, []byte(data), flags, acl)
	must(err)
}

func (conn *Connection) CreateRecordWithPath(p string, r *Record) {
	parent := path.Dir(p)
	conn.ensurePath(parent)

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	data, err := r.Marshal()
	must(err)

	_, err = conn.Create(p, data, flags, acl)
	must(err)
}

func (conn *Connection) Exists(path string) (bool, error) {
	var result bool
	var stat *zk.Stat

	err := retry.RetryWithBackoff(zkRetryOptions, func() (retry.RetryStatus, error) {
		if r, s, err := conn.zkConn.Exists(path); err != nil {
			return retry.RetryContinue, nil
		} else {
			result = r
			stat = s
			return retry.RetryBreak, nil
		}
	})

	conn.stat = stat
	return result, err
}

func (conn *Connection) ExistsAll(paths ...string) (bool, error) {
	for _, path := range paths {
		if exists, err := conn.Exists(path); err != nil || exists == false {
			return exists, err
		}
	}

	return true, nil
}

func (conn *Connection) Get(path string) ([]byte, error) {
	var data []byte

	err := retry.RetryWithBackoff(zkRetryOptions, func() (retry.RetryStatus, error) {
		if d, s, err := conn.zkConn.Get(path); err != nil {
			return retry.RetryContinue, nil
		} else {
			data = d
			conn.stat = s
			return retry.RetryBreak, nil
		}
	})

	return data, err
}

func (conn *Connection) GetW(path string) ([]byte, <-chan zk.Event, error) {
	var data []byte
	var events <-chan zk.Event

	err := retry.RetryWithBackoff(zkRetryOptions, func() (retry.RetryStatus, error) {
		if d, s, evts, err := conn.zkConn.GetW(path); err != nil {
			return retry.RetryContinue, nil
		} else {
			data = d
			conn.stat = s
			events = evts
			return retry.RetryBreak, nil
		}
	})

	return data, events, err
}

func (conn *Connection) Set(path string, data []byte) error {
	_, err := conn.zkConn.Set(path, data, conn.stat.Version)
	return err
}

func (conn *Connection) Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	return conn.zkConn.Create(path, data, flags, acl)
}

func (conn *Connection) Children(path string) ([]string, error) {
	var children []string

	err := retry.RetryWithBackoff(zkRetryOptions, func() (retry.RetryStatus, error) {
		if c, s, err := conn.zkConn.Children(path); err != nil {
			return retry.RetryContinue, nil
		} else {
			children = c
			conn.stat = s
			return retry.RetryBreak, nil
		}
	})

	return children, err
}

func (conn *Connection) ChildrenW(path string) ([]string, <-chan zk.Event, error) {
	var children []string
	var eventChan <-chan zk.Event

	err := retry.RetryWithBackoff(zkRetryOptions, func() (retry.RetryStatus, error) {
		if c, s, evts, err := conn.zkConn.ChildrenW(path); err != nil {
			return retry.RetryContinue, nil
		} else {
			children = c
			conn.stat = s
			eventChan = evts
			return retry.RetryBreak, nil
		}
	})

	return children, eventChan, err
}

// update a map field for the znode. path is the znode path. key is the top-level key in
// the MapFields, mapProperty is the inner key, and value is the. For example:
//
// mapFields":{
// "eat1-app993.stg.linkedin.com_11932,BizProfile,p31_1,SLAVE":{
//   "CURRENT_STATE":"ONLINE"
//   ,"INFO":""
// }
// if we want to set the CURRENT_STATE to ONLINE, we call
// UpdateMapField("/RELAY/INSTANCES/{instance}/CURRENT_STATE/{sessionID}/{db}", "eat1-app993.stg.linkedin.com_11932,BizProfile,p31_1,SLAVE", "CURRENT_STATE", "ONLINE")
func (conn *Connection) UpdateMapField(path string, key string, property string, value string) error {
	data, err := conn.Get(path)
	if err != nil {
		return err
	}

	// convert the result into Record
	node, err := NewRecordFromBytes(data)
	if err != nil {
		return err
	}

	// update the value
	node.SetMapField(key, property, value)

	// mashall to bytes
	data, err = node.Marshal()
	if err != nil {
		return err
	}

	// copy back to zookeeper
	err = conn.Set(path, data)
	return err
}

func (conn *Connection) UpdateSimpleField(path string, key string, value string) {

	// get the current node
	data, err := conn.Get(path)
	must(err)

	// convert the result into Record
	node, err := NewRecordFromBytes(data)
	must(err)

	// update the value
	node.SetSimpleField(key, value)

	// mashall to bytes
	data, err = node.Marshal()
	must(err)

	// copy back to zookeeper
	err = conn.Set(path, data)
	must(err)
}

func (conn *Connection) GetSimpleFieldValueByKey(path string, key string) string {
	data, err := conn.Get(path)
	must(err)

	node, err := NewRecordFromBytes(data)
	must(err)

	if node.SimpleFields == nil {
		return ""
	}

	v := node.GetSimpleField(key)
	if v == nil {
		return ""
	} else {
		return v.(string)
	}
}

func (conn *Connection) GetSimpleFieldBool(path string, key string) bool {
	result := conn.GetSimpleFieldValueByKey(path, key)

	if strings.ToUpper(result) == "TRUE" {
		return true
	} else {
		return false
	}
}

func (conn *Connection) Delete(path string) error {
	return conn.zkConn.Delete(path, -1)
}

func (conn *Connection) DeleteTree(path string) error {
	if exists, err := conn.Exists(path); !exists || err != nil {
		return err
	}

	children, err := conn.Children(path)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		err := conn.zkConn.Delete(path, -1)
		return err
	}

	for _, c := range children {
		p := path + "/" + c
		e := conn.DeleteTree(p)
		if e != nil {
			return e
		}
	}

	return conn.Delete(path)
}

func (conn *Connection) RemoveMapFieldKey(path string, key string) error {
	data, err := conn.Get(path)
	if err != nil {
		return err
	}

	node, err := NewRecordFromBytes(data)
	if err != nil {
		return err
	}

	node.RemoveMapField(key)

	data, err = node.Marshal()
	if err != nil {
		return err
	}

	// save the data back to zookeeper
	err = conn.Set(path, data)
	return err
}

func (conn *Connection) IsClusterSetup(cluster string) (bool, error) {
	if conn.IsConnected() == false {
		if err := conn.Connect(); err != nil {
			return false, err
		}
	}

	keys := KeyBuilder{cluster}

	return conn.ExistsAll(
		keys.cluster(),
		keys.idealStates(),
		keys.participantConfigs(),
		keys.propertyStore(),
		keys.liveInstances(),
		keys.instances(),
		keys.externalView(),
		keys.stateModels(),
		keys.controller(),
		keys.controllerErrors(),
		keys.controllerHistory(),
		keys.controllerMessages(),
		keys.controllerStatusUpdates(),
	)
}

func (conn *Connection) GetRecordFromPath(path string) (*Record, error) {
	if data, err := conn.Get(path); err != nil {
		return nil, err
	} else {
		return NewRecordFromBytes(data)
	}
}

func (conn *Connection) SetRecordForPath(path string, r *Record) error {
	if exists, _ := conn.Exists(path); !exists {
		conn.ensurePath(path)
	}

	data, err := r.Marshal()
	if err != nil {
		return err
	}

	// need to get the stat.version before calling set
	conn.Lock()

	if _, err := conn.Get(path); err != nil {
		conn.Unlock()
		return err
	}

	if err := conn.Set(path, data); err != nil {
		conn.Unlock()
		return err
	}

	conn.Unlock()
	return nil

}

// EnsurePath makes sure the specified path exists.
// If not, create it
func (conn *Connection) ensurePath(p string) error {
	if exists, _ := conn.Exists(p); exists {
		return nil
	}

	parent := path.Dir(p)

	if exists, _ := conn.Exists(parent); !exists {
		conn.ensurePath(parent)
	}

	conn.CreateEmptyNode(p)
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
