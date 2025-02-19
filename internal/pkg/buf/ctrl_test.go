// Copyright 2022 Metrika Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buf

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestControllerDrainBatch uses:
// - a controller with a priority buffer of size bufSz,
//   n pre inserted metrics and batch size 1.
// - a goroutine to run the controller
// and after stopping the controller checks:
// - drain callback batch size is equal to MaxDrainBatchLen
func TestControllerDrainBatch(t *testing.T) {
	n := 5

	drainCh := make(chan ItemBatch, n)
	onDrain := func(b ItemBatch) error {
		drainCh <- b

		return nil
	}

	conf := ControllerConf{
		BufDrainFreq:        1 * time.Millisecond,
		BufLenLimit:         n,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()

	select {
	case b := <-drainCh:
		require.Len(t, b, conf.BufLenLimit)
		require.Equal(t, b, m)
		require.Len(t, drainCh, 0)
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for drain callback")
	}

	pb = NewPriorityBuffer(time.Duration(0))
	err = pb.Insert(m...)
	require.NoError(t, err)

	conf.BufLenLimit = 1
	ctrl = NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()
	require.Equal(t, 0, pb.Len())

	for i := 0; i < n; i++ {
		select {
		case b := <-drainCh:
			require.Len(t, b, conf.BufLenLimit)
			require.Equal(t, b[0], m[i])
		case <-time.After(100 * time.Millisecond):
			t.Error("timeout waiting for drain callback")
		}
	}
}

func TestControllerDrainCallback(t *testing.T) {
	n := 5

	drainCh := make(chan ItemBatch, n)
	onDrain := func(b ItemBatch) error {
		drainCh <- b

		return nil
	}

	conf := ControllerConf{
		BufDrainFreq:        1 * time.Millisecond,
		BufLenLimit:         n,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()

	select {
	case b := <-drainCh:
		require.Equal(t, b, m)
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for drain callback")
	}
	require.Equal(t, 0, pb.Len())
}

func TestControllerDrainCallbackErr(t *testing.T) {
	n := 5
	onDrain := func(b ItemBatch) error {
		return fmt.Errorf("drain test error")
	}

	conf := ControllerConf{
		BufDrainFreq:        75 * time.Millisecond,
		BufLenLimit:         n,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()
	require.Equal(t, n+1, pb.Len()) // +1 to compensate for agent.net.error event
}

// TestControllerDrain uses:
// - a controller with a priority buffer of size bufSz,
//   n pre inserted metrics and batch size 2*n.
// - a goroutine to run the controller
// and checks:
// - batch smaller than MaxDrainBatchLen is evicted periodically
func TestControllerDrain(t *testing.T) {
	n := 5

	drainCh := make(chan ItemBatch, n)
	onDrain := func(b ItemBatch) error {
		drainCh <- b

		return nil
	}

	conf := ControllerConf{
		BufDrainFreq:        1 * time.Millisecond,
		BufLenLimit:         2 * n,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()

	select {
	case b := <-drainCh:
		require.Equal(t, b, m)
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for drain callback")
	}
	require.Equal(t, 0, pb.Len())
}

func TestControllerClose(t *testing.T) {
	n := 5

	onDrain := func(b ItemBatch) error {
		return nil
	}

	conf := ControllerConf{
		BufDrainFreq:        1 * time.Millisecond,
		BufLenLimit:         1,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	go ctrl.Start()
	<-time.After(100 * time.Millisecond)

	ctrl.Stop()

	assert.Equal(t, 0, pb.Len())
}

func TestController_HeapAllocLimitError(t *testing.T) {
	n := 5

	onDrain := func(b ItemBatch) error {
		return nil
	}

	conf := ControllerConf{
		BufDrainFreq:        1 * time.Millisecond,
		BufLenLimit:         1,
		OnBufRemoveCallback: onDrain,
	}

	pb := NewPriorityBuffer(time.Duration(0))
	m := newTestItemBatch(n)

	err := pb.Insert(m...)
	require.NoError(t, err)

	ctrl := NewController(conf, pb)
	// update cached memstats to mimic the limit error
	ctrl.ControllerConf.MaxHeapAllocBytes = 100
	ctrl.memstats.HeapAlloc = 200
	ctrl.memstatsUpdatedAt = time.Now()

	// re-insert should be rejected
	err = ctrl.BufInsert(m[0])
	require.Error(t, err)
	require.IsType(t, ErrHeapAllocLimit, err)
}
