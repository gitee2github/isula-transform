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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	. "github.com/golang/mock/gomock"
	. "github.com/google/go-cmp/cmp"
	"github.com/opencontainers/runtime-spec/specs-go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli"
	"isula.org/isula-transform/pkg/isulad"
	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
)

const (
	transformTestCtrID = "511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2"
	notExistCtrID      = "notexist"
	incorrectCtrID     = "incorrectformat"
	incorrectFile      = "incorrect.json"
)

func TestNew(t *testing.T) {
	applyFlags := func(s *flag.FlagSet, flags ...cli.StringFlag) {
		for _, f := range flags {
			f.Apply(s)
		}
	}
	graphFlag := cli.StringFlag{Name: "docker-graph"}
	stateFlag := cli.StringFlag{Name: "docker-state"}

	Convey("TestNew", t, func() {
		Convey("default config", func() {
			flags := flag.NewFlagSet("", flag.ContinueOnError)
			applyFlags(flags, graphFlag, stateFlag)
			ctx := cli.NewContext(nil, flags, nil)
			got := New(ctx)
			expect := &dockerTransformer{
				BaseTransformer: transform.BaseTransformer{
					Name:      "docker",
					StateRoot: "/var/run/docker",
					GraphRoot: "/var/lib/docker",
				},
			}
			So(reflect.DeepEqual(got, expect), ShouldBeTrue)
		})

		Convey("user defined", func() {
			graphFlag.Value = "/test/lib/docker"
			stateFlag.Value = "/test/run/docker"
			flags := flag.NewFlagSet("", flag.ContinueOnError)
			applyFlags(flags, graphFlag, stateFlag)
			ctx := cli.NewContext(nil, flags, nil)
			got := New(ctx)
			expect := &dockerTransformer{
				BaseTransformer: transform.BaseTransformer{
					Name:      "docker",
					StateRoot: "/test/run/docker",
					GraphRoot: "/test/lib/docker",
				},
			}
			So(reflect.DeepEqual(got, expect), ShouldBeTrue)
		})
	})
}

func Test_dockerConfigEngine_transformHostConfig(t *testing.T) {
	// init
	hostCfgFile := types.Hostconfig
	tmpdir, err := ioutil.TempDir("", "isula-transform")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	if err := initDockerTransformTest(tmpdir, transformTestCtrID, hostCfgFile, hostCfgFile, true); err != nil {
		t.Skipf("init test transform HostConfig failed: %v", err)
	}
	dt := getTestDockerTransformer(tmpdir)

	Convey("Test_dockerConfigEngine_transformHostConfig", t, func() {
		Convey("container not exist", func() {
			_, _, err := dt.transformHostConfig(notExistCtrID)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})

		Convey("incorrect json format", func() {
			if err := initDockerTransformTest(tmpdir, incorrectCtrID,
				incorrectFile, hostCfgFile, false); err != nil {
				t.Skipf("prepare test incorrect json format failed: %v", err)
			}
			_, _, err := dt.transformHostConfig(incorrectCtrID)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})

		Convey("transform successfully", func() {
			hGot, lGot, err := dt.transformHostConfig(transformTestCtrID)
			hExpect := &types.IsuladHostConfig{
				NetworkMode: "host",
				Runtime:     "lcr",
				IpcMode:     "shareable",
				RestartPolicy: &types.RestartPolicy{
					Name:              "always",
					MaximumRetryCount: 0,
				},
				ShmSize: 67108864,
			}
			lExpect := &container.LogConfig{
				Type: "json-file",
				Config: map[string]string{
					"max-size": "30KB",
				},
			}
			So(err, ShouldBeNil)
			So(Diff(hExpect, hGot), ShouldBeBlank)
			So(Diff(lExpect, lGot), ShouldBeBlank)
		})
	})
}

