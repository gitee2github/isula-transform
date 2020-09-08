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

package isulad

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/sys/unix"
	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
)

const itTestCtrID = "isulatransformittestctr"

var testIsuladTool = &Tool{
	graph:       "/var/lib/isulad",
	runtime:     "lcr",
	storageType: transform.Overlay2,
}

func TestInitIsuladTool(t *testing.T) {
	Convey("TestInitIsuladTool", t, func() {
		Convey("wrong container runtime", func() {
			err := InitIsuladTool(&DaemonConfig{
				Runtime: "kata",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not support runtime")
		})

		Convey("wrong storage driver", func() {
			err := InitIsuladTool(&DaemonConfig{
				StorageDriver: "aufs",
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not support storage driver")
		})

		Convey("default init", func() {
			So(InitIsuladTool(&DaemonConfig{}), ShouldBeNil)
			So(GetIsuladTool().graph, ShouldEqual, testIsuladTool.graph)
			So(GetIsuladTool().runtime, ShouldEqual, testIsuladTool.runtime)
			So(GetIsuladTool().storageType, ShouldEqual, testIsuladTool.storageType)
			So(GetIsuladTool().storageDriver, ShouldNotBeNil)
		})
	})
}

func TestIsuladTool_GetterFunc(t *testing.T) {
	Convey("TestIsuladTool_GetterFunc", t, func() {
		Convey("StorageType", func() {
			So(testIsuladTool.StorageType(), ShouldEqual, transform.Overlay2)
		})

		Convey("BaseStorageDriver", func() {
			So(testIsuladTool.BaseStorageDriver(), ShouldBeNil)
		})

		Convey("Runtime", func() {
			So(testIsuladTool.Runtime(), ShouldEqual, defaultRuntime)
		})
	})
}

func TestIsuladTool_GetPathFunc(t *testing.T) {
	Convey("TestIsuladTool_GetPathFunc", t, func() {
		Convey("GetRuntimePath", func() {
			want := "/var/lib/isulad/engines/lcr"
			So(testIsuladTool.GetRuntimePath(), ShouldEqual, want)
		})

		Convey("GetHostCfgPath", func() {
			want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.Hostconfig)
			So(testIsuladTool.GetHostCfgPath(itTestCtrID), ShouldEqual, want)
		})

		Convey("GetConfigV2Path", func() {
			want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.V2config)
			So(testIsuladTool.GetConfigV2Path(itTestCtrID), ShouldEqual, want)
		})

		Convey("GetOciConfigPath", func() {
			want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.Ociconfig)
			So(testIsuladTool.GetOciConfigPath(itTestCtrID), ShouldEqual, want)
		})

		Convey("GetNetworkFilePath", func() {
			Convey("GetHostnamePath", func() {
				want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.Hostname)
				So(testIsuladTool.GetNetworkFilePath(itTestCtrID, types.Hostname), ShouldEqual, want)
			})

			Convey("GetHostsPath", func() {
				want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.Hosts)
				So(testIsuladTool.GetNetworkFilePath(itTestCtrID, types.Hosts), ShouldEqual, want)
			})

			Convey("GetResolvPath", func() {
				want := filepath.Join("/var/lib/isulad/engines/lcr", itTestCtrID, types.Resolv)
				So(testIsuladTool.GetNetworkFilePath(itTestCtrID, types.Resolv), ShouldEqual, want)
			})
		})
	})
}

