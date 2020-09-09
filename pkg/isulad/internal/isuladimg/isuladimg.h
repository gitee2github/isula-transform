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

#include <isulad/image_api.h>

extern int init_isulad_image_module(char *graph, char *state, char *driver, char **opts, size_t len, int check);
extern char *isulad_img_prepare_rootfs(char *type, char *id, char *name);