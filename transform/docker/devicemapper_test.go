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
	"sort"
	"testing"

	"github.com/docker/docker/api/types/container"
	. "github.com/google/go-cmp/cmp"
	. "github.com/smartystreets/goconvey/convey"
	"isula.org/isula-transform/types"
)

func Test_deviceMapperDriver_changesFilter(t *testing.T) {
	Convey("Test_deviceMapperDriver_changesFilter", t, func() {
		dm := &deviceMapperDriver{}

		testDiff := []container.ContainerChangeResponseItem{
			{Kind: changeItem, Path: "/etc"},            // root filter
			{Kind: delItem, Path: "/etc/delfile"},       // delete save
			{Kind: changeItem, Path: "/etc/os-release"}, // change save
			{Kind: changeItem, Path: "/root"},           // root filter
			{Kind: addItem, Path: "/root/add"},          // add save
			{Kind: addItem, Path: "/root/padd"},         // parent filter
			{Kind: addItem, Path: "/root/padd/subadd"},  // add save
			{Kind: addItem, Path: "/root/mount"},        // mount filter
		}
		testMount := map[string]types.Mount{
			"/root/mount": {Destination: "/root/mount"},
		}

		expect := []container.ContainerChangeResponseItem{
			{Kind: delItem, Path: "/etc/delfile"},       // delete save
			{Kind: changeItem, Path: "/etc/os-release"}, // change save
			{Kind: addItem, Path: "/root/add"},          // add save
			{Kind: addItem, Path: "/root/padd/subadd"},  // add save
		}

		got := dm.changesFilter(testDiff, testMount)

		sortChanges := func(changes []container.ContainerChangeResponseItem) {
			sort.SliceStable(changes, func(i, j int) bool {
				return changes[i].Path < changes[j].Path
			})
		}
		sortChanges(expect)
		sortChanges(got)

		So(Diff(got, expect), ShouldBeBlank)
	})
}
