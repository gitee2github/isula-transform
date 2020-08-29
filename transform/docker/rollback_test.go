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
	"container/list"
	"context"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_rollback(t *testing.T) {
	Convey("Test rollback", t, func() {
		wg := new(sync.WaitGroup)

		Convey("Test rollback with cancel", func() {
			l := list.New()
			ctx, cancel := context.WithCancel(context.Background())
			rb := newRollback(ctx, wg)
			rb.wait()
			rb.register(func() {
				l.PushBack("first register, last execute")
			})
			rb.register(func() {
				l.PushBack("last register, first execute")
			})
			cancel()
			wg.Wait()
			rb.close()
			So(l.Len(), ShouldEqual, 2)
			So(l.Front().Value, ShouldEqual, "last register, first execute")
			l.Remove(l.Front())
			So(l.Front().Value, ShouldEqual, "first register, last execute")
			l.Remove(l.Front())
		})

		Convey("Test rollback with close", func() {
			l := list.New()
			ctx, cancel := context.WithCancel(context.Background())
			rb := newRollback(ctx, wg)
			rb.wait()
			rb.register(func() {
				l.PushBack("first register, last execute")
			})
			rb.register(func() {
				l.PushBack("last register, first execute")
			})
			rb.close()
			cancel()
			wg.Wait()
			So(l.Len(), ShouldEqual, 0)
		})
	})
}