//go:generate mockgen -destination mock_storage.go -package docker isula.org/isula-transform/transform StorageDriver
func Test_dockerConfigEngine_transformV2Config(t *testing.T) {
	// init
	configV2File := "config.v2.json"
	tmpdir, err := ioutil.TempDir("", "isula-transform")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	if err := initDockerTransformTest(tmpdir, transformTestCtrID, configV2File, configV2File, true); err != nil {
		t.Skipf("init test transform HostConfig failed: %v", err)
	}
	dt := getTestDockerTransformer(tmpdir)
	// mock StorageDriver
	ctrl := NewController(t)
	defer ctrl.Finish()
	sd := NewMockStorageDriver(ctrl)
	InOrder(
		sd.EXPECT().GenerateRootFs(Any(), Any()).Return("", fmt.Errorf("mock generate failed")),
		sd.EXPECT().GenerateRootFs(Any(), Any()).Return("newRootFS", nil),
	)
	dt.sd = sd

	Convey("Test_dockerConfigEngine_transformV2Config", t, func() {
		Convey("container not exist", func() {
			_, err := dt.transformV2Config(notExistCtrID)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})

		Convey("incorrect json format", func() {
			if err := initDockerTransformTest(tmpdir, incorrectCtrID, incorrectFile, configV2File, false); err != nil {
				t.Skipf("prepare test incorrect json format failed: %v", err)
			}
			_, err := dt.transformV2Config(incorrectCtrID)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})

		Convey("load successfully", func() {
			Convey("generate rootfs failed", func() {
				_, err := dt.transformV2Config(transformTestCtrID)
				So(err.Error(), ShouldContainSubstring, "mock generate failed")
			})

			Convey("transform successfully", func() {
				opts := []v2ConfigReconcileOpt{
					v2ConfigWithLogConfig(nil, ""),
				}
				got, err := dt.transformV2Config(transformTestCtrID, opts...)
				expect := &types.IsuladV2Config{
					CommonConfig: &types.CommonConfig{
						Path: "bash",
						Args: []string{},
						Config: &types.ContainerCfg{
							Tty:       true,
							OpenStdin: true,
							Hostname:  "localhost.localdomain",
							Env: []string{
								"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
							},
							Cmd: []string{
								"bash",
							},
							Labels: make(map[string]string),
							Image:  "isulatransformtestcontainer:image",
							Annotations: map[string]string{
								"cgroup.dir":             "/isulad",
								"log.console.driver":     "json-file",
								"log.console.file":       "none",
								"log.console.filerotate": "7",
								"log.console.filesize":   "30KB",
								"native.umask":           "secure",
								"rootfs.mount":           "/var/lib/isulad/mnt/rootfs",
							},
						},
						Created:                time.Unix(1579744800, 000000000).Local(),
						Image:                  "isulatransformtestcontainer:image",
						ImageType:              "oci",
						HostnamePath:           tmpdir + "/lib/isulad/engines/lcr/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/hostname",
						HostsPath:              tmpdir + "/lib/isulad/engines/lcr/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/hosts",
						ResolvConfPath:         tmpdir + "/lib/isulad/engines/lcr/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/resolv.conf",
						OriginHostnamePath:     "/var/lib/docker/containers/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/hostname",
						OriginHostsPath:        "/var/lib/docker/containers/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/hosts",
						OriginResolvConfPath:   "/var/lib/docker/containers/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/resolv.conf",
						ShmPath:                tmpdir + "/lib/isulad/engines/lcr/511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2/mounts/shm",
						LogPath:                "none",
						LogDriver:              "json-file",
						BaseFs:                 "newRootFS",
						MountPoints:            make(map[string]types.Mount),
						Name:                   "isulatransformtestcontainer",
						RestartCount:           0,
						ID:                     "511e7f915e3f5dc09b36a49657125eea4b36a05f862ab3dd01e0b9b2",
						HasBeenManuallyStopped: false,
						HasBeenStartedBefore:   true,
					},
					Image: "sha256:dc6b3e3cf28225d72351d5dbddc35ea08a08ad83725043903df61448c9e466a0",
					State: &types.ContainerState{
						Running:   false,
						Pid:       17373,
						StartedAt: time.Unix(1579744800, 000000000).Local(),
					},
				}
				So(err, ShouldBeNil)
				// types.ContainerState.FinishedAt use time.Now().Local()
				// need to sync from got
				expect.State.FinishedAt = got.State.FinishedAt
				So(Diff(expect, got), ShouldBeBlank)
			})
		})
	})
}

