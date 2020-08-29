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

package docker

import (
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	. "github.com/google/go-cmp/cmp"
	"github.com/opencontainers/runtime-spec/specs-go"
	. "github.com/smartystreets/goconvey/convey"
	"isula.org/isula-transform/types"
)

const (
	reconcileTestCtrID              = "isulatransformreconciletestcontainer"
	reconcileTestConnectCtrID       = "1234567890123456789012345678901234567890123456789012345678901234"
	reconcileTestImage              = "isulatransformreconciletestcontainer:image"
	startUnixTimeStamp        int64 = 1579744800
)

func Test_reconcileV2Config(t *testing.T) {
	Convey("Test_reconcileV2Config", t, func() {
		baseDockerPath := "/var/lib/docker/containers/" + reconcileTestCtrID
		baseLcrPath := "/var/lib/isulad/engines/lcr/" + reconcileTestCtrID
		baseCfg := &types.IsuladV2Config{
			CommonConfig: &types.CommonConfig{
				Name:           "/test_container",
				HostnamePath:   baseDockerPath + "/hostname",
				HostsPath:      baseDockerPath + "/hosts",
				ResolvConfPath: baseDockerPath + "/resolv.conf",
				Config: &types.ContainerCfg{
					Annotations: make(map[string]string),
				},
				Created: time.Unix(startUnixTimeStamp, 000000000),
				ID:      reconcileTestCtrID,
			},
			State: &types.ContainerState{
				Running:   true,
				StartedAt: time.Unix(startUnixTimeStamp, 000000000),
			},
		}
		expectCfg := &types.IsuladV2Config{
			CommonConfig: &types.CommonConfig{
				Name:                 "test_container",
				OriginHostnamePath:   baseDockerPath + "/hostname",
				OriginHostsPath:      baseDockerPath + "/hosts",
				OriginResolvConfPath: baseDockerPath + "/resolv.conf",
				HostnamePath:         baseLcrPath + "/hostname",
				HostsPath:            baseLcrPath + "/hosts",
				ResolvConfPath:       baseLcrPath + "/resolv.conf",
				ShmPath:              baseLcrPath + "/mounts/shm",
				Config: &types.ContainerCfg{
					Annotations: map[string]string{
						"rootfs.mount": "/var/lib/isulad/mnt/rootfs",
					},
				},
				Created: time.Unix(startUnixTimeStamp, 000000000).Local(),
				ID:      reconcileTestCtrID,
			},
			State: &types.ContainerState{
				Running:   false,
				StartedAt: time.Unix(startUnixTimeStamp, 000000000).Local(),
			},
		}

		Convey("reconcile image", func() {
			expectCfg.CommonConfig.Image = reconcileTestImage
			expectCfg.CommonConfig.ImageType = "oci"
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithImage(reconcileTestImage + "@sha256:b41eab535e9ce1239966da529a85e624d117f4d2f8911827d0d7c1e93e92a3e3"),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			// types.ContainerState.FinishedAt use time.Now().Local()
			// need to sync from base
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})

		Convey("docker CgroupParent", func() {
			originCgroupParent := "/docker/" + reconcileTestCtrID
			expectCfg.CommonConfig.Config.Annotations["cgroup.dir"] = "/isulad"
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithCgroupParent(originCgroupParent),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})

		Convey("user CgroupParent", func() {
			originCgroupParent := "/test/" + reconcileTestCtrID
			expectCfg.CommonConfig.Config.Annotations["cgroup.dir"] = "/test"
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithCgroupParent(originCgroupParent),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})

		Convey("json-file log driver", func() {
			expectCfg.CommonConfig.LogDriver = logDriverJSONFile
			expectCfg.CommonConfig.LogPath = baseLcrPath + "/console.log"
			expectCfg.CommonConfig.Config.Annotations["log.console.driver"] = logDriverJSONFile
			expectCfg.CommonConfig.Config.Annotations["log.console.file"] = baseLcrPath + "/console.log"

			Convey("user define", func() {
				logConfig := &container.LogConfig{
					Type: logDriverJSONFile,
					Config: map[string]string{
						"max-file": "3",
						"max-size": "10M",
						"env":      "not retain",
					},
				}
				expectCfg.CommonConfig.Config.Annotations["log.console.filerotate"] = "3"
				expectCfg.CommonConfig.Config.Annotations["log.console.filesize"] = "10M"
				opts := []v2ConfigReconcileOpt{
					v2ConfigWithLogConfig(logConfig, baseLcrPath),
				}
				reconcileV2Config(baseCfg, baseLcrPath, opts...)
				expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
				So(Diff(baseCfg, expectCfg), ShouldBeBlank)
			})

			Convey("use default", func() {
				logConfig := &container.LogConfig{
					Type:   "json-file",
					Config: map[string]string{},
				}
				expectCfg.CommonConfig.Config.Annotations["log.console.filerotate"] = "7"
				expectCfg.CommonConfig.Config.Annotations["log.console.filesize"] = defaultLogSize
				opts := []v2ConfigReconcileOpt{
					v2ConfigWithLogConfig(logConfig, baseLcrPath),
				}
				reconcileV2Config(baseCfg, baseLcrPath, opts...)
				expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
				So(Diff(baseCfg, expectCfg), ShouldBeBlank)
			})
		})

		Convey("syslog log driver", func() {
			logConfig := &container.LogConfig{
				Type: "syslog",
				Config: map[string]string{
					"tag":             "test",
					"syslog-facility": "local1",
					"env":             "not retain",
				},
			}
			expectCfg.CommonConfig.LogDriver = logDriverSyslog
			expectCfg.CommonConfig.Config.Annotations["log.console.driver"] = logDriverSyslog
			expectCfg.CommonConfig.Config.Annotations["log.console.tag"] = "test"
			expectCfg.CommonConfig.Config.Annotations["log.console.facility"] = "local1"
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithLogConfig(logConfig, baseLcrPath),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})

		Convey("not support log driver", func() {
			logConfig := &container.LogConfig{
				Type: "journald",
				Config: map[string]string{
					"tag": "test",
				},
			}
			expectCfg.CommonConfig.LogDriver = defaultLogDriver
			expectCfg.CommonConfig.LogPath = defaultLogPath
			expectCfg.CommonConfig.Config.Annotations["log.console.driver"] = defaultLogDriver
			expectCfg.CommonConfig.Config.Annotations["log.console.file"] = defaultLogPath
			expectCfg.CommonConfig.Config.Annotations["log.console.filerotate"] = "7"
			expectCfg.CommonConfig.Config.Annotations["log.console.filesize"] = defaultLogSize
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithLogConfig(logConfig, baseLcrPath),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})

		Convey("oom score adj", func() {
			oomScore := 100
			expectCfg.CommonConfig.Config.Annotations["proc.oom_score_adj"] = "100"
			opts := []v2ConfigReconcileOpt{
				v2ConfigWithOomScoreAdj(oomScore),
			}
			reconcileV2Config(baseCfg, baseLcrPath, opts...)
			expectCfg.State.FinishedAt = baseCfg.State.FinishedAt
			So(Diff(baseCfg, expectCfg), ShouldBeBlank)
		})
	})
}

