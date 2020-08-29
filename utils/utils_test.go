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

package utils

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckFileValid(t *testing.T) {
	Convey("TestCheckFileValid", t, func() {
		Convey("file not exist", func() {
			err := CheckFileValid("/not/exist/in/host")
			So(err, ShouldBeError)
		})

		Convey("size of file more then 10M", func() {
			f, err := ioutil.TempFile("", "largefile")
			if err != nil {
				t.Skipf("create tmp file failed: %v", err)
			}
			defer func() {
				f.Close()
				os.Remove(f.Name())
			}()

			Convey("normal", func() {
				if err := f.Truncate(maxFileSize - 1); err != nil {
					t.Skipf("resize tmp file failed: %v", err)
				}
				got := CheckFileValid(f.Name())
				So(got, ShouldBeNil)
			})

			Convey("size of file more then 10M", func() {
				if err := f.Truncate(maxFileSize + 1); err != nil {
					t.Skipf("resize tmp file failed: %v", err)
				}
				got := CheckFileValid(f.Name())
				So(got, ShouldBeError)
				So(got.Error(), ShouldContainSubstring, "lager then MAX_FILE_SIZE(10M)")
			})
		})

		Convey("is a directory", func() {
			path, err := ioutil.TempDir("", "iamadir")
			if err != nil {
				t.Skipf("create tmp large file failed: %v", err)
			}
			defer os.RemoveAll(path)
			got := CheckFileValid(path)
			So(got, ShouldBeError)
			So(got.Error(), ShouldContainSubstring, "should not be a directory")
		})
	})
}
