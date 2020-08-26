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
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"isula.org/isula-transform/api/isula"
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

var (
	gBaseStorageDriver BaseStorageDriver
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

type baseStorageDriver struct {
	imgClient isula.ImageServiceClient
}

func initBaseStorageDriver(addr string) error {
	client, err := newIsuladImgClient(addr)
	if err != nil {
		return err
	}
	gBaseStorageDriver = &baseStorageDriver{imgClient: client}
	return nil
}

// GenerateRootFs returns a new rootfs path of container
func (sd *baseStorageDriver) GenerateRootFs(id, image string) (string, error) {
	req := &isula.ContainerPrepareRequest{
		Image: image,
		Id:    id,
		Name:  id,
	}
	resp, err := sd.imgClient.ContainerPrepare(context.Background(), req)
	if err != nil {
		return "", err
	}
	if msg := resp.GetErrmsg(); msg != "" {
		removeReq := &isula.ContainerRemoveRequest{
			NameId: id,
		}
		rResp, rErr := sd.imgClient.ContainerRemove(context.Background(), removeReq)
		logrus.Infof("isulad-img remove container: %v, err: %v", rResp, rErr)
		return "", fmt.Errorf("isulad-img prepare failed: %s", msg)
	}
	return resp.MountPoint, nil
}

// CleanupRootFs cleans up container data storaged in the isulad
func (sd *baseStorageDriver) CleanupRootFs(id string) {
	req := &isula.ContainerRemoveRequest{
		NameId: id,
	}
	// During the rollback, only information is collected
	_, err := sd.imgClient.ContainerRemove(context.Background(), req)
	if err != nil {
		logrus.Warnf("isulad-img remove container %s: %v", id, err)
	} else {
		logrus.Infof("isulad-img remove container %s successful", id)
	}
}

// MountRootFs mounts the rw layer of container
func (sd *baseStorageDriver) MountRootFs(id string) error {
	req := &isula.ContainerMountRequest{
		NameId: id,
	}
	resp, err := sd.imgClient.ContainerMount(context.Background(), req)
	if err != nil {
		return err
	}
	if msg := resp.GetErrmsg(); msg != "" {
		return fmt.Errorf("isulad-img mount failed: %s", msg)
	}
	return nil
}

// UmountRootFs umounts the rw layer of container
func (sd *baseStorageDriver) UmountRootFs(id string) error {
	req := &isula.ContainerUmountRequest{
		NameId: id,
	}
	resp, err := sd.imgClient.ContainerUmount(context.Background(), req)
	if err != nil {
		return err
	}
	if msg := resp.GetErrmsg(); msg != "" {
		req.Force = true
		fResp, fErr := sd.imgClient.ContainerUmount(context.Background(), req)
		logrus.Infof("isulad-img force umount container: %v, err: %v", fResp, fErr)
		if fErr == nil && fResp.GetErrmsg() == "" {
			return nil
		}
		return fmt.Errorf("isulad-img umount failed: %s", msg)
	}
	return nil
}

func dialOpt(ctx context.Context, addr string) (net.Conn, error) {
	// dialer to support unix dial
	dialer := func(addr string, timeout time.Duration) (net.Conn, error) {
		proto, address := "unix", strings.TrimPrefix(addr, "unix://")
		return net.DialTimeout(proto, address, timeout)
	}
	if deadline, ok := ctx.Deadline(); ok {
		return dialer(addr, time.Until(deadline))
	}
	return dialer(addr, isuladImgTimeout)
}

func newIsuladImgClient(addr string, opts ...grpc.DialOption) (isula.ImageServiceClient, error) {
	if addr == "" {
		addr = defaultAddress
	}
	opts = append(opts, grpc.WithInsecure(), grpc.WithContextDialer(dialOpt))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return isula.NewImageServiceClient(conn), nil
}
