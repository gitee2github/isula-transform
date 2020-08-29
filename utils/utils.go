/*
 * Copyright (c) 2020 Huawei Technologies Co., Ltd.
 * isula-transform is licensed under the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Create: 2020-04-28
 */

// Package utils contains common functions
package utils

import (
	"os"

	"github.com/pkg/errors"
)

const (
	maxFileSize int64 = 10 * 1024 * 1024
)

// CheckFileValid make sure the path is a file which less then 10M
func CheckFileValid(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "stat file %s", path)
	}
	if fileInfo.IsDir() {
		return errors.Errorf("%s should not be a directory", path)
	}
	// file should be a file less than 10M
	if fileInfo.Size() > maxFileSize {
		return errors.Errorf("size of %s is lager then MAX_FILE_SIZE(10M)", path)
	}
	return nil
}
