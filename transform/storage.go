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

package transform

import (
	"time"

	"isula.org/isula-transform/types"
)

// StorageType define graph storage driver
type StorageType string

// support storage driver
const (
	Overlay2         StorageType = "overlay2"
	DeviceMapper     StorageType = "devicemapper"
	defaultAddress               = "unix:///var/run/isuald/isula_image.sock"
	isuladImgTimeout             = 10 * time.Second
)

// StorageDriver defines methods for creating and rolling storage resources
type StorageDriver interface {
	// GenerateRootFs returns a new rootfs path used by container
	GenerateRootFs(id, image string) (string, error)
	// TransformRWLayer migrates container read-write layer data
	TransformRWLayer(ctr *types.IsuladV2Config, oldRootFs string) error
	// Cleanup olls back the image operation when the transformation fails
	Cleanup(id string)
}

// BaseStorageDriver contains the common functions used by StorageDriver
type BaseStorageDriver interface {
	GenerateRootFs(id, image string) (string, error)
	CleanupRootFs(id string)
	MountRootFs(id string) error
	UmountRootFs(id string) error
}
