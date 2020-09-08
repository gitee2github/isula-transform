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

// Package docker implement transformer for transform docker container
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
	"isula.org/isula-transform/pkg/isulad"
	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
	"isula.org/isula-transform/utils"
)

type containerStatus int

const (
	notExist containerStatus = iota
	hasBeenTransformed
	needTransform
)

var (
	defaultDockerHostProto = "unix"
	defaultDockerHostAddr  = "/var/run/docker.sock"
	defaultExecRoot        = "/var/run/docker" // Root directory for execution state files
	defaultDataRoot        = "/var/lib/docker" // Root directory of the Docker runtime
	containerdRuntime      = "io.containerd.runtime.v1.linux"
	containerdNameSpace    = "moby"

	defaultTimeout = 10 * time.Second
	containerIDLen = 64
)

type dockerClient interface {
	ContainerDiff(context.Context, string) ([]container.ContainerChangeResponseItem, error)
	ContainerPause(context.Context, string) error
}

type dockerTransformer struct {
	ctrs   *sync.Map
	client dockerClient
	sd     transform.StorageDriver
	transform.BaseTransformer
}

func init() {
	transform.Register("docker", New)
}

// New return a transform engine for docker container
func New(ctx *cli.Context) transform.Transformer {
	var opts []transform.EngineOpt
	graphRoot := ctx.GlobalString("docker-graph")
	stateRoot := ctx.GlobalString("docker-state")
	opts = append(opts, transform.EngineWithGraph(graphRoot), transform.EngineWithState(stateRoot))
	return newWithConfig(opts...)
}

// newWithConfig create a transform engine for docker container with specific config
func newWithConfig(opts ...transform.EngineOpt) transform.Transformer {
	var e dockerTransformer
	for _, o := range opts {
		o(&e.BaseTransformer)
	}
	if e.StateRoot == "" {
		e.StateRoot = defaultExecRoot
	}
	if e.GraphRoot == "" {
		e.GraphRoot = defaultDataRoot
	}
	e.Name = "docker"
	return &e
}

func (t *dockerTransformer) Init() error {
	var retErr error
	c := &http.Client{
		Timeout: 2 * defaultTimeout,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout(defaultDockerHostProto, defaultDockerHostAddr, defaultTimeout)
			},
			DisableCompression: true,
		},
		CheckRedirect: docker.CheckRedirect,
	}
	t.client, retErr = docker.NewClientWithOpts(docker.WithHTTPClient(c))
	if retErr != nil {
		logrus.Errorf("create docker client failed: %v", retErr)
		return errors.Wrap(retErr, "create docker client failed")
	}
	t.sd, retErr = t.initStorageDriver()
	if retErr != nil {
		logrus.Errorf("init storage driver failed: %v", retErr)
		return errors.Wrap(retErr, "init storage driver failed")
	}
	return t.initContainers()
}

func (t *dockerTransformer) Transform(ids []string, all bool, retCh chan transform.Result) {
	if all {
		t.ctrs.Range(func(k, _ interface{}) bool {
			id, _ := k.(string)
			ids = append(ids, id)
			return true
		})
	}

	signalWg := new(sync.WaitGroup)
	signalCtx := t.handleSignal()

	var wg sync.WaitGroup
	for _, ctr := range ids {
		wg.Add(1)
		go func(id string) {
			var ret transform.Result
			switch ctrID, ok := t.matchID(id); ok {
			case notExist:
				ret.Msg = fmt.Sprintf("transform %s: container was not found", id)
			case hasBeenTransformed:
				ret.Ok = true
				ret.Msg = fmt.Sprintf("transform %s: container has been transformed", id)
			case needTransform:
				err := t.transform(ctrID, newRollback(signalCtx, signalWg))
				if err != nil {
					ret.Msg = fmt.Sprintf("transform %s: %s", id, err.Error())
				} else {
					ret.Ok = true
					ret.Msg = fmt.Sprintf("transform %s: success", id)
				}
			default:
			}
			retCh <- ret
			wg.Done()
		}(ctr)
	}
	wg.Wait()
	signalWg.Wait()
	close(retCh)
}

