package gohelix

import (
	"strings"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skip TestRunCommand")
	}

	commands := "/usr/bin/whoami"

	output, err := RunCommand(commands)
	if err != nil {
		t.Error(err.Error())
	}

	trimed := strings.TrimSpace(output)
	if trimed != "vagrant" {
		t.Error("wrong output: " + trimed + ", expected: vagrant")
	}
}

func TestCreateTestCluster(t *testing.T) {
	t.Parallel()

	now := time.Now().Local()
	cluster := "UtilTest_TestCreateTestCluster_" + now.Format("20060102150405")

	if testing.Short() {
		t.Skip("Skip TestCreateTestCluster")
	}

	if err := AddTestCluster(cluster); err != nil {
		t.Error(err.Error())
	}

	if err := DropTestCluster(cluster); err != nil {
		t.Error(err.Error())
	}
}
