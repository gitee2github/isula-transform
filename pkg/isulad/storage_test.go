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
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"isula.org/isula-transform/api/isula"
)

const (
	bsdTestCtrID    = "isulatransformbsdtestcontainer"
	bsdTestCtrImage = "isulatransformbsdtestimage"
)

func TestBaseStorageDriverCreate(t *testing.T) {
	ic := new(mockImgClient)
	sd := isuladStorageDriver{imgClient: ic}
	ic.remove = func(ctx context.Context, in *isula.ContainerRemoveRequest, opts ...grpc.CallOption) (*isula.ContainerRemoveResponse, error) {
		return &isula.ContainerRemoveResponse{Errmsg: "remove no error"}, nil
	}
	Convey("TestBaseStorageDriver_GenerateRootFs", t, func() {
		Convey("prepare return err", func() {
			ic.prepare = func(ctx context.Context, in *isula.ContainerPrepareRequest, opts ...grpc.CallOption) (*isula.ContainerPrepareResponse, error) {
				return nil, fmt.Errorf("err: image client prepare failed")
			}
			rootfs, err := sd.GenerateRootFs(bsdTestCtrID, bsdTestCtrImage)
			So(err, ShouldBeError)
			So(rootfs, ShouldBeBlank)
		})

		Convey("prepare return err msg", func() {
			ic.prepare = func(ctx context.Context, in *isula.ContainerPrepareRequest, opts ...grpc.CallOption) (*isula.ContainerPrepareResponse, error) {
				return &isula.ContainerPrepareResponse{Errmsg: "errMsg: image client prepare failed"}, nil
			}
			rootfs, err := sd.GenerateRootFs(bsdTestCtrID, bsdTestCtrImage)
			So(err, ShouldBeError)
			So(rootfs, ShouldBeBlank)
		})

		Convey("successfully prepare", func() {
			testRootPath := "/test/rootfs"
			ic.prepare = func(ctx context.Context, in *isula.ContainerPrepareRequest, opts ...grpc.CallOption) (*isula.ContainerPrepareResponse, error) {
				return &isula.ContainerPrepareResponse{MountPoint: testRootPath}, nil
			}
			rootfs, err := sd.GenerateRootFs(bsdTestCtrID, bsdTestCtrImage)
			So(err, ShouldBeNil)
			So(rootfs, ShouldEqual, testRootPath)
		})
	})

	Convey("TestBaseStorageDriver_CleanupRootFs", t, func() {
		sd.CleanupRootFs(bsdTestCtrID)
	})
}

func TestBaseStorageDriver_MountRootFs(t *testing.T) {
	ic := new(mockImgClient)
	sd := isuladStorageDriver{imgClient: ic}
	Convey("TestBaseStorageDriver_MountRootFs", t, func() {
		Convey("mount return err", func() {
			errContent := "err: image client mount failed"
			ic.mount = func(ctx context.Context, in *isula.ContainerMountRequest, opts ...grpc.CallOption) (*isula.ContainerMountResponse, error) {
				return nil, fmt.Errorf(errContent)
			}
			err := sd.MountRootFs(bsdTestCtrID)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, errContent)
		})

		Convey("mount return err msg", func() {
			errMsgContent := "errMsg: image client mount failed"
			ic.mount = func(ctx context.Context, in *isula.ContainerMountRequest, opts ...grpc.CallOption) (*isula.ContainerMountResponse, error) {
				return &isula.ContainerMountResponse{Errmsg: errMsgContent}, nil
			}
			err := sd.MountRootFs(bsdTestCtrID)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, errMsgContent)
		})

		Convey("mount successfully", func() {
			ic.mount = func(ctx context.Context, in *isula.ContainerMountRequest, opts ...grpc.CallOption) (*isula.ContainerMountResponse, error) {
				return nil, nil
			}
			err := sd.MountRootFs(bsdTestCtrID)
			So(err, ShouldBeNil)
		})
	})
}

