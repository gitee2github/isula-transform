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

package main

import "github.com/urfave/cli"

var basicFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "log",
		Usage: "specific output log file path",
		Value: "/var/log/isula-kits/transform.log",
	},
	cli.StringFlag{
		Name:  "log-level",
		Usage: "Customize the level of logging for collection, allowed: debug, info, warn, error",
		Value: "info",
	},
	cli.StringFlag{
		Name:   "container-type",
		Usage:  "origin container type",
		Value:  "docker",
		Hidden: true,
	},
	cli.StringFlag{
		Name:  "isulad-config-file",
		Usage: "iSulad configuration file path",
		Value: "/etc/isulad/daemon.json",
	},
}

var dockerFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "docker-graph",
		Usage: "graph root of docker",
		Value: "/var/lib/docker",
	},
	cli.StringFlag{
		Name:  "docker-state",
		Usage: "state root of docker",
		Value: "/var/run/docker",
	},
}

var containerFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "all",
		Usage: "transform all containers",
	},
}

var transformFlags = [][]cli.Flag{basicFlags, dockerFlags, containerFlags}
