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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"isula.org/isula-transform/types"
)

const (
	logDriverJSONFile = "json-file"
	logDriverSyslog   = "syslog"

	defaultLogSize   = "30KB"
	defaultLogRotate = "7"
	defaultLogDriver = logDriverJSONFile
	defaultLogPath   = "none"

	defaultCgroupDir = "/isulad"

	minValidImagePartsLen = 2
)

type v2ConfigReconcileOpt func(*types.IsuladV2Config)

func v2ConfigWithImage(image string) v2ConfigReconcileOpt {
	return func(v2 *types.IsuladV2Config) {
		v2.CommonConfig.Image = imageRemoveSuffixDigest(image)
		v2.CommonConfig.ImageType = "oci"
	}
}

func v2ConfigWithCgroupParent(cgroupParent string) v2ConfigReconcileOpt {
	return func(v2 *types.IsuladV2Config) {
		var cgroupDir string
		if strings.HasPrefix(cgroupParent, "/docker/") || !strings.HasSuffix(cgroupParent, "/"+v2.CommonConfig.ID) {
			cgroupDir = defaultCgroupDir
		} else {
			cgroupDir = strings.TrimSuffix(cgroupParent, "/"+v2.CommonConfig.ID)
		}
		v2.CommonConfig.Config.Annotations["cgroup.dir"] = cgroupDir
	}
}

func v2ConfigWithOomScoreAdj(oomScore int) v2ConfigReconcileOpt {
	return func(v2 *types.IsuladV2Config) {
		v2.CommonConfig.Config.Annotations["proc.oom_score_adj"] = strconv.Itoa(oomScore)
	}
}

func v2ConfigWithFilesLimit(filesLimit int64) v2ConfigReconcileOpt {
	return func(v2 *types.IsuladV2Config) {
		v2.CommonConfig.Config.Annotations["files.limit"] = strconv.FormatInt(filesLimit, 10)
	}
}

func v2ConfigWithLogConfig(cfg *container.LogConfig, basePath string) v2ConfigReconcileOpt {
	return func(v2 *types.IsuladV2Config) {
		if cfg == nil {
			cfg = &container.LogConfig{
				Type: "default",
			}
		}
		switch cfg.Type {
		case logDriverJSONFile:
			// docker allowed logopts:
			//   max-file、max-size、compress、labels、env、env-regex、tag
			// isulad only support:
			//   max-file, max-size
			v2.CommonConfig.LogDriver = logDriverJSONFile
			v2.CommonConfig.Config.Annotations["log.console.driver"] = logDriverJSONFile
			v2.CommonConfig.LogPath = filepath.Join(basePath, "console.log")
			v2.CommonConfig.Config.Annotations["log.console.file"] = v2.CommonConfig.LogPath
			if rotate, exist := cfg.Config["max-file"]; exist {
				v2.CommonConfig.Config.Annotations["log.console.filerotate"] = rotate
			} else {
				v2.CommonConfig.Config.Annotations["log.console.filerotate"] = defaultLogRotate
			}
			if size, exist := cfg.Config["max-size"]; exist {
				v2.CommonConfig.Config.Annotations["log.console.filesize"] = size
			} else {
				v2.CommonConfig.Config.Annotations["log.console.filesize"] = defaultLogSize
			}
		case logDriverSyslog:
			v2.CommonConfig.LogDriver = logDriverSyslog
			v2.CommonConfig.Config.Annotations["log.console.driver"] = logDriverSyslog
			// docker allowed LogOpts:
			//   env, env-regex, labels, syslog-facility, tag, syslog-format
			//   syslog-address, syslog-tls-ca-cert, syslog-tls-cert, syslog-tls-key, syslog-tls-skip-verify
			// isulad only support:
			//   tag, facility
			if tag, exist := cfg.Config["tag"]; exist {
				v2.CommonConfig.Config.Annotations["log.console.tag"] = tag
			}
			if facility, exist := cfg.Config["syslog-facility"]; exist {
				v2.CommonConfig.Config.Annotations["log.console.facility"] = facility
			}
		default:
			// use isulad default driver without file
			v2.CommonConfig.LogDriver = defaultLogDriver
			v2.CommonConfig.LogPath = defaultLogPath
			v2.CommonConfig.Config.Annotations["log.console.driver"] = defaultLogDriver
			v2.CommonConfig.Config.Annotations["log.console.file"] = defaultLogPath
			v2.CommonConfig.Config.Annotations["log.console.filerotate"] = defaultLogRotate
			v2.CommonConfig.Config.Annotations["log.console.filesize"] = defaultLogSize
		}
	}
}

func genV2OptsFromHostCfg(h *types.IsuladHostConfig) []v2ConfigReconcileOpt {
	if h == nil {
		return nil
	}
	var opts []v2ConfigReconcileOpt
	if h.OomScoreAdj != 0 {
		opts = append(opts, v2ConfigWithOomScoreAdj(h.OomScoreAdj))
	}
	if h.FilesLimit != 0 {
		opts = append(opts, v2ConfigWithFilesLimit(h.FilesLimit))
	}
	return opts
}

