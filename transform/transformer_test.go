/*
 * Copyright (c) 2020 Huawei Technologies Co., Ltd.
 * isula-transform is licensed under the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Create: 2020-04-24
 */

package transform

import (
	"flag"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli"
)

type mockTransformer struct{}

func newMockTransformer(*cli.Context) Transformer { return &mockTransformer{} }

func (mt *mockTransformer) Init() error { return nil }

func (mt *mockTransformer) Transform([]string, bool, chan Result) {}

func TestTransformers(t *testing.T) {
	defer delete(transformers, "mock")
	Convey("TestTransformers", t, func() {
		Convey("TestRegister", func() {
			Register("mock", newMockTransformer)
			So(transformers, ShouldContainKey, "mock")
		})

		Convey("TestGetTransformer", func() {
			flags := flag.NewFlagSet("", flag.ContinueOnError)
			typeFlag := cli.StringFlag{Name: "container-type"}

			Convey("Get support transformer", func() {
				typeFlag.Value = "mock"
				typeFlag.Apply(flags)
				ctx := cli.NewContext(nil, flags, nil)
				So(GetTransformer(ctx), ShouldNotBeNil)
			})

			Convey("Not support transformer", func() {
				typeFlag.Value = "isulad"
				typeFlag.Apply(flags)
				ctx := cli.NewContext(nil, flags, nil)
				So(GetTransformer(ctx), ShouldBeNil)
			})
		})
	})
}

func TestEngineOpt(t *testing.T) {
	Convey("TestEngineOpt", t, func() {
		base := &BaseTransformer{}

		Convey("TestEngineWithGraph", func() {
			graph := "/test/lib/isulad"
			opt := EngineWithGraph(graph)
			opt(base)
			So(base.GraphRoot, ShouldEqual, graph)
		})

		Convey("TestEngineWithState", func() {
			state := "/test/run/isulad"
			opt := EngineWithState(state)
			opt(base)
			So(base.StateRoot, ShouldEqual, state)
		})
	})
}
