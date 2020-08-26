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
	"os/exec"
	"strings"

	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
)

type overlayDriver struct {
	transform.BaseStorageDriver
}

func newOverlayDriver(base transform.BaseStorageDriver) transform.StorageDriver {
	return &overlayDriver{base}
}

func (od *overlayDriver) GenerateRootFs(id, image string) (string, error) {
	return od.BaseStorageDriver.GenerateRootFs(id, image)
}

// only copy diff from old to new
func (od *overlayDriver) TransformRWLayer(ctr *types.IsuladV2Config, oldRootFs string) error {
	srcRoot := strings.TrimSuffix(oldRootFs, "/merged")
	destRoot := strings.TrimSuffix(ctr.CommonConfig.BaseFs, "/merged")
	if err := exec.Command("cp", "-ra", srcRoot+"/diff", destRoot).Run(); err != nil {
		return err
	}
	return nil
}

func (od *overlayDriver) Cleanup(id string) {
	od.BaseStorageDriver.CleanupRootFs(id)
}