func TestBaseStorageDriver_UmountRootFs(t *testing.T) {
	ic := new(mockImgClient)
	sd := isuladStorageDriver{imgClient: ic}
	Convey("TestBaseStorageDriver_UmountRootFs", t, func() {
		Convey("umount return err", func() {
			errContent := "err: image client umount failed"
			ic.umount = func(ctx context.Context, in *isula.ContainerUmountRequest, opts ...grpc.CallOption) (*isula.ContainerUmountResponse, error) {
				return nil, fmt.Errorf(errContent)
			}
			err := sd.UmountRootFs(bsdTestCtrID)
			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, errContent)
		})

		Convey("umount return err msg", func() {
			errMsgContent := "errMsg: image client umount failed"

			Convey("force umount successfully", func() {
				ic.umount = func(ctx context.Context, in *isula.ContainerUmountRequest, opts ...grpc.CallOption) (*isula.ContainerUmountResponse, error) {
					if in.Force {
						return nil, nil
					}
					return &isula.ContainerUmountResponse{Errmsg: errMsgContent}, nil
				}
				err := sd.UmountRootFs(bsdTestCtrID)
				So(err, ShouldBeNil)
			})

			Convey("force umount return err", func() {
				ic.umount = func(ctx context.Context, in *isula.ContainerUmountRequest, opts ...grpc.CallOption) (*isula.ContainerUmountResponse, error) {
					return &isula.ContainerUmountResponse{Errmsg: errMsgContent}, nil
				}
				err := sd.UmountRootFs(bsdTestCtrID)
				So(err, ShouldBeError)
				So(err.Error(), ShouldContainSubstring, errMsgContent)
			})
		})
	})
}

type mockImgClient struct {
	prepare func(ctx context.Context, in *isula.ContainerPrepareRequest, opts ...grpc.CallOption) (*isula.ContainerPrepareResponse, error)
	remove  func(ctx context.Context, in *isula.ContainerRemoveRequest, opts ...grpc.CallOption) (*isula.ContainerRemoveResponse, error)
	mount   func(ctx context.Context, in *isula.ContainerMountRequest, opts ...grpc.CallOption) (*isula.ContainerMountResponse, error)
	umount  func(ctx context.Context, in *isula.ContainerUmountRequest, opts ...grpc.CallOption) (*isula.ContainerUmountResponse, error)
}

// create rootfs for container
func (mic *mockImgClient) ContainerPrepare(ctx context.Context, in *isula.ContainerPrepareRequest, opts ...grpc.CallOption) (*isula.ContainerPrepareResponse, error) {
	return mic.prepare(ctx, in, opts...)
}

// remove rootfs of container
func (mic *mockImgClient) ContainerRemove(ctx context.Context, in *isula.ContainerRemoveRequest, opts ...grpc.CallOption) (*isula.ContainerRemoveResponse, error) {
	return mic.remove(ctx, in, opts...)
}

// mount rwlayer for container
func (mic *mockImgClient) ContainerMount(ctx context.Context, in *isula.ContainerMountRequest, opts ...grpc.CallOption) (*isula.ContainerMountResponse, error) {
	return mic.mount(ctx, in, opts...)
}

// umount rwlayer of container
func (mic *mockImgClient) ContainerUmount(ctx context.Context, in *isula.ContainerUmountRequest, opts ...grpc.CallOption) (*isula.ContainerUmountResponse, error) {
	return mic.umount(ctx, in, opts...)
}

// ListImages lists existing images.
func (mic *mockImgClient) ListImages(ctx context.Context, in *isula.ListImagesRequest, opts ...grpc.CallOption) (*isula.ListImagesResponse, error) {
	panic("not implemented") // TODO: Implement
}

// ImageStatus returns the status of the image. If the image is not
// present, returns a response with ImageStatusResponse.Image set to
// nil.
func (mic *mockImgClient) ImageStatus(ctx context.Context, in *isula.ImageStatusRequest, opts ...grpc.CallOption) (*isula.ImageStatusResponse, error) {
	panic("not implemented") // TODO: Implement
}