func (t *dockerTransformer) transform(id string, rb *rollback) error {
	var (
		retErr error

		hostCfg *types.IsuladHostConfig
		logCfg  *container.LogConfig
		v2Cfg   *types.IsuladV2Config
		ociCfg  *specs.Spec

		oldRootFs string
	)

	rb.wait()
	defer func() {
		if retErr != nil {
			rb.run()
		}
		rb.close()
	}()

	logrus.Infof("start to transform %s", id)

	// before transform, pause container to suspend all processes in a container
	retErr = t.client.ContainerPause(context.Background(), id)
	if retErr != nil && !strings.Contains(retErr.Error(), "already paused") {
		logrus.Errorf("pause container %s failed: %v", id, retErr)
		return errors.Wrap(retErr, "pause container")
	}

	// init
	iSulad := isulad.GetIsuladTool()
	retErr = iSulad.PrepareBundleDir(id)
	if retErr != nil {
		logrus.Errorf("prepare bundle dir failed: %v", retErr)
		return errors.Wrap(retErr, "prepare root dir")
	}
	rb.register(func() {
		logrus.Infof("rollback: clean up bundle dir of container %s", id)
		if err := iSulad.Cleanup(id); err != nil {
			logrus.Warnf("rollback: clean up bundle dir of container %s: %v", id, err)
		}
	})

	// start transform
	// transform hostConfig: hostconfig.json
	hostCfg, logCfg, retErr = t.transformHostConfig(id)
	if retErr != nil {
		logrus.Errorf("transform hostconfig failed: %v", retErr)
		return errors.Wrap(retErr, "transform hostconfig")
	}

	// transform config.v2: config.v2.json
	reconcileOpts := append(genV2OptsFromHostCfg(hostCfg),
		v2ConfigWithLogConfig(logCfg, filepath.Join(iSulad.GetRuntimePath(), id)))
	v2Cfg, retErr = t.transformV2Config(id, reconcileOpts...)
	if retErr != nil {
		logrus.Errorf("transform configV2 failed: %v", retErr)
		return errors.Wrap(retErr, "transform configV2")
	}
	rb.register(func() {
		logrus.Infof("rollback: clean up storage register of container %s", id)
		t.sd.Cleanup(id)
	})

	// share shm : mounts/shm
	retErr = iSulad.PrepareShm(v2Cfg.CommonConfig.ShmPath, hostCfg.ShmSize)
	if retErr != nil {
		logrus.Errorf("prepare share shm failed: %v", retErr)
		return errors.Wrap(retErr, "prepare share shm")
	}
	rb.register(func() {
		logrus.Infof("rollback: umount share shm of container %s path %s", id, v2Cfg.CommonConfig.ShmPath)
		if umountErr := unix.Unmount(v2Cfg.CommonConfig.ShmPath, unix.MNT_DETACH); umountErr != nil {
			logrus.Warnf("umount %s err: %v", v2Cfg.CommonConfig.ShmPath, umountErr)
		}
	})

	// copy linux network files : hostname hosts resolve.conf
	files := []string{types.Hostname, types.Hosts, types.Resolv}
	for idx := range files {
		srcF := v2Cfg.CommonConfig.GetOriginNetworkFile(files[idx])
		destF := iSulad.GetNetworkFilePath(id, files[idx])
		retErr = exec.Command("cp", "-a", srcF, destF).Run()
		if retErr != nil {
			logrus.Errorf("copy %s to %s failed", srcF, destF)
			return errors.Wrapf(retErr, "copy %s to %s failed", srcF, destF)
		}
	}

	// oci spec: config.json
	ociCfg, oldRootFs, retErr = t.transformOciConfig(id, v2Cfg.CommonConfig, hostCfg)
	if retErr != nil {
		logrus.Errorf("transform oci spec config failed: %v", retErr)
		return errors.Wrap(retErr, "transform oci spec")
	}

	// copy RWlayer
	retErr = t.sd.TransformRWLayer(v2Cfg, oldRootFs)
	if retErr != nil {
		logrus.Errorf("storage driver transform RWLayer failed: %v", retErr)
		return errors.Wrap(retErr, "transform RWLayer")
	}

	// lcr_create: config  ocihooks.json  seccomp
	ociCfgData, err := json.Marshal(ociCfg)
	if err != nil {
		logrus.Errorf("marshal oci config failed: %s", err)
		return errors.Wrap(err, "marshal oci config")
	}
	retErr = iSulad.LcrCreate(id, ociCfgData)
	if retErr != nil {
		logrus.Error("lcr create failed")
		return retErr
	}

	logrus.Infof("transform %s successfully", id)

	return nil
}