func reconcileV2Config(v2 *types.IsuladV2Config, basePath string, opts ...v2ConfigReconcileOpt) {
	for _, o := range opts {
		o(v2)
	}

	// modify hosts hostname and resolv.conf
	v2.CommonConfig.OriginHostsPath = v2.CommonConfig.HostsPath
	v2.CommonConfig.HostsPath = filepath.Join(basePath, types.Hosts)
	v2.CommonConfig.OriginHostnamePath = v2.CommonConfig.HostnamePath
	v2.CommonConfig.HostnamePath = filepath.Join(basePath, types.Hostname)
	v2.CommonConfig.OriginResolvConfPath = v2.CommonConfig.ResolvConfPath
	v2.CommonConfig.ResolvConfPath = filepath.Join(basePath, types.Resolv)
	v2.CommonConfig.ShmPath = filepath.Join(basePath, "mounts", "shm")

	// add annotations
	v2.CommonConfig.Config.Annotations["rootfs.mount"] = "/var/lib/isulad/mnt/rootfs"

	// fix time format and update state
	v2.CommonConfig.Created = v2.CommonConfig.Created.Local()
	v2.State.Paused = false
	v2.State.Running = false
	v2.State.StartedAt = v2.State.StartedAt.Local()
	v2.State.FinishedAt = time.Now().Local()

	// fix name prefix
	v2.CommonConfig.Name = strings.TrimPrefix(v2.CommonConfig.Name, "/")
}

func imageRemoveSuffixDigest(reference string) string {
	reg := regexp.MustCompile(`([a-zA-Z0-9._\-/:]*)(?:@[[:xdigit:]]+)?`)
	matchs := reg.FindStringSubmatch(reference)
	if len(matchs) < minValidImagePartsLen {
		return ""
	}
	return matchs[1]
}

func reconcileHostConfig(h *types.IsuladHostConfig, runtime string) {
	h.Runtime = runtime
	if h.RestartPolicy.Name == "unless-stopped" {
		logrus.Info("isulad not support unless-stopped policy, transform to always")
		h.RestartPolicy.Name = "always"
	}
	if h.UsernsMode != "" {
		logrus.Infof("isulad not allowed share user namespace %s, replace to nil", h.UsernsMode)
		h.UsernsMode = ""
	}
}

func reconcileOciConfig(s *specs.Spec, c *types.CommonConfig, h *types.IsuladHostConfig) {
	// Annotations opt sync with CommonConfig
	for k, v := range c.Config.Annotations {
		s.Annotations[k] = v
	}

	// rootFs opt
	s.Root.Path = c.BaseFs

	// user-defined cgroup path
	if _, exist := s.Annotations["cgroup.dir"]; !exist {
		s.Annotations["cgroup.dir"] = defaultCgroupDir
	}
	cgroupDir := s.Annotations["cgroup.dir"]
	s.Linux.CgroupsPath = filepath.Join(cgroupDir, c.ID)

	// set linux namespace path
	// pid, ipc and network might be container mode
	for idx := range s.Linux.Namespaces {
		s.Linux.Namespaces[idx].Path = ociAdaptSharedNamespaceContainer(s.Linux.Namespaces[idx], h)
	}

	// when privileged, there is no need to create pty device
	if !h.Privileged {
		ociAddMustDevice(s)
	}

	// mounts opt
	for idx := range s.Mounts {
		var source string
		switch s.Mounts[idx].Destination {
		case "/etc/hostname":
			source = c.HostnamePath
		case "/etc/resolv.conf":
			source = c.ResolvConfPath
		case "/etc/hosts":
			source = c.HostsPath
		case "/dev/shm":
			source = c.ShmPath
			shmModeOpt := "mode=1777"
			shmSizeOpt := "size=" + strconv.FormatInt(h.ShmSize, 10)
			s.Mounts[idx].Options = append(s.Mounts[idx].Options, shmModeOpt, shmSizeOpt)
		default:
			continue
		}
		s.Mounts[idx].Source = source
	}

	// when device.Path contains ":", lxc does not support, remove them
	end := 0
	for _, device := range s.Linux.Devices {
		if !strings.Contains(device.Path, ":") {
			s.Linux.Devices[end] = device
			end++
		}
	}
	s.Linux.Devices = s.Linux.Devices[:end]
}

func ociAdaptSharedNamespaceContainer(ns specs.LinuxNamespace, h *types.IsuladHostConfig) string {
	isContainer := func(mode string) (string, bool) {
		parts := strings.SplitN(mode, ":", 2)
		if len(parts) > 1 && parts[0] == "container" && len(parts[1]) == containerIDLen {
			return parts[1], true
		}
		return "", false
	}

	switch ns.Type {
	case specs.IPCNamespace:
		if ipcCtr, ok := isContainer(h.IpcMode); ok {
			return ipcCtr
		}
	case specs.PIDNamespace:
		if pidCtr, ok := isContainer(h.PidMode); ok {
			return pidCtr
		}
	case specs.NetworkNamespace:
		if netCtr, ok := isContainer(h.NetworkMode); ok {
			return netCtr
		}
	default:
	}
	return ""
}

func ociAddMustDevice(spec *specs.Spec) {
	addDeviceFunc := func(major, minor int64) {
		dev := specs.LinuxDeviceCgroup{
			Type:   "c",
			Allow:  true,
			Major:  &major,
			Minor:  &minor,
			Access: "rwm",
		}

		for _, item := range spec.Linux.Resources.Devices {
			if (item.Major != nil && *dev.Major == *item.Major) &&
				(item.Minor != nil && *dev.Minor == *item.Minor) {
				return
			}
		}
		spec.Linux.Resources.Devices = append(spec.Linux.Resources.Devices, dev)
	}

	// ptmx: PTY master multiplex
	mustDevices := []struct {
		Major int64
		Minor int64
	}{
		{
			Major: 5,
			Minor: 2,
		},
		{
			Major: 136,
			Minor: -1,
		},
	}

	for _, dev := range mustDevices {
		addDeviceFunc(dev.Major, dev.Minor)
	}
}
