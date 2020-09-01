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

/*
#cgo LDFLAGS: -L/usr/lib64 -llcr -llxc
#include <stdlib.h>
#include <lcr/lcrcontainer.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const isuladLogGatherFIFOName = "/isulad_log_gather_fifo"

// LcrLogInit initlizes the lcr log opt
func LcrLogInit(iSuladState, iSuladRuntime, logLevel string) {
	name := C.CString("isulad")
	file := C.CString("fifo:" + iSuladState + isuladLogGatherFIFOName)
	priority := C.CString(logLevel)
	prefix := C.CString(iSuladRuntime)
	quiet := C.int(1)
	defer func() {
		cFreeChar(name, file, priority, prefix)
	}()
	C.lcr_log_init(name, file, priority, prefix, quiet, nil)
}

// lcrCreate call c func lcr_create_from_ocidata to create config, ocihooks.json and seccomp
func lcrCreate(id, lcrPath string, spec []byte) error {
	name := C.CString(id)
	lcrpath := C.CString(lcrPath)
	ociConfigData := unsafe.Pointer(&spec[0])
	defer func() {
		cFreeChar(name, lcrpath)
	}()

	if !C.lcr_create_from_ocidata(name, lcrpath, ociConfigData) {
		return fmt.Errorf("lcr create failed")
	}
	return nil
}

func cFreeChar(cs ...*C.char) {
	for i := range cs {
		C.free(unsafe.Pointer(cs[i]))
	}
}