func (t *dockerTransformer) transformHostConfig(id string) (*types.IsuladHostConfig, *container.LogConfig, error) {
	var isuladHostCfg types.IsuladHostConfig
	var l container.LogConfig
	var hostCfg = struct {
		*types.IsuladHostConfig
		LogConfig *container.LogConfig
	}{
		&isuladHostCfg,
		&l,
	}

	hostCfgPath := filepath.Join(t.GraphRoot, "containers", id, types.Hostconfig)
	err := utils.CheckFileValid(hostCfgPath)
	if err != nil {
		logrus.Errorf("check docker hostconfig.json failed: %v", err)
		return nil, nil, errors.Wrap(err, "check docker hostconfig.json")
	}

	data, err := ioutil.ReadFile(hostCfgPath)
	if err != nil {
		logrus.Errorf("read container %s's host config file %s failed: %v", id, hostCfgPath, err)
		return nil, nil, errors.Wrap(err, "read hostconfig.json")
	}

	err = json.Unmarshal(data, &hostCfg)
	if err != nil {
		logrus.Errorf("can't unmarshal container %s's host config to iSulad type: %v", id, err)
		return nil, nil, errors.Wrap(err, "unmarshal host config data")
	}

	iSulad := isulad.GetIsuladTool()
	reconcileHostConfig(&isuladHostCfg, iSulad.Runtime())
	err = iSulad.SaveConfig(id, &isuladHostCfg, iSulad.MarshalIndent, iSulad.GetHostCfgPath)
	if err != nil {
		logrus.Errorf("save host config to file %s failed", iSulad.GetHostCfgPath(id))
		return nil, nil, errors.Wrap(err, "save hostconfig.json")
	}

	return &isuladHostCfg, &l, nil
}

func (t *dockerTransformer) loadV2Config(id string) (*types.DockerV2Config, error) {
	var dockerV2Cfg = types.DockerV2Config{}

	v2Path := filepath.Join(t.GraphRoot, "containers", id, types.V2config)

	err := utils.CheckFileValid(v2Path)
	if err != nil {
		logrus.Errorf("check docker config.v2.json failed: %v", err)
		return nil, errors.Wrap(err, "check docker config.v2.json")
	}

	data, err := ioutil.ReadFile(v2Path)
	if err != nil {
		logrus.Errorf("can't read container %s's v2 config file %s with %v", id, v2Path, err)
		return nil, errors.Wrap(err, "read config.v2.json")
	}
	err = json.Unmarshal(data, &dockerV2Cfg)
	if err != nil {
		logrus.Errorf("can't unmarshal container %s's v2 config %s with %v", id, v2Path, err)
		return nil, errors.Wrap(err, "unmarshal v2 config data")
	}
	return &dockerV2Cfg, nil
}

func (t *dockerTransformer) transformV2Config(id string, opts ...v2ConfigReconcileOpt) (*types.IsuladV2Config, error) {
	var (
		iSuladState  = types.ContainerState{}
		iSuladCommon = types.CommonConfig{}
		iSuladV2Cfg  = types.IsuladV2Config{
			CommonConfig: &iSuladCommon,
			Image:        "",
			State:        &iSuladState,
		}

		iSulad = isulad.GetIsuladTool()
	)

	ctr, err := t.loadV2Config(id)
	if err != nil {
		logrus.Errorf("load container %s's v2 config failed: %v", id, err)
		return nil, errors.Wrap(err, "load container config")
	}

	commonData, err := json.Marshal(ctr)
	if err != nil {
		logrus.Errorf("internal error: %v", err)
		return nil, errors.Wrap(err, "common data internal error")
	}
	err = json.Unmarshal(commonData, &iSuladCommon)
	if err != nil {
		logrus.Errorf("internal error: %v", err)
		return nil, errors.Wrap(err, "common config internal error")
	}

	stateData, err := json.Marshal(ctr.State)
	if err != nil {
		logrus.Errorf("internal error: %v", err)
		return nil, errors.Wrap(err, "state data internal error")
	}
	err = json.Unmarshal(stateData, &iSuladState)
	if err != nil {
		logrus.Errorf("internal error: %v", err)
		return nil, errors.Wrap(err, "state config internal error")
	}

	iSuladV2Cfg.Image = ctr.ImageID

	basePath := filepath.Join(iSulad.GetRuntimePath(), id)
	opts = append(opts, []v2ConfigReconcileOpt{
		v2ConfigWithImage(ctr.Config.Image),
		v2ConfigWithCgroupParent(ctr.CgroupParent),
	}...)
	reconcileV2Config(&iSuladV2Cfg, basePath, opts...)

	iSuladCommon.BaseFs, err = t.sd.GenerateRootFs(id, iSuladCommon.Image)
	if err != nil {
		logrus.Errorf("storage driver generate new rootfs failed: %v", err)
		return nil, errors.Wrap(err, "generate new rootfs")
	}

	err = iSulad.SaveConfig(id, &iSuladV2Cfg, iSulad.MarshalIndent, iSulad.GetConfigV2Path)
	if err != nil {
		logrus.Errorf("save v2 config to file %s failed", iSulad.GetConfigV2Path(id))
		return nil, errors.Wrap(err, "save config.v2.json")
	}

	return &iSuladV2Cfg, nil
}

