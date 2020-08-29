/*
 * Copyright (c) 2020 Huawei Technologies Co., Ltd.
 * isula-transform is licensed under the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Create: 2020-05-18
 */

package docker

import (
	"context"
	"sync"
	"sync/atomic"
)

const (
	executed   int32 = 1
	unexecuted int32 = 0
)

type rollbackFunc func()

type rollback struct {
	wg      *sync.WaitGroup
	st      *int32
	ctx     context.Context
	rbFuncs []rollbackFunc

	closed  bool
	closeCh chan bool
}

// newRollback returns a rollback instance
func newRollback(ctx context.Context, wg *sync.WaitGroup) *rollback {
	initSt := unexecuted
	return &rollback{
		wg:      wg,
		ctx:     ctx,
		st:      &initSt,
		closeCh: make(chan bool, 1),
	}
}

func (rb *rollback) wait() {
	rb.wg.Add(1)
	go func() {
		defer rb.wg.Done()
		for {
			select {
			case <-rb.ctx.Done():
				rb.run()
				return
			case rb.closed = <-rb.closeCh:
				return
			}
		}
	}()
}

func (rb *rollback) close() {
	rb.closeCh <- true
	close(rb.closeCh)
}

func (rb *rollback) register(f rollbackFunc) {
	rb.rbFuncs = append(rb.rbFuncs, f)
}

func (rb *rollback) run() {
	if rb.closed || atomic.LoadInt32(rb.st) == executed {
		return
	}
	atomic.StoreInt32(rb.st, executed)
	for i := len(rb.rbFuncs) - 1; i >= 0; i-- {
		rb.rbFuncs[i]()
	}
}
