# isula-transform

## Description

`isula-transform` is a tool which converts the configuration of the docker container to a type that isulad can recognize as being loaded.

## Build and Installation

Before building, ensure that [Golang](https://golang.org/) 1.13 or later and [lcr](https://gitee.com/openeuler/lcr) 2.0.1 or later are installed successfully.

Just execute `sudo make && sudo make install` in the code root directory and enjoy it.

## Instructions

Basic usage of isula-transform:

``` txt
NAME:
   isula-transform - transform specify docker container type configuration to iSulad type

USAGE:
   [global options] --all|container_id[ container_id...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log value                 specific output log file path (default: "/var/log/isula-kits/transform.log")
   --log-level value           Customize the level of logging for collection, allowed: debug, info, warn, error (default: "info")
   --isulad-config-file value  iSulad configuration file path (default: "/etc/isulad/daemon.json")
   --docker-graph value        graph root of docker (default: "/var/lib/docker")
   --docker-state value        state root of docker (default: "/var/run/docker")
   --all                       transform all containers
   --help, -h                  show help
   --version, -v               print the version
```

### NOTE

There are a few things to note about using `isula-transform` :

- currently, only docker 18.09 containers are supported to transform to isulad container
- due to isulad's lack of native network capability, docker container needs to configure host network
- `isula-transform` will read the container's OCI configuration, which requires the docker container to be in a pause or running state,  and to be paused if it is in a running state

## Contributions

We always welcome new contributors. And we are happy to provide guidance for the new contributors.

## Licensing

`isula-transform` is licensed under the [Mulan PSL v2](https://license.coscl.org.cn/MulanPSL2/index.html).