func (t *dockerTransformer) transformOciConfig(id string,
	commonCfg *types.CommonConfig, hostCfg *types.IsuladHostConfig) (*specs.Spec, string, error) {
	var ociConfig = specs.Spec{}

	// load
	// path like: /var/run/docker/containerd/daemon/io.containerd.runtime.v1.linux/moby/ctr_id/config.json
	ociPath := filepath.Join(t.StateRoot, "containerd/daemon", containerdRuntime, containerdNameSpace, id, types.Ociconfig)

	err := utils.CheckFileValid(ociPath)
	if err != nil {
		logrus.Errorf("check docker config.json failed: %v", err)
		return nil, "", errors.Wrap(err, "check docker config.json")
	}

	data, err := ioutil.ReadFile(ociPath)
	if err != nil {
		logrus.Errorf("can't read oci config file, check if the container %s is running", id)
		return nil, "", errors.Wrap(err, "read oci config.json")
	}
	err = json.Unmarshal(data, &ociConfig)
	if err != nil {
		logrus.Errorf("can't unmarshal container %s's ociconfig %s with %v", id, ociPath, err)
		return nil, "", errors.Wrap(err, "unmarshal oci config data")
	}

	// reconcile
	oldRoot := ociConfig.Root.Path
	reconcileOciConfig(&ociConfig, commonCfg, hostCfg)

	// save
	iSulad := isulad.GetIsuladTool()
	err = iSulad.SaveConfig(id, &ociConfig, iSulad.MarshalIndent, iSulad.GetOciConfigPath)
	if err != nil {
		logrus.Errorf("save v2 config to file %s failed", iSulad.GetOciConfigPath(id))
		return nil, "", errors.Wrap(err, "save config.json")
	}

	return &ociConfig, oldRoot, nil
}

func (t *dockerTransformer) initStorageDriver() (transform.StorageDriver, error) {
	iSulad := isulad.GetIsuladTool()
	switch iSulad.StorageType() {
	case transform.Overlay2:
		return newOverlayDriver(iSulad.BaseStorageDriver()), nil
	case transform.DeviceMapper:
		return newDeviceMapperDriver(iSulad.BaseStorageDriver(), t.client), nil
	default:
	}
	return nil, fmt.Errorf("unsupported storage driver type: %s", iSulad.StorageType())
}

func (t *dockerTransformer) initContainers() error {
	t.ctrs = &sync.Map{}
	runningCtrsRoot := filepath.Join(t.StateRoot, "containerd/daemon", containerdRuntime, containerdNameSpace)
	infos, err := ioutil.ReadDir(runningCtrsRoot)
	if err != nil {
		return errors.Wrap(err, "init docker container store failed")
	}
	for _, info := range infos {
		if info.IsDir() && len(info.Name()) == containerIDLen {
			t.ctrs.Store(info.Name(), false)
		}
	}
	return nil
}

func (t *dockerTransformer) matchID(id string) (fullID string, status containerStatus) {
	t.ctrs.Range(func(k, v interface{}) bool {
		fullID, status = "", notExist
		tmpID, _ := k.(string)
		transformed, _ := v.(bool)
		if strings.HasPrefix(tmpID, id) {
			if !transformed {
				t.ctrs.Store(k, true)
				status = needTransform
			} else {
				status = hasBeenTransformed
			}
			fullID = tmpID
			return false
		}
		return true
	})
	return
}

func (t *dockerTransformer) handleSignal() context.Context {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if s, ok := <-sigCh; ok {
			logrus.Infof("catch signal %v", s)
			cancel()
			return
		}
	}()

	return ctx
}
