package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/buraksezer/olric-testground-plans/sdks/network"
	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"github.com/testground/sdk-go/run"
	"github.com/testground/sdk-go/runtime"
)

func main() {
	run.Invoke(runf)
}


func runf(runenv *runtime.RunEnv, initCtx *run.InitContext) error {
	netcfg, err := network.New(runenv, initCtx)
	if err != nil {
		runenv.RecordCrash(err)
		return err
	}

	// Deployment scenario: embedded-member
	// This creates a single-node Olric cluster. It's good enough for experimenting.

	// config.New returns a new config.Config with sane defaults. Available values for env:
	// local, lan, wan
	c := config.New("lan")
	c.LogOutput = os.Stdout
	c.BindAddr = append(netcfg.IPv4.IP[:3:3], 1).String()
	c.MemberlistConfig.BindAddr = c.BindAddr

	// Callback function. It's called when this node is ready to accept connections.
	ctx, cancel := context.WithCancel(context.Background())
	c.Started = func() {
		defer cancel()

		runenv.RecordMessage("Olric is ready to accept connections!")
	}

	db, err := olric.New(c)
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("failed to create Olric instance: %v", err))
		return err
	}

	go func() {
		// Call Start at background. It's a blocker call.
		err = db.Start()
		if err != nil {
			runenv.RecordCrash(fmt.Sprintf("olric.Start returned an error: %v", err))
		}
	}()

	<-ctx.Done()

	dm, err := db.NewDMap("bucket-of-arbitrary-items")
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("olric.NewDMap returned an error: %v", err))
		return err
	}

	// Magic starts here!
	runenv.RecordMessage("##")
	runenv.RecordMessage("Operations on a DMap instance:")
	err = dm.Put("string-key", "buraksezer")
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("Failed to call Put: %v", err))
		return err
	}
	stringValue, err := dm.Get("string-key")
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("Failed to call Get: %v", err))
		return err
	}
	fmt.Printf("Value for string-key: %v, reflect.TypeOf: %s\n", stringValue, reflect.TypeOf(stringValue))

	err = dm.Put("uint64-key", uint64(1988))
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("Failed to call Put: %v", err))
		return err
	}
	uint64Value, err := dm.Get("uint64-key")
	if err != nil {
		runenv.RecordCrash(fmt.Sprintf("Failed to call Get: %v", err))
		return err
	}
	fmt.Printf("Value for uint64-key: %v, reflect.TypeOf: %s\n", uint64Value, reflect.TypeOf(uint64Value))
	runenv.RecordMessage("##")

	// Don't forget the call Shutdown when you want to leave the cluster.
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Shutdown(ctx)
}
