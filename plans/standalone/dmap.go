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
	"errors"
	"fmt"
	"time"

	"github.com/buraksezer/olric"
	"github.com/testground/sdk-go/run"
	"github.com/testground/sdk-go/runtime"
)

func DMapGetPut(r *runtime.RunEnv, _ *run.InitContext) error {
	f := func(db *olric.Olric) error {
		dm, err := db.NewDMap("standalone.get-put")
		if err != nil {
			r.RecordFailure(fmt.Errorf("olric.NewDMap returned an error: %w", err))
			return err
		}

		for i := 0; i < 100000; i++ {
			err = dm.Put(toKey(i), toVal(i))
			if err != nil {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 100000; i++ {
			value, err := dm.Get(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
			if !bytes.Equal(value.([]byte), toVal(i)) {
				r.RecordFailure(fmt.Errorf("value is different for: %s", toKey(i)))
			}
		}

		_, err = dm.Get("foobar")
		if errors.Is(err, olric.ErrKeyNotFound) {
			r.RecordFailure(fmt.Errorf("expected olric.ErrKeyNotFound. got: %w", err))
			return err
		}

		return nil
	}

	return olricNode(r, f)
}

func DMapPutDelete(r *runtime.RunEnv, _ *run.InitContext) error {
	f := func(db *olric.Olric) error {
		dm, err := db.NewDMap("standalone.put-delete")
		if err != nil {
			r.RecordFailure(fmt.Errorf("olric.NewDMap returned an error: %w", err))
			return err
		}

		for i := 0; i < 100000; i++ {
			err = dm.Put(toKey(i), toVal(i))
			if err != nil {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 100000; i++ {
			err := dm.Delete(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
		}

		for i := 0; i < 100000; i++ {
			_, err = dm.Get(toKey(i))
			if !errors.Is(err, olric.ErrKeyNotFound) {
				r.RecordFailure(err)
				continue
			}
		}
		return nil
	}

	return olricNode(r, f)
}

func DMapPutEx(r *runtime.RunEnv, _ *run.InitContext) error {
	f := func(db *olric.Olric) error {
		dm, err := db.NewDMap("standalone.DMapPutEx")
		if err != nil {
			r.RecordFailure(fmt.Errorf("olric.NewDMap returned an error: %w", err))
			return err
		}

		for i := 0; i < 1000; i++ {
			err = dm.PutEx(toKey(i), toVal(i), 250*time.Millisecond)
			if err != nil {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 1000; i++ {
			value, err := dm.Get(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
			if !bytes.Equal(value.([]byte), toVal(i)) {
				r.RecordFailure(fmt.Errorf("value is different for: %s", toKey(i)))
			}
		}

		<-time.After(250 * time.Millisecond)

		for i := 0; i < 1000; i++ {
			_, err = dm.Get(toKey(i))
			if !errors.Is(err, olric.ErrKeyNotFound) {
				r.RecordFailure(err)
				continue
			}
		}
		return nil
	}

	return olricNode(r, f)
}

func DMapPutIf(r *runtime.RunEnv, _ *run.InitContext) error {
	f := func(db *olric.Olric) error {
		dm, err := db.NewDMap("standalone.DMapPutIf")
		if err != nil {
			r.RecordFailure(fmt.Errorf("olric.NewDMap returned an error: %w", err))
			return err
		}

		for i := 0; i < 1000; i++ {
			err = dm.PutIf(toKey(i), toVal(i), olric.IfNotFound)
			if err != nil {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 1000; i++ {
			value, err := dm.Get(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
			if !bytes.Equal(value.([]byte), toVal(i)) {
				r.RecordFailure(fmt.Errorf("value is different for: %s", toKey(i)))
			}
		}

		for i := 0; i < 1000; i++ {
			err = dm.PutIf(toKey(i), fmt.Sprintf("%s-ifnotfound", toVal(i)), olric.IfNotFound)
			if err != nil && !errors.Is(err, olric.ErrKeyFound) {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 1000; i++ {
			value, err := dm.Get(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
			if !bytes.Equal(value.([]byte), toVal(i)) {
				r.RecordFailure(fmt.Errorf("value is different for: %s", toKey(i)))
			}
		}

		for i := 0; i < 1000; i++ {
			err = dm.PutIf(toKey(i), fmt.Sprintf("%s-iffound", toVal(i)), olric.IfFound)
			if err != nil {
				r.RecordFailure(err)
			}
		}

		for i := 0; i < 1000; i++ {
			value, err := dm.Get(toKey(i))
			if err != nil {
				r.RecordFailure(err)
				continue
			}
			val := fmt.Sprintf("%s-iffound", toVal(i))
			if value.(string) != val {
				r.RecordFailure(fmt.Errorf("value is different for: %s", toKey(i)))
			}
		}

		return nil
	}

	return olricNode(r, f)
}
