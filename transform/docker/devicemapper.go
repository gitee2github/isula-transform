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
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"isula.org/isula-transform/transform"
	"isula.org/isula-transform/types"
)

const (
	changeItem = iota
	addItem
	delItem
	ignoreItem
)

type diffNode struct {
	// parent    *diffNode
	children  map[string]*diffNode
	operation uint8
	path      string
}

type diffTrie struct {
	root *diffNode
}

func newTrie() *diffTrie {
	return &diffTrie{
		root: &diffNode{
			operation: ignoreItem,
			path:      "/",
			children:  make(map[string]*diffNode),
		},
	}
}

func (t *diffTrie) insert(path string, operation uint8) {
	path = strings.TrimPrefix(path, "/")
	pathItems := strings.Split(path, "/")
	curNode := t.root
	for len(pathItems) > 0 {
		curPath := pathItems[0]
		node, ok := curNode.children[curPath]
		if !ok {
			node = &diffNode{
				operation: operation,
				path:      filepath.Join(curNode.path, curPath),
				children:  make(map[string]*diffNode),
			}
			curNode.children[curPath] = node
		} else {
			node.operation = ignoreItem
		}
		curNode = node
		pathItems = pathItems[1:]
	}
}

func (t *diffTrie) filter() []container.ContainerChangeResponseItem {
	var stack []*diffNode
	var diffs []container.ContainerChangeResponseItem
	node := t.root
	for node == nil {
		return nil
	}
	stack = append(stack, node)
	for len(stack) > 0 {
		node := stack[0]
		if node.operation != ignoreItem {
			diffs = append(diffs, container.ContainerChangeResponseItem{
				Kind: node.operation,
				Path: node.path,
			})
		}
		for _, v := range node.children {
			stack = append(stack, v)
		}
		stack = stack[1:]
	}
	return diffs
}

type deviceMapperDriver struct {
	transform.BaseStorageDriver
	client dockerClient
}

func newDeviceMapperDriver(base transform.BaseStorageDriver, client dockerClient) transform.StorageDriver {
	return &deviceMapperDriver{BaseStorageDriver: base, client: client}
}

func (dm *deviceMapperDriver) GenerateRootFs(id, image string) (string, error) {
	return dm.BaseStorageDriver.GenerateRootFs(id, image)
}

func (dm *deviceMapperDriver) TransformRWLayer(ctr *types.IsuladV2Config, oldRootFs string) error {
	// mount
	if err := dm.BaseStorageDriver.MountRootFs(ctr.CommonConfig.ID); err != nil {
		return err
	}
	// umount
	defer func() {
		if err := dm.BaseStorageDriver.UmountRootFs(ctr.CommonConfig.ID); err != nil {
			logrus.Infof("device mapper umount rootfs failed: %v", err)
		}
	}()

	// get diff
	diff, err := dm.client.ContainerDiff(context.Background(), ctr.CommonConfig.ID)
	if err != nil {
		return err
	}
	changes := dm.changesFilter(diff, ctr.CommonConfig.MountPoints)
	logrus.Infof("device mapper driver get diff form docker: %+v, filter: %+v", diff, changes)
	for idx := range changes {
		src := oldRootFs + changes[idx].Path
		dest := ctr.CommonConfig.BaseFs + changes[idx].Path
		switch changes[idx].Kind {
		case addItem, changeItem:
			// cp
			destRoot := filepath.Dir(dest)
			if err := exec.Command("cp", "-ra", src, destRoot).Run(); err != nil {
				logrus.Errorf("device mapper copy %s to %s filed: %v", src, dest, err)
				return err
			}
		case delItem:
			// delete
			if err := os.RemoveAll(dest); err != nil {
				logrus.Errorf("device mapper remover %s filed: %v", dest, err)
				return err
			}
		default:
		}
	}
	return nil
}

/*
A /xxx/.../something ==> root dir /xxx  C
D /xxx/.../something ==> root dir /xxx C
C /xxx/.../something ==> root dir /xxx C
1. We ignore the parent folder node /xxx/..., so if the properties of the folder
   are modified at the same time, the change is lost.
2. only if the node didn't have any children nodes, we copy or delete it. In
   another words, We only deal with leaf nodes.

note: filter path which match bind mount in container
*/
func (dm *deviceMapperDriver) changesFilter(changes []container.ContainerChangeResponseItem,
	mounts map[string]types.Mount) []container.ContainerChangeResponseItem {
	// create trie
	t := newTrie()
	for _, change := range changes {
		if _, ok := mounts[change.Path]; ok {
			continue
		}
		t.insert(change.Path, change.Kind)
	}
	return t.filter()
}

func (dm *deviceMapperDriver) Cleanup(id string) {
	dm.BaseStorageDriver.CleanupRootFs(id)
}
