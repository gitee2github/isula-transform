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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
)

const (
	rootDirMode     os.FileMode = 0750
	mountsDirMode   os.FileMode = 0700
	cfgFileMode     os.FileMode = 0640
	networkFileMode os.FileMode = 0644

	initMask int = 0022

	defaultISuladGraphPath = "/var/lib/isulad"
	defaultRuntime         = "lcr"
	defaultStorageDriver   = "overlay2"
)

var (
	commonTool *IsuladTool
)

// IsuladTool contains the common functions used by transformer
type IsuladTool struct {
	graphRoot string
	runtime   string

	// storage
	storageType   transform.StorageType
	storageDriver transform.BaseStorageDriver
}

func init() {
	syscall.Umask(initMask)
}

// InitIsuladTool initializes the global iSuladCfgTool with the given parameters
func InitIsuladTool(graphRoot, runtime, storageDriver, imageSrvAddr string) error {
	if graphRoot == "" {
		graphRoot = defaultISuladGraphPath
	}
	if runtime == "" {
		runtime = defaultRuntime
	}
	if storageDriver == "" {
		storageDriver = defaultStorageDriver
	}
	commonTool = &IsuladTool{
		graphRoot:   graphRoot,
		runtime:     runtime,
		storageType: transform.StorageType(storageDriver),
	}

	if err := checkToolConfigValid(); err != nil {
		logrus.Errorf("config of iSuladTool is invalid: %+v", commonTool)
		return errors.Wrap(err, "config of iSuladTool is invalid")
	}

	if err := initBaseStorageDriver(imageSrvAddr); err != nil {
		logrus.Errorf("init global base storage driver failed: %v", err)
		return errors.Wrap(err, "init global base storage driver failed")
	}

	commonTool.storageDriver = globalIsuladStorageDriver

	return nil
}

// GetIsuladCfgTool returns the global isuladtool
func GetIsuladTool() *IsuladTool {
	return commonTool
}

func checkToolConfigValid() error {
	g := GetIsuladTool()
	// runtime
	switch g.runtime {
	case defaultRuntime:
	default:
		return fmt.Errorf("not support runtime: %s", g.runtime)
	}
	// storage driver
	switch g.storageType {
	case transform.Overlay2, transform.DeviceMapper:
	default:
		return fmt.Errorf("not support storage driver: %s", g.runtime)
	}
	return nil
}

// StorageType returns the storage type of isulad
func (ict *IsuladTool) StorageType() transform.StorageType {
	return ict.storageType
}

// BaseStorageDriver returns the global base storage driver tool
func (ict *IsuladTool) BaseStorageDriver() transform.BaseStorageDriver {
	return ict.storageDriver
}

// Runtime returns the runtime of isulad used
func (ict *IsuladTool) Runtime() string {
	return ict.runtime
}

// GetRuntimePath returns the default runtime path of isulad
func (ict *IsuladTool) GetRuntimePath() string {
	return filepath.Join(ict.graphRoot, "engines", ict.runtime)
}

// PrepareBundleDir creates runtime root dir of the container
func (ict *IsuladTool) PrepareBundleDir(id string) error {
	path := filepath.Join(ict.GetRuntimePath(), id)
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return fmt.Errorf("directory %s already exists, container has been or is being transformed", path)
	}
	return os.MkdirAll(path, rootDirMode)
}

// GetHostCfgPath returns path of hostconfig.json
func (ict *IsuladTool) GetHostCfgPath(id string) string {
	return filepath.Join(ict.GetRuntimePath(), id, types.Hostconfig)
}

// GetConfigV2Path returns path of config.v2.json
func (ict *IsuladTool) GetConfigV2Path(id string) string {
	return filepath.Join(ict.GetRuntimePath(), id, types.V2config)
}

// GetOciConfigPath returns path of config.json
func (ict *IsuladTool) GetOciConfigPath(id string) string {
	return filepath.Join(ict.GetRuntimePath(), id, types.Ociconfig)
}

// GetNetworkFilePath returns the path specified file in host, hostname and resolv.conf
func (ict *IsuladTool) GetNetworkFilePath(id, file string) string {
	return filepath.Join(ict.GetRuntimePath(), id, file)
}

// ReadData allows isuladTool to read data from a source
type ReadData func(src interface{}) ([]byte, error)

// FilePath allows isuladTool to obtain the path to the file to be written and saved
type FilePath func(string) string

// SaveConfig allows isuladTool to save data to file
func (ict *IsuladTool) SaveConfig(id string, src interface{}, read ReadData, getPath FilePath) error {
	path := getPath(id)

	_, err := os.Stat(path)
	// getPath should not be exist here
	if err == nil || os.IsExist(err) {
		return errors.Errorf("%s already exist", path)
	}

	data, err := read(src)
	if err != nil {
		logrus.Errorf("save config read data internal error: %v", err)
		return err
	}
	var mode os.FileMode
	switch filepath.Base(path) {
	case types.Hostname, types.Hosts, types.Resolv:
		mode = networkFileMode
	default:
		mode = cfgFileMode
	}
	err = ioutil.WriteFile(path, data, mode)
	if err != nil {
		logrus.Errorf("write data(%s) to file %s failed: %v", string(data), path, err)
		return err
	}
	return nil
}

// MarshalIndent formats the json bytes with indent
func (ict *IsuladTool) MarshalIndent(src interface{}) (bytes []byte, e error) {
	return json.MarshalIndent(src, "", "\t")
}

// Cleanup remove runtime root dir of the container
func (ict *IsuladTool) Cleanup(id string) error {
	path := filepath.Join(ict.GetRuntimePath(), id)
	return os.RemoveAll(path)
}

// PrepareShm creates sharm shm mount point for container
func (ict *IsuladTool) PrepareShm(path string, size int64) error {
	err := os.MkdirAll(path, mountsDirMode)
	if err != nil {
		return err
	}
	shmProperty := "mode=1777,size=" + strconv.FormatInt(size, 10)
	err = unix.Mount("shm", path, "tmpfs", uintptr(unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_NODEV), shmProperty)
	if err != nil {
		return err
	}
	return nil
}

// LcrCreate calls lcr interface to init isulad container
func (ict *IsuladTool) LcrCreate(id string, spec []byte) error {
	return lcrCreate(id, ict.GetRuntimePath(), spec)
}