func Test_dockerConfigEngine_transformOciConfig(t *testing.T) {
	ociCfgFile := "config.json"
	tmpdir, err := ioutil.TempDir("", "isula-transform")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	if err := initDockerTransformTest(tmpdir, transformTestCtrID, ociCfgFile, ociCfgFile, true); err != nil {
		t.Skipf("init test transform HostConfig failed: %v", err)
	}
	dt := getTestDockerTransformer(tmpdir)

	Convey("Test_dockerConfigEngine_transformOciConfig", t, func() {
		Convey("container not exist", func() {
			_, _, err := dt.transformOciConfig(notExistCtrID, nil, nil)
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
		})

		Convey("incorrect json format", func() {
			if err := initDockerTransformTest(tmpdir, incorrectCtrID, incorrectFile, ociCfgFile, false); err != nil {
				t.Skipf("prepare test incorrect json format failed: %v", err)
			}
			_, _, err := dt.transformOciConfig(incorrectCtrID, nil, nil)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})

		Convey("transform successfully", func() {
			var (
				basePath   = isulad.GetIsuladTool().GetRuntimePath() + transformTestCtrID
				hostname   = basePath + "/hostname"
				hosts      = basePath + "/hosts"
				resolvconf = basePath + "/resolv.conf"
				shmPath    = basePath + "/mounts/shm"
				devicesNum = [][]int64{{5, 2}, {136, -1}}
				oldRootFS  = "/old/root/fs"

				common = &types.CommonConfig{
					BaseFs: "/new/root/fs",
					Config: &types.ContainerCfg{
						Annotations: map[string]string{
							"testOci": "testOci",
						},
					},
					HostnamePath:   hostname,
					ResolvConfPath: resolvconf,
					HostsPath:      hosts,
					ShmPath:        shmPath,
					ID:             transformTestCtrID,
				}

				host = &types.IsuladHostConfig{
					ShmSize: 123456,
				}
			)
			got, gotOldRootFS, err := dt.transformOciConfig(transformTestCtrID, common, host)
			expect := &specs.Spec{
				Version: "1.0.1-dev",
				Process: &specs.Process{
					Terminal: true,
					User: specs.User{
						UID: 0,
						GID: 0,
					},
					Args: []string{"bash"},
					Env: []string{
						"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
						"HOSTNAME=localhost.localdomain",
						"TERM=xterm",
					},
					Cwd: "/",
				},
				Root: &specs.Root{
					Path: "/new/root/fs",
				},
				Hostname: "localhost.localdomain",
				Mounts: []specs.Mount{
					{Destination: "/etc/resolv.conf", Type: "bind", Source: resolvconf, Options: []string{"rbind", "rprivate"}},
					{Destination: "/etc/hostname", Type: "bind", Source: hostname, Options: []string{"rbind", "rprivate"}},
					{Destination: "/etc/hosts", Type: "bind", Source: hosts, Options: []string{"rbind", "rprivate"}},
					{Destination: "/dev/shm", Type: "bind", Source: shmPath, Options: []string{"rbind", "rprivate", "mode=1777", "size=123456"}},
				},
				Hooks: &specs.Hooks{},
				Annotations: map[string]string{
					"cgroup.dir":   "/isulad",
					"native.umask": "secure",
					"testOci":      "testOci",
				},
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{
							{Allow: false, Access: "rwm"},
							{Allow: true, Type: "c", Major: &devicesNum[0][0], Minor: &devicesNum[0][1], Access: "rwm"},
							{Allow: true, Type: "c", Major: &devicesNum[1][0], Minor: &devicesNum[1][1], Access: "rwm"},
						},
					},
					CgroupsPath: "/isulad/" + transformTestCtrID,
					Namespaces: []specs.LinuxNamespace{
						{Type: "network"},
					},
				},
			}
			So(err, ShouldBeNil)
			So(gotOldRootFS, ShouldEqual, oldRootFS)
			So(Diff(expect, got), ShouldBeBlank)
		})
	})
}

func Test_dockerTransformer_initStorageDriver(t *testing.T) {
	dt := &dockerTransformer{}
	Convey("Test_dockerTransformer_initStorageDriver", t, func() {
		Convey("overlay2 driver", func() {
			err := isulad.InitIsuladTool("", "", "overlay2", "")
			So(err, ShouldBeNil)
			ol, err := dt.initStorageDriver()
			So(err, ShouldBeNil)
			_, ok := ol.(*overlayDriver)
			So(ok, ShouldBeTrue)
		})

		Convey("devicemapper driver", func() {
			err := isulad.InitIsuladTool("", "", "devicemapper", "")
			So(err, ShouldBeNil)
			ol, err := dt.initStorageDriver()
			So(err, ShouldBeNil)
			_, ok := ol.(*deviceMapperDriver)
			So(ok, ShouldBeTrue)
		})
	})
}