func TestIsuladTool_PrepareRootDir(t *testing.T) {
	path := "/var/lib/isulad/engines/lcr/" + itTestCtrID
	if err := os.RemoveAll(path); err != nil {
		t.Skipf("before remove root dir: %v", err)
	}
	Convey("TestIsuladTool_PrepareRootDir", t, func() {
		Convey("dir already exist", func() {
			err := testIsuladTool.PrepareBundleDir(itTestCtrID)
			defer os.RemoveAll(path)
			So(err, ShouldBeNil)
			err = testIsuladTool.PrepareBundleDir(itTestCtrID)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "already exists")
		})

		Convey("normal", func() {
			err := testIsuladTool.PrepareBundleDir(itTestCtrID)
			So(err, ShouldBeNil)
			defer os.RemoveAll(path)
			info, err := os.Stat(path)
			So(err, ShouldBeNil)
			So(info.IsDir(), ShouldBeTrue)
			So(info.Mode(), ShouldEqual, rootDirMode|os.ModeDir)
		})
	})
}

func TestIsuladTool_SaveConfig(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "IsuladTool")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	Convey("TestIsuladTool_SaveConfig", t, func() {
		Convey("MarshalIndent", func() {
			Convey("read normal", func() {
				src := &struct {
					Usage  string
					Indent struct{ Usage string }
				}{
					Usage:  "test for IsuladTool.SaveConfig read normal",
					Indent: struct{ Usage string }{Usage: "test for IsuladTool.MarshalIndent layers"},
				}
				getPath := func(string) string {
					return filepath.Join(tmpdir, "test.json")
				}
				So(testIsuladTool.SaveConfig(itTestCtrID, src, testIsuladTool.MarshalIndent, getPath), ShouldBeNil)
				got, _ := ioutil.ReadFile(getPath(itTestCtrID))
				want := `{
	"Usage": "test for IsuladTool.SaveConfig read normal",
	"Indent": {
		"Usage": "test for IsuladTool.MarshalIndent layers"
	}
}`
				So(string(got), ShouldEqual, want)
				info, _ := os.Stat(getPath(itTestCtrID))
				So(info.Mode(), ShouldEqual, cfgFileMode)
			})

			Convey("read abnormal", func() {
				src := &struct{ C chan int }{C: make(chan int)}
				getPath := func(string) string { return "" }
				So(testIsuladTool.SaveConfig(itTestCtrID, src, testIsuladTool.MarshalIndent, getPath), ShouldBeError)
			})
		})

		Convey("network file mode", func() {
			read := func(src interface{}) ([]byte, error) {
				return []byte("localhost"), nil
			}
			getPath := func(string) string {
				return filepath.Join(tmpdir, "hosts")
			}
			So(testIsuladTool.SaveConfig(itTestCtrID, nil, read, getPath), ShouldBeNil)
			got, _ := ioutil.ReadFile(getPath(itTestCtrID))
			want := "localhost"
			So(string(got), ShouldEqual, want)
			info, _ := os.Stat(getPath(itTestCtrID))
			So(info.Mode(), ShouldEqual, networkFileMode)
		})
	})
}

func TestIsuladTool_Cleanup(t *testing.T) {
	path := "/var/lib/isulad/engines/lcr/" + itTestCtrID
	if err := os.MkdirAll(path, rootDirMode); err != nil {
		t.Skipf("make root dir: %v", err)
	}
	Convey("TestIsuladTool_Cleanup", t, func() {
		err := testIsuladTool.Cleanup(itTestCtrID)
		So(err, ShouldBeNil)
		info, err := os.Stat(path)
		So(info, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "no such file or directory")
	})
}

func TestIsuladTool_PrepareShm(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "IsuladTool")
	if err != nil {
		t.Skipf("make temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)
	Convey("TestIsuladTool_PrepareShm", t, func() {
		var shmPath = filepath.Join(tmpdir, "mounts/shm")
		var shmSize int64 = 67108864
		So(testIsuladTool.PrepareShm(shmPath, shmSize), ShouldBeEmpty)
		defer func(path string) {
			if err := unix.Unmount(path, unix.MNT_DETACH); err != nil {
				t.Logf("umount path err: %v", err)
			}
		}(shmPath)
		info, err := os.Stat(shmPath)
		So(err, ShouldBeNil)
		So(info.Mode(), ShouldEqual, os.ModeDir|os.ModeSticky|os.ModePerm)
	})
}
