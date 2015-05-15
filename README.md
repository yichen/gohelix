# gohelix: Golang binding for Apache Helix

Gohelix is an experimental helix client for Golang. It currently support functionalities of Helix spectator role, and partial support for Helix participant role. It also provides a command line utility that mirrows the functions of `helix-admin.sh`.

## Quick Start

The following steps demonstrate the what `gohelix` can do (and cannot do) how to setup a Helix cluster using `gohelix`. Since `gohelix` is a client-side library, it still has dependency on the zookeeper and a working helix controller. 

### Prerequisit 
More information about zookeeper can be found here: 

* [http://zookeeper.apache.org/doc/r3.3.3/zookeeperStarted.html](http://zookeeper.apache.org/doc/r3.3.3/zookeeperStarted.html)

Instructions on setting up a Helix controller can be found here:

* [http://helix.apache.org/0.7.0-incubating-docs/Quickstart.html](http://helix.apache.org/0.7.0-incubating-docs/Quickstart.html)

### Download and install helix cli

```shell
go get github.com/yichen/gohelix/helix
```

This will install the command line utility `helix`. After installation, run `helix` from command line with no arguments you should see a list of subcommands. The commands follow this format:

```
helix -z [zookeeper address] {subcommand} arguments...
```

That is, the first argument is the zookeeper address. If this is not specified, it is default to `localhost:2181`.

### Use helix cli

* To create a cluster from zookeeper:

```
helix -z localhost:2181 addCluster MYCLUSTER
```

* Add three nodes to be managed by helix for the cluster MYCLUSTER. The three nodes are `localhost:12913`, `localhost:12914`, `localhost:12915`

```
helix -z localhost:2181 addNode MYCLUSTER localhost:12913
helix -z localhost:2181 addNode MYCLUSTER localhost:12914
helix -z localhost:2181 addNode MYCLUSTER localhost:12915
```

* Add a database `myDB` as a resource to be managed by helix cluster

Now, we want the helix cluster to manage the resource `myDB`. Define the resource to have 8 partitions, and we will use a state model `MasterSlave`

```
helix -z localhost:2181 addResource MYCLUSTER myDB 8 MasterSlave
```

* To inspect the cluster

To list all clusters managed by helix:

```
helix -z localhost:2181 listClusters
```

To show details of a cluster

```
helix -z localhost:2181 listClusterInfo MYCLUSTER
```

* To remove a cluster from helix:

```
helix -z localhost:2181 dropCluster MYCLUSTER
```


## Helix Spectator

The helix cli tool is playing the Helix Spectator role. The following example shows how to make the most of the spectator role by listening to the cluster state changes. 

```go

    // define the external view change listener
    evListener := func(ev []*gohelix.Record, context *gohelix.Context) {
        fmt.Println("ExternalViewChangeListener")
    }

    // define the live instance change listener
    liListener := func(liveInstances []*gohelix.Record, context *gohelix.Context) {
        liveInstanceList := ""
        for _, i := range liveInstances {
            if len(liveInstanceList) == 0 {
                liveInstanceList += i.ID
            } else {
                liveInstanceList += ", " + i.ID
            }
        }

        fmt.Println("LiveInstanceChangeListener: " + liveInstanceList)
    }

    // create an instance of helix manager with zookeeper connection string
    manager := gohelix.NewHelixManager("localhost:2181")
    spectator := manager.NewSpectator("MYCLUSTER")
    spectator.AddExternalViewChangeListener(evListener)
    spectator.AddLiveInstanceChangeListener(liListener)

    // context is a way to pass data into the listeners
    context := gohelix.NewContext()
    spectator.SetContext(context)

    // Connect and start the spectator. This is when the client connects to
    // the zookeeper and also attach the listeners to the event loop
    spectator.Connect()
    defer spectator.Disconnect()

    // with Spectator instance connected, we can retrieve helix data from
    // zookeeper. For example, get the current external view of the cluster
    externalView := spectator.GetExternalView()

    // Get the current ideal state
    idealState := spectator.GetIdealState()

    // Get the participant configs
    configs := spectator.GetInstanceConfigs()

    // Get the current live instances
    liveInstances := spectator.GetLiveInstances()

```


# Helix Participant

```go
    manager := gohelix.NewHelixManager(zk)
    participant := manager.NewParticipant(cluster, host, port)

    // creaet OnlineOffline state model
    sm := gohelix.NewStateModel([]gohelix.Transition{
        {"ONLINE", "OFFLINE", func(partition string) {
            fmt.Println("ONLINE-->OFFLINE")
        }},
        {"OFFLINE", "ONLINE", func(partition string) {
            fmt.Println("OFFLINE-->ONLINE")
        }},
    })

    participant.RegisterStateModel(stateModel, sm)

    err := participant.Connect()
    if err != nil {
        fmt.Println(err.Error())
        return
    }

    // block until SIGINT and SIGTERM
    c := make(chan os.Signal, 2)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
```


# Command-line tool `helix`

To install the command-line tool:

```
go install github.com/yichen/gohelix/helix
```

To get the help:

```
helix -h
```

Most of the commands are mirroring the commands provided by Apache Helix `helix-admin.sh`



