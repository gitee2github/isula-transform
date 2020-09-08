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

#include "isuladimg.h"

char *util_strdup_s(const char *src)
{
    char *dst = NULL;

    if (src == NULL)
    {
        return NULL;
    }

    dst = strdup(src);
    if (dst == NULL)
    {
        abort();
    }

    return dst;
}

int init_isulad_image_module(char *graph, char *state, char *driver, char **opts, size_t len, int check)
{
    int ret = -1;
    isulad_daemon_configs *conf = NULL;
    char **storage_opts;
    size_t i = 0;

    if (graph == NULL || state == NULL || driver == NULL)
    {
        return -1;
    }

    conf = safe_malloc(sizeof(isulad_daemon_configs));
    conf->graph = util_strdup_s(graph);
    conf->state = util_strdup_s(state);
    conf->storage_driver = util_strdup_s(driver);
    storage_opts = malloc(sizeof(char *) * len);
    for (i = 0; i < len; i++)
    {
        storage_opts[i] = util_strdup_s(opts[i]);
    }
    conf->storage_opts = storage_opts;
    conf->storage_opts_len = len;

    if (check == 1)
    {
        conf->image_layer_check = true;
    }
    else
    {
        conf->image_layer_check = false;
    }

    ret = image_module_init(conf);
    free_isulad_daemon_configs(conf);
    return ret;
}

char *isulad_img_prepare_rootfs(char *type, char *id, char *name)
{
    char *real_rootfs = NULL;
    im_prepare_request *req = NULL;

    if (type == NULL || id == NULL || name == NULL)
    {
        return NULL;
    }

    req = safe_malloc(sizeof(im_prepare_request));
    req->container_id = util_strdup_s(id);
    req->image_type = util_strdup_s(type);
    req->image_name = util_strdup_s(name);
    if (im_prepare_container_rootfs(req, &real_rootfs) != 0)
    {
        real_rootfs = NULL;
    }

    free_im_prepare_request(req);
    return real_rootfs;
}