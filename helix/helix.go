// provide command-line tools that is similar to helix-admin.sh
// see http://helix.apache.org/0.7.0-incubating-docs/Quickstart.html

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/yichen/gohelix"
)

func main() {
	app := cli.NewApp()
	app.Name = "helix"
	app.Usage = "helix"
	app.Author = "Yi Chen"
	app.Email = "yichen@outlook.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "zkSvr, z",
			Usage:  "zookeeper connection string",
			Value:  "localhost:2181",
			EnvVar: "ZOOKEEPER",
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "show debug output",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "addCluster",
			Usage: "add a cluster to helix",

			Action: func(c *cli.Context) {
				if err := mustArgc(c, 1); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().First()
				admin.AddCluster(cluster)
			},
		},
		{
			Name:  "dropCluster",
			Usage: "remove a cluster",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().First()
				err := admin.DropCluster(cluster)
				if err == nil {
					fmt.Println("Cluster '" + cluster + "' deleted.")
				} else {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "addNode",
			Usage: "add a node to cluster",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().First()
				node := c.Args().Get(1)
				// if node is the form of host:port, convert to host_port
				if strings.Contains(node, ":") {
					node = strings.Replace(node, ":", "_", 1)
				}

				if err := admin.AddNode(cluster, node); err != nil {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "dropNode",
			Usage: "drop a node from cluster",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().First()
				node := c.Args().Get(1)
				// if node is the form of host:port, convert to host_port
				if strings.Contains(node, ":") {
					node = strings.Replace(node, ":", "_", 1)
				}

				if err := admin.DropNode(cluster, node); err != nil {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "addResource",
			Usage: "add a resource to cluster",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 4); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				resource := c.Args().Get(1)
				partitions, err := strconv.Atoi(c.Args().Get(2))
				if err != nil {
					fmt.Println("Invalid parameter")
					return
				}
				stateModel := c.Args().Get(3)
				if err = admin.AddResource(cluster, resource, partitions, stateModel); err != nil {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "enableResource",
			Usage: "enable a resource",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				resource := c.Args().Get(1)

				if err := admin.EnableResource(cluster, resource); err != nil {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "disableResource",
			Usage: "disable a resource",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
				}

				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				resource := c.Args().Get(1)

				if err := admin.DisableResource(cluster, resource); err != nil {
					fmt.Println(err.Error())
				}
			},
		},
		{
			Name:  "listClusterInfo",
			Usage: "list existing cluster resources and instances",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 1); err != nil {
					fmt.Println(err.Error())
				}
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				info, err := admin.ListClusterInfo(cluster)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Println(info)
				}
			},
		},
		{
			Name:  "listClusters",
			Usage: "list existing cluster resources and instances",
			Action: func(c *cli.Context) {

				if err := mustArgc(c, 0); err != nil {
					fmt.Println(err.Error())
				}
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				clusters, err := admin.ListClusters()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(clusters)
			},
		},
		{
			Name:  "listResources",
			Usage: "list existing cluster resources",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 1); err != nil {
					fmt.Println(err.Error())
				}
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				resources, err := admin.ListResources(cluster)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(resources)
			},
		},
		{
			Name:  "listInstances",
			Usage: "list existing cluster instances",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 1); err != nil {
					fmt.Println(err.Error())
				}
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				instances, err := admin.ListInstances(cluster)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(instances)
			},
		},
		{
			Name:  "listInstanceInfo",
			Usage: "list detail info of a cluster instance",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 2); err != nil {
					fmt.Println(err.Error())
					return
				}
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				cluster := c.Args().Get(0)
				instance := c.Args().Get(1)

				info, err := admin.ListInstanceInfo(cluster, instance)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(info)
			},
		},
		{
			Name:  "setConfig",
			Usage: "configure the helix cluster",
			Action: func(c *cli.Context) {
				if err := mustArgc(c, 3); err != nil {
					fmt.Println(err.Error())
					return
				}

				scope := c.Args().Get(0)
				admin := gohelix.Admin{c.GlobalString("zkSvr")}
				if strings.ToLower(scope) != "cluster" {
					fmt.Println("Not supported")
					return
				}

				cluster := c.Args().Get(1)
				configTuple := strings.Split(c.Args().Get(2), "=")
				properties := map[string]string{}
				properties[strings.TrimSpace(configTuple[0])] = strings.TrimSpace(configTuple[1])
				if err := admin.SetConfig(cluster, scope, properties); err != nil {
					panic(err)
				}
			},
		},
		{
			Name:  "participant",
			Usage: "helix -z <zk> participant -c <cluster> -s <host> -p <port> -t <stateMode>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cluster, c",
					Usage: "cluster name",
				},
				cli.StringFlag{
					Name: "host, s",
				},
				cli.StringFlag{
					Name: "port, p",
				},
				cli.StringFlag{
					Name: "stateModelType, t",
				},
			},
			Action: func(c *cli.Context) {
				cluster := c.String("cluster")
				host := c.String("host")
				port := c.String("port")
				stateModel := c.String("stateModelType")

				startHelixParticipant(c.GlobalString("zkSvr"), cluster, host, port, stateModel)
			},
		},
		{
			Name:  "spectator",
			Usage: "helix -z <zk> spectator -c <cluster>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cluster, c",
					Usage: "cluster name",
				},
			},
			Action: func(c *cli.Context) {
				cluster := c.String("cluster")

				startHelixSpectator(c.GlobalString("zkSvr"), cluster)
			},
		},
	}

	app.Run(os.Args)
}

func mustArgc(c *cli.Context, n int) error {
	if len(c.Args()) != n {
		return fmt.Errorf("Wrong number of arguments")
	}
	return nil
}

// ./start-helix-participant.sh --zkSvr localhost:2199 --cluster MYCLUSTER --host localhost --port 12913 --stateModelType MasterSlave
// sample command:
// helix -z localhost:2181 participant  -c MYCLUSTER -s localhost -p 12913 -t MasterSlave
func startHelixParticipant(zk string, cluster string, host string, port string, stateModel string) {
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
}

// helix -z localhost:2181 spectator -c MYCLUSTER
func startHelixSpectator(zk string, cluster string) {

	evListener := func(ev []*gohelix.Record, context *gohelix.Context) {
		trigger := context.Get("trigger")
		if trigger != nil {
			fmt.Println("ExternalViewChangeListener: " + trigger.(string))
		}

		for _, record := range ev {
			r := *record
			fmt.Println(r)
		}
	}

	liListener := func(liveInstances []*gohelix.Record, context *gohelix.Context) {
		live := ""
		for _, i := range liveInstances {
			if len(live) == 0 {
				live += i.ID
			} else {
				live += ", " + i.ID
			}
		}

		fmt.Println("LiveInstanceChangeListener: " + live)
	}

	csListener := func(instance string, currentState []*gohelix.Record, context *gohelix.Context) {
		fmt.Println("CurrentStateChangeListener: " + instance)
	}

	manager := gohelix.NewHelixManager(zk)
	context := gohelix.NewContext()
	spectator := manager.NewSpectator(cluster)
	spectator.AddExternalViewChangeListener(evListener)
	spectator.AddLiveInstanceChangeListener(liListener)

	// TODO: hard-coded values
	spectator.AddCurrentStateChangeListener("localhost_12913", csListener)
	spectator.AddCurrentStateChangeListener("localhost_12914", csListener)
	spectator.AddCurrentStateChangeListener("localhost_12915", csListener)

	spectator.SetContext(context)
	spectator.Connect()
	defer spectator.Disconnect()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}
