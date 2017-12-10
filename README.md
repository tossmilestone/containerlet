## A simplest container implemented in golang

`containerlet` is a simplest container implementd in golang, it could run command in container with basic isolations, one can learn how container is implemeted from this project. Currently it only supports to run in linux os.

### Install

```
go install
```

The default `rootfs` directory is in `/tmp/rootfs`, you need to populate the directory with `rootfs` files.

### Usage

```
# containerlet <cmd>
```

`cmd` is the command you want to run in the container.