func Test_imageRemoveSuffixDigest(t *testing.T) {
	var (
		baseImage     = "ubuntu"
		imgTag        = ":18.04"
		imgDigest     = "@sha256:dc6b3e3cf28225d72351d5dbddc35ea08a08ad83725043903df61448c9e466a0"
		withTag       = baseImage + imgTag
		withDigest    = baseImage + imgDigest
		fullReference = baseImage + imgTag + imgDigest
		invalidImage  = "*&"
	)

	Convey("Test_imageRemoveSuffixDigest", t, func() {
		Convey("with tag only", func() {
			So(imageRemoveSuffixDigest(withTag), ShouldEqual, withTag)
		})

		Convey("with digest only", func() {
			So(imageRemoveSuffixDigest(withDigest), ShouldEqual, baseImage)
		})

		Convey("full reference", func() {
			So(imageRemoveSuffixDigest(fullReference), ShouldEqual, withTag)
		})

		Convey("invalid reference", func() {
			So(imageRemoveSuffixDigest(invalidImage), ShouldBeBlank)
		})
	})
}

func Test_reconcileHostConfig(t *testing.T) {
	Convey("Test_reconcileHostConfig", t, func() {
		runtime := "lcr"
		h := &types.IsuladHostConfig{
			RestartPolicy: &types.RestartPolicy{
				Name:              "unless-stopped",
				MaximumRetryCount: 0,
			},
			UsernsMode: "host",
		}
		reconcileHostConfig(h, runtime)
		So(h.Runtime, ShouldEqual, runtime)
		So(h.RestartPolicy.Name, ShouldEqual, "always")
		So(h.UsernsMode, ShouldEqual, "")
	})
}

