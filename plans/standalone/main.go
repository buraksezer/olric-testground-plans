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
	"context"
	"fmt"
	"net"
	"time"

	"github.com/testground/sdk-go/network"
	"github.com/testground/sdk-go/run"
	"github.com/testground/sdk-go/runtime"
	"github.com/testground/sdk-go/sync"
)

var testcases = map[string]interface{}{
	"DMapGetPut":    DMapGetPut,
	"DMapPutDelete": DMapPutDelete,
	"DMapPutEx":     DMapPutEx,
	"DMapPutIf":     DMapPutIf,
}

func toKey(i int) string {
	return fmt.Sprintf("%09d", i)
}

func toVal(i int) []byte {
	return []byte(fmt.Sprintf("%010d", i))
}

func getIPAddress(r *runtime.RunEnv) (net.IP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	if !r.TestSidecar {
		return nil, fmt.Errorf("this plan must be run with sidecar enabled")
	}

	client := sync.MustBoundClient(ctx, r)
	netclient := network.NewClient(client, r)
	netclient.MustWaitNetworkInitialized(ctx)
	return netclient.MustGetDataNetworkIP(), nil
}

func main() {
	run.InvokeMap(testcases)
}
