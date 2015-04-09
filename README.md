goHelix: Golang binding for Apache Helix
-----

# Start a Helix spectator

```go

    evListener := func(ev []*gohelix.Record, context *gohelix.Context) {
        fmt.Println("ExternalViewChangeListener")
    }

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

    manager := gohelix.NewHelixManager(zk)
    context := gohelix.NewContext()
    spectator := manager.NewSpectator(cluster)
    spectator.AddExternalViewChangeListener(evListener)
    spectator.AddLiveInstanceChangeListener(liListener)

    // passing data into the listener
    spectator.SetContext(context)

    // Connect and start the spectator
    spectator.Connect()
    defer spectator.Disconnect()
```


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







