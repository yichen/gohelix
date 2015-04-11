package main

import (
	"fmt"

	"github.com/yichen/gohelix"
)

func main() {
	cluster := "MYCLUSTER"
	resource := "myDB"
	replica := "6"

	fmt.Println("Stop controller")
	if err := gohelix.StopController(); err != nil {
		fmt.Println("Failed to stop controller: " + err.Error())
	}

	fmt.Println("Drop MYCLUSTER")
	if err := gohelix.DropTestCluster(cluster); err != nil {
		fmt.Println("Failed to drop cluster: " + err.Error())
		return
	}

	fmt.Println("Create MYCLUSTER")
	if err := gohelix.AddTestCluster(cluster); err != nil {
		fmt.Println("Failed to create MYCLUSTER: " + err.Error())
		return
	}

	fmt.Println("Add node 12913")
	if err := gohelix.AddNode(cluster, "localhost", "12913"); err != nil {
		fmt.Println("Failed to add node 12913: " + err.Error())
		return
	}

	fmt.Println("Add node 12914")
	if err := gohelix.AddNode(cluster, "localhost", "12914"); err != nil {
		fmt.Println("Failed to add node 12914: " + err.Error())
		return
	}

	fmt.Println("Add node 12915")
	if err := gohelix.AddNode(cluster, "localhost", "12915"); err != nil {
		fmt.Println("Failed to add node 12915: " + err.Error())
		return
	}

	fmt.Println("Add resource myDB")
	if err := gohelix.AddResource(cluster, resource, replica); err != nil {
		fmt.Println("Failed to add resource: " + err.Error())
		return
	}

	fmt.Println("Rebalance")
	if err := gohelix.Rebalance(cluster, resource, "3"); err != nil {
		fmt.Println("Failed to rebalance: " + err.Error())
		return
	}

	fmt.Println("Start controller")
	if err := gohelix.StartController(); err != nil {
		fmt.Println("Failed to start controller: " + err.Error())
		return
	}

	fmt.Println("SUCCESS!")
}
