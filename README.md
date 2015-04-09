goHelix: Golang binding for Apache Helix
-----

This is an experimental helix client for Golang. It is under active development and not yet ready for production.

# Start a Helix spectator

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


The `Spectator` instance also provide 

# Start a participant

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


# Command-line too `helix`

To install the command-line tool:

```
go install github.com/yichen/gohelix/helix
```

To get the help:

```
helix -h
```

Most of the commands are mirroring the commands provided by Apache Helix `helix-admin.sh`



