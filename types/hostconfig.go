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

// RestartPolicy is mapped according to isulad's definition.
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// BlockIOWeightDevice is mapped according to isulad's definition.
type BlockIOWeightDevice struct {
	Path   string
	Weight uint16
}

// BlockIODeviceReadBps is mapped according to isulad's definition.
type BlockIODeviceReadBps struct {
	Path string
	Rate uint64
}

// BlockIODeviceWriteBps is mapped according to isulad's definition.
type BlockIODeviceWriteBps struct {
	Path string
	Rate uint64
}

// Ulimit is mapped according to isulad's definition.
type Ulimit struct {
	Name string
	Hard int64
	Soft int64
}

// Hugetlb is mapped according to isulad's definition.
type Hugetlb struct {
	PageSize string
	Limit    uint64
}

// Device is mapped according to isulad's definition.
type Device struct {
	CgroupPermissions string
	PathInContainer   string
	PathOnHost        string
}

// HostChannel is mapped according to isulad's definition.
type HostChannel struct {
	PathOnHost      string
	PathInContainer string
	Permissions     string
	Size            uint64
}

// IsuladHostConfig maps the container hostconfig of isulad
// its structure is consistent with isulad
type IsuladHostConfig struct {
	Binds               []string                 `json:"Binds,omitempty"`
	NetworkMode         string                   `json:"NetworkMode,omitempty"`
	GroupAdd            []string                 `json:"GroupAdd,omitempty"`
	IpcMode             string                   `json:"IpcMode,omitempty"`
	PidMode             string                   `json:"PidMode,omitempty"`
	Privileged          bool                     `json:"Privileged,omitempty"`
	SystemContainer     bool                     `json:"SystemContainer,omitempty"`
	NsChangeFiles       []string                 `json:"NsChangeFiles,omitempty"`
	UserRemap           string                   `json:"UserRemap,omitempty"`
	ShmSize             int64                    `json:"ShmSize,omitempty"`
	AutoRemove          bool                     `json:"AutoRemove,omitempty"`
	AutoRemoveBak       bool                     `json:"AutoRemoveBak,omitempty"`
	ReadonlyRootfs      bool                     `json:"ReadonlyRootfs,omitempty"`
	UTSMode             string                   `json:"UTSMode,omitempty"`
	UsernsMode          string                   `json:"UsernsMode,omitempty"`
	Sysctls             map[string]string        `json:"Sysctls,omitempty"`
	Runtime             string                   `json:"Runtime,omitempty"`
	RestartPolicy       *RestartPolicy           `json:"RestartPolicy,omitempty"`
	CapAdd              []string                 `json:"CapAdd,omitempty"`
	CapDrop             []string                 `json:"CapDrop,omitempty"`
	DNS                 []string                 `json:"Dns,omitempty"`
	DNSOptions          []string                 `json:"DnsOptions,omitempty"`
	DNSSearch           []string                 `json:"DnsSearch,omitempty"`
	ExtraHosts          []string                 `json:"ExtraHosts,omitempty"`
	HookSpec            string                   `json:"HookSpec,omitempty"`
	CPUShares           int64                    `json:"CPUShares,omitempty"`
	Memory              int64                    `json:"Memory,omitempty"`
	OomScoreAdj         int                      `json:"OomScoreAdj,omitempty"`
	BlkioWeight         uint16                   `json:"BlkioWeight,omitempty"`
	BlkioWeightDevice   []*BlockIOWeightDevice   `json:"BlkioWeightDevice,omitempty"`
	BlkioDeviceReadBps  []*BlockIODeviceReadBps  `json:"BlkioDeviceReadBps,omitempty"`
	BlkioDeviceWriteBps []*BlockIODeviceWriteBps `json:"BlkioDeviceWriteBps,omitempty"`
	CPUPeriod           int64                    `json:"CPUPeriod,omitempty"`
	CPUQuota            int64                    `json:"CPUQuota,omitempty"`
	CPURealtimePeriod   int64                    `json:"CPURealtimePeriod,omitempty"`
	CPURealtimeRuntime  int64                    `json:"CPURealtimeRuntime,omitempty"`
	CpusetCpus          string                   `json:"CpusetCpus,omitempty"`
	CpusetMems          string                   `json:"CpusetMems,omitempty"`
	Devices             []*Device                `json:"Devices,omitempty"`
	SecurityOpt         []string                 `json:"SecurityOpt,omitempty"`
	StorageOpt          map[string]string        `json:"StorageOpt,omitempty"`
	KernelMemory        int64                    `json:"KernelMemory,omitempty"`
	MemoryReservation   int64                    `json:"MemoryReservation,omitempty"`
	MemorySwap          int64                    `json:"MemorySwap,omitempty"`
	OOMKillDisable      bool                     `json:"OomKillDisable,omitempty"`
	PidsLimit           int64                    `json:"PidsLimit,omitempty"`
	FilesLimit          int64                    `json:"FilesLimit,omitempty"`
	Ulimits             []*Ulimit                `json:"Ulimits,omitempty"`
	Hugetlbs            []*Hugetlb               `json:"Hugetlbs,omitempty"`
	HostChannel         *HostChannel             `json:"HostChannel,omitempty"`
	EnvTargetFile       string                   `json:"EnvTargetFile,omitempty"`
	ExternalRootfs      string                   `json:"ExternalRootfs,omitempty"`
	CgroupParent        string                   `json:"CgroupParent,omitempty"`
}
