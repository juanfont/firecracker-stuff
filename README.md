# firecracker stuff

Random experiments for Headscale integration tests...

## Building rootfs

https://github.com/bkleiner/ubuntu-firecracker

```bash
dd if=/dev/zero of=rootfs.ext4 bs=1M count=100
mkfs.ext4 rootfs.ext4
mkdir /tmp/my-rootfs
sudo mount rootfs.ext4 /tmp/my-rootfs


```

## Links

- https://jvns.ca/blog/2021/01/22/day-44--got-some-vms-to-start-in-firecracker/

- https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md#getting-the-firecracker-binary
