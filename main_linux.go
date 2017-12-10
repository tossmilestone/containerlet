package main

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
)

var registeredInitializers = make(map[string]func())


func init() {
    initializer, exist := registeredInitializers["containerInitialize"]
    if !exist {
        registeredInitializers["containerInitialize"] = containerInitialize
    }

    if initializer, exist = registeredInitializers[os.Args[0]]; exist {
        initializer()
        os.Exit(0)
    }
}

func containerInitialize() {
    containerMain(os.Args[1:]...)
}

func main() {
    if len(os.Args) < 2 {
        usage()
        return
    }
    runContainer(os.Args[1:]...)
}

func usage() {
    fmt.Println("Usage: containerlet <cmd>")
}

func containerMain(args ...string) {
    fmt.Printf("Enter container[%d]\n", os.Getpid())
    hostname, _ := syscall.ByteSliceFromString("container")
    syscall.Sethostname(hostname)

    new_root := "/tmp/rootfs"
    putold := "/tmp/rootfs/old_root"

    // Mount other file systems
    proc := fmt.Sprintf("%s/proc", new_root)
    must(syscall.Mount("proc", proc, "proc", 0, ""))

    // Mount custom config
    if _, err := os.Stat("conf/hosts"); err == nil {
        hosts := fmt.Sprintf("%s/etc/hosts", new_root)
        must(syscall.Mount("conf/hosts", hosts, "none", syscall.MS_BIND, ""))
    }
    if _, err := os.Stat("conf/hostname"); err == nil {
        hostname_path := fmt.Sprintf("%s/etc/hostname", new_root)
        must(syscall.Mount("conf/hostname", hostname_path, "none", syscall.MS_BIND, ""))
    }
    if _, err := os.Stat("conf/resolv.conf"); err == nil {
        resolv := fmt.Sprintf("%s/etc/resolv.conf", new_root)
        must(syscall.Mount("conf/resolv.conf", resolv, "none", syscall.MS_BIND, ""))
    }

    // Pivot root
    must(syscall.Mount(new_root, new_root, "", syscall.MS_BIND|syscall.MS_REC, ""))
    must(os.Mkdir(putold, 0700))
    defer func() {
        must(os.RemoveAll(putold))
    }()
    must(syscall.PivotRoot(new_root, putold))
    must(os.Chdir("/"))
    must(syscall.Unmount("/old_root", syscall.MNT_DETACH))
    must(os.RemoveAll("/old_root"))

    cmd := exec.Command(args[0], args[1:]...)

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        fmt.Println("error", err)
        os.Exit(1)
    }

    fmt.Println("Exit container")
}

func runContainer(args ...string)  {
    fmt.Printf("Parent[%d] start container\n", os.Getpid())
    containerArgs := append([] string{"containerInitialize"}, args...)
    cmd := &exec.Cmd{
        Path: "/proc/self/exe",
        Args: containerArgs,
    }
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER |
            syscall.CLONE_NEWIPC,
        UidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: 0,
                HostID:      os.Getuid(),
                Size:        1,
            },
        },
        GidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: 0,
                HostID:      os.Getgid(),
                Size:        1,
            },
        },
    }
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        fmt.Println("error", err)
        os.Exit(1)
    }
    fmt.Println("Exit parent")
}

func must(err error) {
    if err != nil {
        panic(err)
    }
}
