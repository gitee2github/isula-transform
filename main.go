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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gopkg.in/natefinch/lumberjack.v2"
	"isula.org/isula-transform/transform"
	_ "isula.org/isula-transform/transform/register"
	"isula.org/isula-transform/utils"
)

const (
	exitNormal = iota
	exitInitErr
	exitTransformErr

	maxConcurrentTransform = 128
	maxPerLogFileSize      = 10 // megabytes
)

var (
	version   string
	gitCommit string
)

func genVersion() string {
	versions := []string{
		"version:" + version,
		"commit:" + gitCommit,
	}
	return strings.Join(versions, "\t")
}

func main() {
	app := &cli.App{
		Name:      "isula-transform",
		Usage:     "transform specify docker container type configuration to iSulad type",
		UsageText: "[global options] --all|container_id[ container_id...]",
		Version:   genVersion(),
		Action:    start,
	}
	for _, v := range transformFlags {
		app.Flags = append(app.Flags, v...)
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitInitErr)
	}
	os.Exit(exitNormal)
}

func start(ctx *cli.Context) error {
	logInit(ctx)
	if err := transformInit(ctx); err != nil {
		return cli.NewExitError(err.Error(), exitInitErr)
	}
	return doTransform(ctx)
}

func logInit(ctx *cli.Context) {
	logPath := ctx.GlobalString("log")
	logRoot := filepath.Dir(logPath)
	if err := os.MkdirAll(logRoot, 0660); err != nil {
		logrus.SetOutput(os.Stdout)
		logrus.Infof("create the log directory %s failed: %v, using STDOUT", logRoot, err)
	} else {
		logrus.SetOutput(&lumberjack.Logger{
			Filename: logPath,
			MaxSize:  maxPerLogFileSize,
			Compress: true,
		})
	}

	logrus.SetLevel(transLogLevel(ctx.GlobalString("log-level")))
}

func transLogLevel(lvl string) logrus.Level {
	switch strings.ToLower(lvl) {
	case "error":
		return logrus.ErrorLevel
	case "warn":
		return logrus.WarnLevel
	case "debug":
		return logrus.DebugLevel
	default:
	}
	return logrus.InfoLevel
}

func transformInit(ctx *cli.Context) error {
	var iSuladCfg = struct {
		Graph         string `json:"graph"`
		State         string `json:"state"`
		Runtime       string `json:"default-runtime"`
		LogLevel      string `json:"log-level"`
		LogDriver     string `json:"log-driver"`
		StorageDriver string `json:"storage-driver"`
		ImageServer   string `json:"image-server-sock-addr"`
	}{}
	iSuladCfgFile := ctx.GlobalString("isulad-config-file")
	if err := utils.CheckFileValid(iSuladCfgFile); err != nil {
		return errors.Wrapf(err, "check isulad daemon config failed")
	}
	iSuladCfgData, err := ioutil.ReadFile(iSuladCfgFile)
	if err != nil {
		logrus.Errorf("read isulad daemon config failed: %v, file path: %s", err, iSuladCfgFile)
		return errors.Wrapf(err, "read isulad daemon config failed")
	}
	err = json.Unmarshal(iSuladCfgData, &iSuladCfg)
	if err != nil {
		logrus.Errorf("unmarshal isulad daemon config failed: %v, file path: %s", err, iSuladCfgFile)
		return errors.Wrapf(err, "unmarshal isulad daemon config failed")
	}

	logrus.Debugf("isulad daemon config: %+v", iSuladCfg)
	err = transform.InitIsuladTool(iSuladCfg.Graph, iSuladCfg.Runtime, iSuladCfg.StorageDriver, iSuladCfg.ImageServer)
	if err != nil {
		return errors.Wrapf(err, "transform init failed")
	}
	if iSuladCfg.LogDriver != "file" {
		logrus.Infof("isula daemon log driver is %s, can't redirect to file", iSuladCfg.LogDriver)
	} else {
		transform.LcrLogInit(iSuladCfg.State, iSuladCfg.Runtime, iSuladCfg.LogLevel)
	}
	return nil
}

func doTransform(ctx *cli.Context) error {
	e := transform.GetTransformer(ctx)
	if e == nil {
		return cli.NewExitError("get transform engine failed", exitInitErr)
	}
	if err := e.Init(); err != nil {
		return cli.NewExitError("transform engine init failed", exitInitErr)
	}

	var ids []string
	all := ctx.GlobalBool("all")
	if !all {
		if ctx.Args().Present() {
			ids = append(ctx.Args().Tail(), ctx.Args().First())
		} else {
			exitMsg := "isula-transform requires at least one container id as an input or setting the --all flag"
			return cli.NewExitError(exitMsg, exitInitErr)
		}
	}

	exitCode := exitNormal
	retCh := make(chan transform.Result, maxConcurrentTransform)
	go e.Transform(ids, all, retCh)
	for ret := range retCh {
		if !ret.Ok {
			exitCode = exitTransformErr
			fmt.Fprintln(os.Stderr, ret.Msg)
		} else {
			fmt.Fprintln(os.Stdout, ret.Msg)
		}
	}
	if exitCode != exitNormal {
		return cli.NewExitError("The transformation has been completed, but at least one failed", exitTransformErr)
	}
	return nil
}
