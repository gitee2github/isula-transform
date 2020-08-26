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

// Package transform defines the transformer interface and
// provides a common tool for transform to an iSulad container
package transform

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// container type
const (
	// DOCKER
	DOCKER = "docker"
	// CONTAINERD
	CONTAINERD = "containerd"
	// CRIO
	CRIO = "cri-o"
)

// NewFunc create new Transformer
type NewFunc func(ctx *cli.Context) Transformer

var (
	transformers map[string]NewFunc
)

func init() {
	transformers = make(map[string]NewFunc)
}

// Register registers a transformer factory func
func Register(typ string, newFunc NewFunc) {
	transformers[typ] = newFunc
}

// Result contains the success of and the output of the transformation
type Result struct {
	Msg string
	Ok  bool
}

// Transformer defines common container transform engine interface
type Transformer interface {
	Init() error
	Transform([]string, bool, chan Result)
}

// BaseTransformer contains the base members of transformer
type BaseTransformer struct {
	Name      string
	GraphRoot string
	StateRoot string
}

// EngineOpt allows configuring a BaseEngineCfg
type EngineOpt func(e *BaseTransformer)

// EngineWithGraph sets the graph root of the BaseEngineCfg
func EngineWithGraph(graphRoot string) EngineOpt {
	return func(e *BaseTransformer) {
		e.GraphRoot = graphRoot
	}
}

// EngineWithState sets the statr root of the BaseEngineCfg
func EngineWithState(stateRoot string) EngineOpt {
	return func(e *BaseTransformer) {
		e.StateRoot = stateRoot
	}
}

// GetTransformer returns the specified transformer
func GetTransformer(ctx *cli.Context) Transformer {
	typ := ctx.GlobalString("container-type")
	if newFunc, exist := transformers[typ]; exist {
		return newFunc(ctx)
	}
	logrus.Errorf("not support container type: %s", typ)
	return nil
}