//  Get image information
func (mic *mockImgClient) ImageInfo(ctx context.Context, in *isula.ImageInfoRequest, opts ...grpc.CallOption) (*isula.ImageInfoResponse, error) {
	panic("not implemented") // TODO: Implement
}

// PullImage pulls an image with authentication config.
func (mic *mockImgClient) PullImage(ctx context.Context, in *isula.PullImageRequest, opts ...grpc.CallOption) (*isula.PullImageResponse, error) {
	panic("not implemented") // TODO: Implement
}

// RemoveImage removes the image.
// This call is idempotent, and must not return an error if the image has
// already been removed.
func (mic *mockImgClient) RemoveImage(ctx context.Context, in *isula.RemoveImageRequest, opts ...grpc.CallOption) (*isula.RemoveImageResponse, error) {
	panic("not implemented") // TODO: Implement
}

// ImageFSInfo returns information of the filesystem that is used to store images.
func (mic *mockImgClient) ImageFsInfo(ctx context.Context, in *isula.ImageFsInfoRequest, opts ...grpc.CallOption) (*isula.ImageFsInfoResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Load image from file
func (mic *mockImgClient) LoadImage(ctx context.Context, in *isula.LoadImageRequest, opts ...grpc.CallOption) (*isula.LoadImageResponose, error) {
	panic("not implemented") // TODO: Implement
}

// Import rootfs to be image
func (mic *mockImgClient) Import(ctx context.Context, in *isula.ImportRequest, opts ...grpc.CallOption) (*isula.ImportResponose, error) {
	panic("not implemented") // TODO: Implement
}

// isulad image services
// get all Container rootfs
func (mic *mockImgClient) ListContainers(ctx context.Context, in *isula.ListContainersRequest, opts ...grpc.CallOption) (*isula.ListContainersResponse, error) {
	panic("not implemented") // TODO: Implement
}

// export container rootfs
func (mic *mockImgClient) ContainerExport(ctx context.Context, in *isula.ContainerExportRequest, opts ...grpc.CallOption) (*isula.ContainerExportResponse, error) {
	panic("not implemented") // TODO: Implement
}

// get filesystem usage of container
func (mic *mockImgClient) ContainerFsUsage(ctx context.Context, in *isula.ContainerFsUsageRequest, opts ...grpc.CallOption) (*isula.ContainerFsUsageResponse, error) {
	panic("not implemented") // TODO: Implement
}

// get status of graphdriver
func (mic *mockImgClient) GraphdriverStatus(ctx context.Context, in *isula.GraphdriverStatusRequest, opts ...grpc.CallOption) (*isula.GraphdriverStatusResponse, error) {
	panic("not implemented") // TODO: Implement
}

// get metadata of graphdriver
func (mic *mockImgClient) GraphdriverMetadata(ctx context.Context, in *isula.GraphdriverMetadataRequest, opts ...grpc.CallOption) (*isula.GraphdriverMetadataResponse, error) {
	panic("not implemented") // TODO: Implement
}

// login registry
func (mic *mockImgClient) Login(ctx context.Context, in *isula.LoginRequest, opts ...grpc.CallOption) (*isula.LoginResponse, error) {
	panic("not implemented") // TODO: Implement
}

// logout registry
func (mic *mockImgClient) Logout(ctx context.Context, in *isula.LogoutRequest, opts ...grpc.CallOption) (*isula.LogoutResponse, error) {
	panic("not implemented") // TODO: Implement
}

// health check service
func (mic *mockImgClient) HealthCheck(ctx context.Context, in *isula.HealthCheckRequest, opts ...grpc.CallOption) (*isula.HealthCheckResponse, error) {
	panic("not implemented") // TODO: Implement
}

// Add a tag to the image
func (mic *mockImgClient) TagImage(ctx context.Context, in *isula.TagImageRequest, opts ...grpc.CallOption) (*isula.TagImageResponse, error) {
	panic("not implemented") // TODO: Implement
}
