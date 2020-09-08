/*
 * Copyright (c) 2020 Huawei Technologies Co., Ltd.
 * isula-transform is licensed under the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Create: 2020-09-04
 */

package isuladimg

/*
#cgo LDFLAGS: -lisulad_img -lisula_libutils
#include "isuladimg.h"
*/
import "C"

import (
	"unsafe"

	"github.com/pkg/errors"
)

var (
	imageTypeOCI string = "oci"
)

type Operation int

const (
	RemoveOp Operation = iota
	MountOp
	UmountOp
)

// InitLib initializes the dynamic link library of libisulad_img
func InitLib(graph, state, driverType string, driverOpts []string, check bool) error {
	var imageLayerCheck C.int

	cGraph := C.CString(graph)
	defer C.free(unsafe.Pointer(cGraph))
	cState := C.CString(state)
	defer C.free(unsafe.Pointer(cState))
	cStorageDriver := C.CString(driverType)
	defer C.free(unsafe.Pointer(cStorageDriver))
	opts := make([]*C.char, len(driverOpts))
	for i := range driverOpts {
		opts[i] = C.CString(driverOpts[i])
		defer C.free(unsafe.Pointer(opts[i]))
	}
	var cOpts **C.char
	if len(driverOpts) == 0 {
		cOpts = nil
	} else {
		cOpts = (**C.char)(unsafe.Pointer(&opts[0]))
	}
	if check {
		imageLayerCheck = C.int(1)
	} else {
		imageLayerCheck = C.int(0)
	}

	if ret := C.init_isulad_image_module(cGraph, cState, cStorageDriver,
		cOpts, C.size_t(len(driverOpts)), imageLayerCheck); ret != 0 {
		return errors.Errorf("init libisulad_img.so get ret code: %d", ret)
	}
	return nil
}

// PrepareRootfs calls isulad_img_prepare_rootfs to prepare container rootfs
func PrepareRootfs(id, image string) string {
	imageType := C.CString(imageTypeOCI)
	defer C.free(unsafe.Pointer(imageType))
	containerID := C.CString(id)
	defer C.free(unsafe.Pointer(containerID))
	imageName := C.CString(image)
	defer C.free(unsafe.Pointer(imageName))

	realRootfs := C.isulad_img_prepare_rootfs(imageType, containerID, imageName)
	mountPoint := C.GoString(realRootfs)
	return mountPoint
}

// SwitchOperation choose different Operation for container rootfs
func SwitchOperation(op Operation, id, image string) C.int {
	imageType := C.CString(imageTypeOCI)
	defer C.free(unsafe.Pointer(imageType))
	containerID := C.CString(id)
	defer C.free(unsafe.Pointer(containerID))
	imageName := C.CString(image)
	defer C.free(unsafe.Pointer(imageName))

	switch op {
	case RemoveOp:
		return C.im_remove_container_rootfs(imageType, containerID)
	case MountOp:
		return C.im_mount_container_rootfs(imageType, imageName, containerID)
	case UmountOp:
		return C.im_umount_container_rootfs(imageType, imageName, containerID)
	}
	return -1
}
