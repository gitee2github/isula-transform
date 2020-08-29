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
 *
 * Since some of this code is derived from docker, their copyright
 * is retained here....
 *
 *   Copyright 2013-2018 Docker, Inc.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 *
 * The original version of the DockerV2Config struct can be found at
 * https://github.com/docker/engine/blob/9552f2b2fddeb0c2537b350f4b159ffe525d7a42/container/container.go
 *
 */

package types

import "time"

// DockerV2Config maps the docker config.v2.json
type DockerV2Config struct {
	State                  *ContainerState
	ID                     string
	Created                time.Time
	Path                   string
	Args                   []string
	CgroupParent           string
	Config                 *ContainerCfg
	ImageID                string `json:"Image"`
	LogPath                string
	Name                   string
	Driver                 string
	OS                     string
	MountLabel             string
	ProcessLabel           string
	RestartCount           int
	MountPoints            map[string]*Mount
	AppArmorProfile        string
	HostnamePath           string
	HostsPath              string
	ShmPath                string
	ResolvConfPath         string
	SeccompProfile         string
	Managed                bool
	HasBeenStartedBefore   bool
	HasBeenManuallyStopped bool
	NoNewPrivileges        bool
}
