// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package native

import (
	"encoding/json"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

func init() {
	register("sloadTracer", newStorageTracer)
}

type storageState = map[common.Address]*storageAccount
type storageAccount struct {
	Storage map[common.Hash]common.Hash
}

type storageTracer struct {
	env       *vm.EVM
	prestate  storageState
	create    bool
	to        common.Address
	interrupt uint32 // Atomic flag to signal execution interruption
	reason    error  // Textual reason for the interruption
	Output    string
	Error     string
}

type tracerResult struct {
	Prestate storageState `json:"storage"`
	Output   string       `json:"output"`
	Error    string       `json:"error"`
}

func newStorageTracer() tracers.Tracer {
	// First callframe contains tx context info
	// and is populated on start and end.
	return &storageTracer{prestate: storageState{}}
}

// CaptureStart implements the EVMLogger interface to initialize the tracing operation.
func (t *storageTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	t.env = env
	t.create = create
	t.to = to

	t.lookupAccount(from)
	t.lookupAccount(to)
}

// CaptureEnd is called after the call finishes to finalize the tracing.
func (t *storageTracer) CaptureEnd(output []byte, gasUsed uint64, _ time.Duration, err error) {
	if t.create {
		// Exclude created contract.
		delete(t.prestate, t.to)
	}

	if err != nil {
		t.Error = err.Error()
		if err.Error() == "execution reverted" && len(output) > 0 {
			t.Output = bytesToHex(output)
		}
	} else {
		t.Output = bytesToHex(output)
	}
}

func (t *storageTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	stack := scope.Stack
	stackData := stack.Data()
	stackLen := len(stackData)
	switch {
	case stackLen >= 1 && op == vm.SLOAD:
		slot := common.Hash(stackData[stackLen-1].Bytes32())
		t.lookupStorage(scope.Contract.Address(), slot)
	}
}

// CaptureFault implements the EVMLogger interface to trace an execution fault.
func (t *storageTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, _ *vm.ScopeContext, depth int, err error) {
}

// CaptureEnter is called when EVM enters a new scope (via call, create or selfdestruct).
func (t *storageTracer) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

// CaptureExit is called when EVM exits a scope, even if the scope didn't
// execute any code.
func (t *storageTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *storageTracer) GetResult() (json.RawMessage, error) {

	_res := tracerResult{
		Prestate: t.prestate,
		Error:    t.Error,
		Output:   t.Output,
	}

	res, err := json.Marshal(_res)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(res), t.reason
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *storageTracer) Stop(err error) {
	t.reason = err
	atomic.StoreUint32(&t.interrupt, 1)
}

// lookupAccount fetches details of an account and adds it to the prestate
// if it doesn't exist there.
func (t *storageTracer) lookupAccount(addr common.Address) {
	if _, ok := t.prestate[addr]; ok {
		return
	}
	t.prestate[addr] = &storageAccount{
		Storage: make(map[common.Hash]common.Hash),
	}
}

// lookupStorage fetches the requested storage slot and adds
// it to the prestate of the given contract. It assumes `lookupAccount`
// has been performed on the contract before.
func (t *storageTracer) lookupStorage(addr common.Address, key common.Hash) {
	if _, ok := t.prestate[addr].Storage[key]; ok {
		return
	}
	t.prestate[addr].Storage[key] = t.env.StateDB.GetState(addr, key)
}