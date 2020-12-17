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

package types

import "time"

// Mount is mapped according to isulad's definition.
type Mount struct {
	Type        string `json:",omitempty"`
	Destination string `json:",omitempty"`
	Driver      string `json:",omitempty"`
	Key         string `json:",omitempty"`
	Name        string `json:",omitempty"`
	Named       string `json:",omitempty"`
	Propagation string `json:",omitempty"`
	RW          bool   `json:",omitempty"`
	Relabel     string `json:",omitempty"`
	Source      string `json:",omitempty"`
}

// HealthCheckCfg is mapped according to isulad's definition.
type HealthCheckCfg struct {
	Test            []string
	Interval        int64
	Timeout         int64
	StartPeriod     int64
	Retries         int
	ExitOnUnhealthy bool
}

// ContainerCfg is mapped according to isulad's definition.
type ContainerCfg struct {
	Hostname        string              `json:"hostname,omitempty"`
	DomainName      string              `json:"DomainName,omitempty"`
	User            string              `json:"User,omitempty"`
	AttachStdin     bool                `json:"AttachStdin,omitempty"`
	AttachStdout    bool                `json:"AttachStdout,omitempty"`
	AttachStderr    bool                `json:"AttachStderr,omitempty"`
	ExposedPorts    map[string]struct{} `json:"ExposedPorts,omitempty"`
	PublishService  string              `json:"PublishService,omitempty"`
	Tty             bool                `json:"Tty,omitempty"`
	OpenStdin       bool                `json:"OpenStdin,omitempty"`
	StdinOnce       bool                `json:"StdinOnce,omitempty"`
	Env             []string            `json:"Env,omitempty"`
	Cmd             []string            `json:"Cmd,omitempty"`
	ArgsEscaped     bool                `json:"ArgsEscaped,omitempty"`
	NetworkDisabled bool                `json:"NetworkDisabled,omitempty"`
	Image           string              `json:"Image,omitempty"`
	Volumes         map[string]struct{} `json:"Volumes,omitempty"`
	WorkingDir      string              `json:"WorkingDir,omitempty"`
	Entrypoint      []string            `json:"Entrypoint,omitempty"`
	MacAddress      string              `json:"MacAddress,omitempty"`
	Onbuild         []string            `json:"Onbuild,omitempty"`
	Labels          map[string]string   `json:"Labels,omitempty"`
	LogDriver       string              `json:"LogDriver,omitempty"`
	Annotations     map[string]string   `json:"Annotations,omitempty"`
	StopSignal      string              `json:"StopSignal,omitempty"`
	HealthCheck     *HealthCheckCfg     `json:"Healthcheck,omitempty"`
	SystemContainer bool                `json:"SystemContainer,omitempty"`
	NsChangeOpt     string              `json:"NsChangeOpt,omitempty"`
	Mounts          map[string]string   `json:"Mounts,omitempty"`
}

// HealthLog is mapped according to isulad's definition.
type HealthLog struct {
	Start    string `json:",omitempty"`
	End      string `json:",omitempty"`
	ExitCode int    `json:",omitempty"`
	Output   string `json:",omitempty"`
}

// HealthCfg is mapped according to isulad's definition.
type HealthCfg struct {
	Status        string      `json:",omitempty"`
	FailingStreak int         `json:",omitempty"`
	Log           []HealthLog `json:",omitempty"`
}

// ContainerState is mapped according to isulad's definition.
type ContainerState struct {
	Dead              bool       `json:"Dead,omitempty"`
	Error             string     `json:"Error,omitempty"`
	ExitCode          int        `json:"ExitCode,omitempty"`
	FinishedAt        time.Time  `json:"FinishedAt,omitempty"`
	OOMKilled         bool       `json:"OomKilled,omitempty"`
	Paused            bool       `json:"Paused,omitempty"`
	Pid               int        `json:"Pid,omitempty"`
	PPid              int        `json:"PPid,omitempty"`
	StartTime         uint64     `json:"StartTime,omitempty"`
	PStartTime        uint64     `json:"PStartTime,omitempty"`
	RemovalInprogress bool       `json:"RemovalInprogress,omitempty"`
	Restarting        bool       `json:"Restarting,omitempty"`
	Running           bool       `json:"Running,omitempty"`
	StartedAt         time.Time  `json:"StartedAt,omitempty"`
	Starting          bool       `json:"Starting,omitempty"`
	Health            *HealthCfg `json:"Health,omitempty"`
}

// CommonConfig is mapped according to isulad's definition.
type CommonConfig struct {
	Path                   string           `json:"Path,omitempty"`
	Args                   []string         `json:"Args,omitempty"`
	Config                 *ContainerCfg    `json:"Config,omitempty"`
	Created                time.Time        `json:"Created,omitempty"`
	HasBeenManuallyStopped bool             `json:"HasBeenManuallyStopped,omitempty"`
	HasBeenStartedBefore   bool             `json:"HasBeenStartedBefore,omitempty"`
	Image                  string           `json:"Image,omitempty"`
	ImageType              string           `json:"ImageType,omitempty"`
	HostnamePath           string           `json:"HostnamePath,omitempty"`
	HostsPath              string           `json:"HostsPath,omitempty"`
	ResolvConfPath         string           `json:"ResolvConfPath,omitempty"`
	ShmPath                string           `json:"ShmPath,omitempty"`
	LogPath                string           `json:"LogPath,omitempty"`
	BaseFs                 string           `json:"BaseFs,omitempty"`
	MountPoints            map[string]Mount `json:"MountPoints,omitempty"`
	Name                   string           `json:"Name"`
	RestartCount           int              `json:"RestartCount,omitempty"`
	ID                     string           `json:"id"`

	MountLabel      string
	ProcessLabel    string
	SeccompProfile  string
	NoNewPrivileges bool

	// Backup network files path
	OriginHostnamePath   string `json:"-"`
	OriginHostsPath      string `json:"-"`
	OriginResolvConfPath string `json:"-"`
}

// GetOriginNetworkFile returns the path specified file in host, hostname and resolv.conf
func (cc *CommonConfig) GetOriginNetworkFile(name string) string {
	switch name {
	case Hostname:
		return cc.OriginHostnamePath
	case Hosts:
		return cc.OriginHostsPath
	case Resolv:
		return cc.OriginResolvConfPath
	}
	return ""
}

// IsuladV2Config maps the isulad config.v2.json
// its structure is consistent with isulad
type IsuladV2Config struct {
	CommonConfig *CommonConfig   `json:"CommonConfig,omitempty"`
	Image        string          `json:"Image,omitempty"`
	State        *ContainerState `json:"State,omitempty"`
}
