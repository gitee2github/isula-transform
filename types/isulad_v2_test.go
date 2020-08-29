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

package types

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCommonConfig_GetOriginNetworkFile(t *testing.T) {
	var (
		hostname = "/test/hostname"
		hosts    = "/test/hosts"
		resolv   = "/test/resolv.conf"
	)
	Convey("TestCommonConfig_GetOriginNetworkFile", t, func() {
		common := &CommonConfig{
			OriginHostnamePath:   hostname,
			OriginHostsPath:      hosts,
			OriginResolvConfPath: resolv,
		}
		Convey("get hostname", func() {
			So(common.GetOriginNetworkFile(Hostname), ShouldEqual, hostname)
		})
		Convey("get hosts", func() {
			So(common.GetOriginNetworkFile(Hosts), ShouldEqual, hosts)
		})
		Convey("get resolv.conf", func() {
			So(common.GetOriginNetworkFile(Resolv), ShouldEqual, resolv)
		})
		Convey("get empty", func() {
			So(common.GetOriginNetworkFile("notANetworkFile"), ShouldBeBlank)
		})
	})
}
