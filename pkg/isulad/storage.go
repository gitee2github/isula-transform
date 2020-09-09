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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"isula.org/isula-transform/pkg/isulad/internal/isuladimg"
)

type isuladStorageDriver struct{}

func (sd *isuladStorageDriver) GenerateRootFs(id, image string) (string, error) {
	mountPoint := isuladimg.PrepareRootfs(id, image)
	if mountPoint == "" {
		return "", errors.New("isuladimg returns nil rootfs")
	}
	return mountPoint, nil
}

func (sd *isuladStorageDriver) CleanupRootFs(id string) {
	if ret := isuladimg.SwitchOperation(isuladimg.RemoveOp, id, ""); ret != 0 {
		logrus.Warnf("remove container %s's rootfs get code: %d", id, ret)
	} else {
		logrus.Infof("remove container %s's rootfs successful", id)
	}
}

func (sd *isuladStorageDriver) MountRootFs(id, image string) error {
	if ret := isuladimg.SwitchOperation(isuladimg.MountOp, id, image); ret != 0 {
		return errors.Errorf("mount container %s's rootfs get ret code: %d", id, ret)
	}
	return nil
}

func (sd *isuladStorageDriver) UmountRootFs(id, image string) error {
	if ret := isuladimg.SwitchOperation(isuladimg.UmountOp, id, image); ret != 0 {
		return errors.Errorf("umount container %s's rootfs get ret code: %d", id, ret)
	}
	return nil
}
