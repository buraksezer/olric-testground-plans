// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"github.com/testground/sdk-go/runtime"
)

type tgLogger struct {
	runenv *runtime.RunEnv
}

func (l *tgLogger) Write(b []byte) (int, error) {
	b = bytes.TrimSuffix(b, []byte("\n"))
	switch {
	case bytes.Contains(b, []byte("[ERROR]")):
		l.runenv.RecordFailure(fmt.Errorf("%s", string(b)))
	case bytes.Contains(b, []byte("[FATAL]")):
		l.runenv.RecordCrash(fmt.Errorf("%s", string(b)))
	default:
		l.runenv.RecordMessage(string(b))
	}
	return len(b), nil
}

func olricNode(runenv *runtime.RunEnv, f func(db *olric.Olric) error) error {
	ip, err := getIPAddress(runenv)
	if err != nil {
		return err
	}

	tl := &tgLogger{
		runenv: runenv,
	}

	// config.New returns a new config.Config with sane defaults. Available values for env:
	// local, lan, wan
	c := config.New("lan")
	c.LogOutput = tl
	c.BindAddr = ip.String()
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

	err = f(db)
	if err != nil {
		runenv.RecordFailure(err)
	}

	// Don't forget the call Shutdown when you want to leave the cluster.
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return db.Shutdown(ctx)

}