func Test_reconcileOciConfig(t *testing.T) {
	Convey("Test_reconcileOciConfig", t, func() {
		var deviceNumber [][]int64 = [][]int64{{5, 2}, {136, -1}}

		baseSpec := &specs.Spec{
			Annotations: make(map[string]string),
			Root: &specs.Root{
				Path: "/old/path/of/base/rootfs",
			},
			Linux: &specs.Linux{
				Namespaces: []specs.LinuxNamespace{
					{Type: "pid"},
					{Type: "network"},
					{Type: "ipc"},
					{Type: "uts"},
				},
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Allow: true, Type: "c", Major: &deviceNumber[0][0], Minor: &deviceNumber[0][1], Access: "rwm"},
					},
				},
				CgroupsPath: "/docker/" + reconcileTestCtrID,
				Devices: []specs.LinuxDevice{
					{Path: "/dev/test/0:0:0:0"},
					{Path: "/dev/null"},
				},
			},
			Mounts: []specs.Mount{
				{Destination: "/etc/hostname", Source: "/old/path/of/hostname"},
				{Destination: "/etc/hosts", Source: "/old/path/of/hosts"},
				{Destination: "/etc/resolv.conf", Source: "/old/path/of/resolv.conf"},
				{Destination: "/dev/shm", Source: "/old/path/of/mounts/shm"},
				{Destination: "/data", Source: "/data"},
			},
		}

		baseCommonCfg := &types.CommonConfig{
			ID:             reconcileTestCtrID,
			BaseFs:         "/new/path/of/base/rootfs",
			HostnamePath:   "/new/path/of/hostname",
			HostsPath:      "/new/path/of/hosts",
			ResolvConfPath: "/new/path/of/resolv.conf",
			ShmPath:        "/new/path/of/mounts/shm",
			Config: &types.ContainerCfg{
				Annotations: map[string]string{"testAnnotations": "wuhan is a heroical city"},
			},
		}

		expectSpec := &specs.Spec{
			Annotations: map[string]string{
				"cgroup.dir":      "/isulad",
				"testAnnotations": "wuhan is a heroical city",
			},
			Root: &specs.Root{
				Path: "/new/path/of/base/rootfs",
			},
			Linux: &specs.Linux{
				Namespaces: []specs.LinuxNamespace{
					{Type: "pid", Path: reconcileTestConnectCtrID},
					{Type: "network", Path: ""},
					{Type: "ipc", Path: ""},
					{Type: "uts", Path: ""},
				},
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Allow: true, Type: "c", Major: &deviceNumber[0][0], Minor: &deviceNumber[0][1], Access: "rwm"},
						{Allow: true, Type: "c", Major: &deviceNumber[1][0], Minor: &deviceNumber[1][1], Access: "rwm"},
					},
				},
				CgroupsPath: "/isulad/" + reconcileTestCtrID,
				Devices: []specs.LinuxDevice{
					{Path: "/dev/null"},
				},
			},
			Mounts: []specs.Mount{
				{Destination: "/etc/hostname", Source: "/new/path/of/hostname"},
				{Destination: "/etc/hosts", Source: "/new/path/of/hosts"},
				{Destination: "/etc/resolv.conf", Source: "/new/path/of/resolv.conf"},
				{Destination: "/dev/shm", Source: "/new/path/of/mounts/shm", Options: []string{"mode=1777", "size=0"}},
				{Destination: "/data", Source: "/data"},
			},
		}
		reconcileOciConfig(baseSpec, baseCommonCfg, &types.IsuladHostConfig{
			PidMode:     "container:" + reconcileTestConnectCtrID,
			NetworkMode: "container:lessen64",
			IpcMode:     "notcontainer:lessen64",
		})
		So(Diff(baseSpec, expectSpec), ShouldBeBlank)
	})
}

func Test_genV2OptsFromHostCfg(t *testing.T) {
	Convey("Test_reconcileOciConfig", t, func() {
		Convey("nil hostConfig", func() {
			opts := genV2OptsFromHostCfg(nil)
			So(opts, ShouldBeNil)
		})

		Convey("test the functionality of generation opts", func() {
			v2 := &types.IsuladV2Config{
				CommonConfig: &types.CommonConfig{
					Config: &types.ContainerCfg{
						Annotations: make(map[string]string),
					},
				},
			}

			Convey("OomScoreAdj", func() {
				h := &types.IsuladHostConfig{
					OomScoreAdj: 1000,
				}
				opts := genV2OptsFromHostCfg(h)
				for _, o := range opts {
					o(v2)
				}
				oomScore, exist := v2.CommonConfig.Config.Annotations["proc.oom_score_adj"]
				So(exist, ShouldBeTrue)
				So(oomScore, ShouldEqual, strconv.Itoa(h.OomScoreAdj))
			})

			Convey("files limit", func() {
				h := &types.IsuladHostConfig{
					FilesLimit: 1000,
				}
				opts := genV2OptsFromHostCfg(h)
				for _, o := range opts {
					o(v2)
				}
				oomScore, exist := v2.CommonConfig.Config.Annotations["files.limit"]
				So(exist, ShouldBeTrue)
				So(oomScore, ShouldEqual, strconv.FormatInt(h.FilesLimit, 10))
			})
		})
	})
}