func Test_dockerTransformer_initContainers(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "isula-transform")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	dt := getTestDockerTransformer(tmpdir)
	err = os.MkdirAll(dt.StateRoot+"/containerd/daemon/"+containerdRuntime, 0700)
	if err != nil {
		t.Skipf("make docker graph dir: %v", err)
	}

	Convey("Test_dockerTransformer_initContainers", t, func() {
		ctrsRoot := filepath.Join(dt.StateRoot, "containerd/daemon", containerdRuntime, containerdNameSpace)

		Convey("containers is not directory", func() {
			_, err = os.OpenFile(ctrsRoot, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
			if err != nil {
				t.Skipf("create graph root contains: %v", err)
			}
			defer os.RemoveAll(ctrsRoot)
			initErr := dt.initContainers()
			So(initErr, ShouldBeError)
			So(initErr.Error(), ShouldContainSubstring, "init docker container store failed")
		})

		Convey("containers contains item invalid", func() {
			err = os.MkdirAll(ctrsRoot, 0700)
			if err != nil {
				t.Skipf("make docker containers dir: %v", err)
			}
			defer os.RemoveAll(ctrsRoot)
			ctrs := []string{
				"ebe35e6089fa868bfde477bb2cc749a88b1e93dc66fa199b0ed6927f10f86b5a",
				"d77fdb420e801f1e9cf081b5bd5f2948047e5a7a790d12a81edd8f4c0f4fef4d",
				"lessthen64",
			}
			for _, ctr := range ctrs {
				err := os.MkdirAll(ctrsRoot+"/"+ctr, 0700)
				if err != nil {
					t.Skipf("make running container %s dir: %v", ctr, err)
				}
			}
			fileCtr := ctrsRoot + "/" + "d32d678c960aeb50d55107ed1aaf0c82957a88c431d422657a5df6a6920f0446"
			if _, err := os.OpenFile(fileCtr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640); err != nil {
				t.Skipf("create graph root contains: %v", err)
			}
			initErr := dt.initContainers()
			So(initErr, ShouldBeNil)
			var loadCtrs []string
			dt.ctrs.Range(func(key, value interface{}) bool {
				ctr, _ := key.(string)
				loadCtrs = append(loadCtrs, ctr)
				return true
			})
			expectCtrs := ctrs[0:2]
			sort.StringSlice(loadCtrs).Sort()
			sort.StringSlice(expectCtrs).Sort()
			So(Diff(loadCtrs, expectCtrs), ShouldBeBlank)
		})
	})
}

func Test_dockerTransformer_matchID(t *testing.T) {
	Convey("Test_dockerTransformer_matchID", t, func() {
		dt := &dockerTransformer{
			ctrs: &sync.Map{},
		}
		testCtrs := []string{
			"ebe35e6089fa868bfde477bb2cc749a88b1e93dc66fa199b0ed6927f10f86b5a",
			"d77fdb420e801f1e9cf081b5bd5f2948047e5a7a790d12a81edd8f4c0f4fef4d",
			"8af0fdebdceb85143394db47d145a3161038ab8b723583fd5f3c8d39470a3017",
		}
		for _, ctr := range testCtrs {
			dt.ctrs.Store(ctr, false)
		}

		Convey("not exist", func() {
			id, st := dt.matchID("notexist")
			So(id, ShouldBeBlank)
			So(st, ShouldEqual, notExist)
		})

		Convey("update transform status", func() {
			prefix := "d77"
			id, st := dt.matchID(prefix)
			So(id, ShouldEqual, "d77fdb420e801f1e9cf081b5bd5f2948047e5a7a790d12a81edd8f4c0f4fef4d")
			So(st, ShouldEqual, needTransform)
			id, st = dt.matchID(prefix)
			So(id, ShouldEqual, "d77fdb420e801f1e9cf081b5bd5f2948047e5a7a790d12a81edd8f4c0f4fef4d")
			So(st, ShouldEqual, hasBeenTransformed)
		})
	})
}

func loadTestData(name string) ([]byte, error) {
	return ioutil.ReadFile("./testdata/" + name)
}

func initDockerTransformTest(repalceVar, id, src, dest string, initIsuladTool bool) error {
	var fileRoot string

	switch dest {
	case types.V2config, types.Hostconfig:
		fileRoot = repalceVar + "/lib/docker/containers/" + id
	case types.Ociconfig:
		fileRoot = repalceVar + "/run/docker/containerd/daemon/io.containerd.runtime.v1.linux/moby/" + id
	}
	if err := os.MkdirAll(fileRoot, 0700); err != nil {
		return err
	}
	testData, err := loadTestData(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fileRoot+"/"+dest, testData, 0640); err != nil {
		return err
	}
	if initIsuladTool {
		graph := filepath.Join(repalceVar + "/lib/isulad")
		_ = isulad.InitIsuladTool(graph, "", "", "")
		return isulad.GetIsuladTool().PrepareBundleDir(id)
	}
	return nil
}

func getTestDockerTransformer(tmpdir string) *dockerTransformer {
	return &dockerTransformer{
		BaseTransformer: transform.BaseTransformer{
			Name:      "docker",
			StateRoot: tmpdir + "/run/docker",
			GraphRoot: tmpdir + "/lib/docker",
		},
	}
}
