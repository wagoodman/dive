package export

import (
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/joschi/dive/dive/image/docker"
)

func Test_Export(t *testing.T) {
	result := docker.TestAnalysisFromArchive(t, "../../.data/test-docker-image.tar")

	export := NewExport(result)
	payload, err := export.Marshal()
	if err != nil {
		t.Errorf("Test_Export: unable to export analysis: %v", err)
	}

	expectedResult := `{
  "layer": [
    {
      "index": 0,
      "id": "28cfe03618aa2e914e81fdd90345245c15f4478e35252c06ca52d238fd3cc694",
      "digestId": "sha256:23bc2b70b2014dec0ac22f27bb93e9babd08cdd6f1115d0c955b9ff22b382f5a",
      "sizeBytes": 1154361,
      "command": "#(nop) ADD file:ce026b62356eec3ad1214f92be2c9dc063fe205bd5e600be3492c4dfb17148bd in / ",
      "fileList": [
        {
          "path": "bin/[",
          "typeFlag": 48,
          "linkName": "",
          "size": 1075464,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/[[",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/acpid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/add-shell",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/addgroup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/adduser",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/adjtimex",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ar",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/arch",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/arp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/arping",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ash",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/awk",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/base64",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/basename",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/beep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/blkdiscard",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/blkid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/blockdev",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/bootchartd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/brctl",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/bunzip2",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/busybox",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/bzcat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/bzip2",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cal",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chattr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chgrp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chmod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chown",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chpasswd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chpst",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chroot",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chrt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/chvt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cksum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/clear",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cmp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/comm",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/conspy",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cpio",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/crond",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/crontab",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cryptpw",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cttyhack",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/cut",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/date",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/deallocvt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/delgroup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/deluser",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/depmod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/devmem",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/df",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dhcprelay",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/diff",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dirname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dmesg",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dnsd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dnsdomainname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dos2unix",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dpkg",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dpkg-deb",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/du",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dumpkmap",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/dumpleases",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/echo",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ed",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/egrep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/eject",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/env",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/envdir",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/envuidgid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ether-wake",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/expand",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/expr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/factor",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fakeidentd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fallocate",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/false",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fatattr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fbset",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fbsplash",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fdflush",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fdformat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fdisk",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fgconsole",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fgrep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/find",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/findfs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/flock",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fold",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/free",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/freeramdisk",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fsck",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fsck.minix",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fsfreeze",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fstrim",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fsync",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ftpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ftpget",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ftpput",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/fuser",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/getconf",
          "typeFlag": 48,
          "linkName": "",
          "size": 77880,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/getopt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/getty",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/grep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/groups",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/gunzip",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/gzip",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/halt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hdparm",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/head",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hexdump",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hexedit",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hostid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hostname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/httpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hush",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/hwclock",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/i2cdetect",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/i2cdump",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/i2cget",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/i2cset",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/id",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ifconfig",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ifdown",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ifenslave",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ifplugd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ifup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/inetd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/init",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/insmod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/install",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ionice",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/iostat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ip",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ipaddr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ipcalc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ipcrm",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ipcs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/iplink",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ipneigh",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/iproute",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/iprule",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/iptunnel",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/kbd_mode",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/kill",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/killall",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/killall5",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/klogd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/last",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/less",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/link",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/linux32",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/linux64",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/linuxrc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ln",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/loadfont",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/loadkmap",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/logger",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/login",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/logname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/logread",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/losetup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lpq",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lpr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ls",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lsattr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lsmod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lsof",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lspci",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lsscsi",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lsusb",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lzcat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lzma",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/lzop",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/makedevs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/makemime",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/man",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/md5sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mdev",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mesg",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/microcom",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkdir",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkdosfs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mke2fs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkfifo",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkfs.ext2",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkfs.minix",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkfs.vfat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mknod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkpasswd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mkswap",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mktemp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/modinfo",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/modprobe",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/more",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mount",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mountpoint",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mpstat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/mv",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nameif",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nanddump",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nandwrite",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nbd-client",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/netstat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nice",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nl",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nmeter",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nohup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nproc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nsenter",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nslookup",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ntpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/nuke",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/od",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/openvt",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/partprobe",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/passwd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/paste",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/patch",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pgrep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pidof",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ping",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ping6",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pipe_progress",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pivot_root",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pkill",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pmap",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/popmaildir",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/poweroff",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/powertop",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/printenv",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/printf",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ps",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pscan",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pstree",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pwd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/pwdx",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/raidautorun",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rdate",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rdev",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/readahead",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/readlink",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/readprofile",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/realpath",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/reboot",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/reformime",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/remove-shell",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/renice",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/reset",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/resize",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/resume",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rev",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rm",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rmdir",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rmmod",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/route",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rpm",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rpm2cpio",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rtcwake",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/run-init",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/run-parts",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/runlevel",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/runsv",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/runsvdir",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/rx",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/script",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/scriptreplay",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sed",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sendmail",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/seq",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setarch",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setconsole",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setfattr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setfont",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setkeycodes",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setlogcons",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setpriv",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setserial",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setsid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/setuidgid",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sh",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sha1sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sha256sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sha3sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sha512sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/showkey",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/shred",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/shuf",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/slattach",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sleep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/smemcap",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/softlimit",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sort",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/split",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ssl_client",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/start-stop-daemon",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/stat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/strings",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/stty",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/su",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sulogin",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sum",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sv",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/svc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/svlogd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/svok",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/swapoff",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/swapon",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/switch_root",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sync",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/sysctl",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/syslogd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tac",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tail",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tar",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/taskset",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tcpsvd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tee",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/telnet",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/telnetd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/test",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tftp",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tftpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/time",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/timeout",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/top",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/touch",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tr",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/traceroute",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/traceroute6",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/true",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/truncate",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tty",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ttysize",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/tunctl",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubiattach",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubidetach",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubimkvol",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubirename",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubirmvol",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubirsvol",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/ubiupdatevol",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/udhcpc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/udhcpd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/udpsvd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uevent",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/umount",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unexpand",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uniq",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unix2dos",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unlink",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unlzma",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unshare",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unxz",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/unzip",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uptime",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/users",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/usleep",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uudecode",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/uuencode",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/vconfig",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/vi",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/vlock",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/volname",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/w",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/wall",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/watch",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/watchdog",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/wc",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/wget",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/which",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/who",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/whoami",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/whois",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/xargs",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/xxd",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/xz",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/xzcat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/yes",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/zcat",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin/zcip",
          "typeFlag": 49,
          "linkName": "bin/[",
          "size": 0,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "bin",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "dev",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/group",
          "typeFlag": 48,
          "linkName": "",
          "size": 307,
          "fileMode": 436,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "etc/localtime",
          "typeFlag": 48,
          "linkName": "",
          "size": 127,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "etc/network/if-down.d",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/network/if-post-down.d",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/network/if-pre-up.d",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/network/if-up.d",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/network",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "etc/passwd",
          "typeFlag": 48,
          "linkName": "",
          "size": 340,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "etc/shadow",
          "typeFlag": 48,
          "linkName": "",
          "size": 243,
          "fileMode": 384,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "etc",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "home",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 65534,
          "gid": 65534,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "tmp",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2148532735,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "usr/sbin",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 1,
          "gid": 1,
          "isDir": true
        },
        {
          "path": "usr",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "var/spool/mail",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 8,
          "gid": 8,
          "isDir": true
        },
        {
          "path": "var/spool",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "var/www",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "var",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 1,
      "id": "1871059774abe6914075e4a919b778fa1561f577d620ae52438a9635e6241936",
      "digestId": "sha256:a65b7d7ac139a0e4337bc3c73ce511f937d6140ef61a0108f7d4b8aab8d67274",
      "sizeBytes": 6405,
      "command": "#(nop) ADD file:139c3708fb6261126453e34483abd8bf7b26ed16d952fd976994d68e72d93be2 in /somefile.txt ",
      "fileList": [
        {
          "path": "somefile.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 436,
          "uid": 0,
          "gid": 0,
          "isDir": false
        }
      ]
    },
    {
      "index": 2,
      "id": "49fe2a475548bfa4d493fc796fce41f30704e3d4cbff3e45dd3e06f463236d1d",
      "digestId": "sha256:93e208d471756ffbac88cf9c25feb442007f221d3bd73231e27b747a0a68927c",
      "sizeBytes": 0,
      "command": "mkdir -p /root/example/really/nested",
      "fileList": [
        {
          "path": "root/example/really/nested",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root/example/really",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 3,
      "id": "80cd2ca1ffc89962b9349c80280c2bc551acbd11e09b16badb0669f8e2369020",
      "digestId": "sha256:4abad3abe3cb99ad7a492a9d9f6b3d66287c1646843c74128bbbec4f7be5aa9e",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile1.txt",
      "fileList": [
        {
          "path": "root/example/somefile1.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 4,
      "id": "c99e2f8d3f6282668f0d30dc1db5e67a51d7a1dcd7ff6ddfa0f90760836778ec",
      "digestId": "sha256:14c9a6ffcb6a0f32d1035f97373b19608e2d307961d8be156321c3f1c1504cbf",
      "sizeBytes": 6405,
      "command": "chmod 444 /root/example/somefile1.txt",
      "fileList": [
        {
          "path": "root/example/somefile1.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 292,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 5,
      "id": "5eca617bdc3bc06134fe957a30da4c57adb7c340a6d749c8edc4c15861c928d7",
      "digestId": "sha256:778fb5770ef466f314e79cc9dc418eba76bfc0a64491ce7b167b76aa52c736c4",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile2.txt",
      "fileList": [
        {
          "path": "root/example/somefile2.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 6,
      "id": "f07c3eb887572395408f8e11a07af945e4da5f02b3188bb06b93fad713ca0b99",
      "digestId": "sha256:f275b8a31a71deb521cc048e6021e2ff6fa52bedb25c9b7bbe129a0195ddca5f",
      "sizeBytes": 6405,
      "command": "cp /somefile.txt /root/example/somefile3.txt",
      "fileList": [
        {
          "path": "root/example/somefile3.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 7,
      "id": "461885fc22589158dee3c5b9f01cc41c87805439f58b4399d733b51aa305cbf9",
      "digestId": "sha256:dd1effc5eb19894c3e9b57411c98dd1cf30fa1de4253c7fae53c9cea67267d83",
      "sizeBytes": 6405,
      "command": "mv /root/example/somefile3.txt /root/saved.txt",
      "fileList": [
        {
          "path": "root/example/.wh.somefile3.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 0,
          "fileMode": 0,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/example",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root/saved.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 8,
      "id": "a10327f68ffed4afcba78919052809a8f774978a6b87fc117d39c53c4842f72c",
      "digestId": "sha256:8d1869a0a066cdd12e48d648222866e77b5e2814f773bb3bd8774ab4052f0f1d",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.saved.txt",
      "fileList": [
        {
          "path": "root/.saved.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 9,
      "id": "f2fc54e25cb7966dc9732ec671a77a1c5c104e732bd15ad44a2dc1ac42368f84",
      "digestId": "sha256:bc2e36423fa31a97223fd421f22c35466220fa160769abf697b8eb58c896b468",
      "sizeBytes": 0,
      "command": "rm -rf /root/example/",
      "fileList": [
        {
          "path": "root/.wh.example",
          "typeFlag": 48,
          "linkName": "",
          "size": 0,
          "fileMode": 0,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 10,
      "id": "aad36d0b05e71c7e6d4dfe0ca9ed6be89e2e0d8995dafe83438299a314e91071",
      "digestId": "sha256:7f648d45ee7b6de2292162fba498b66cbaaf181da9004fcceef824c72dbae445",
      "sizeBytes": 2187,
      "command": "#(nop) ADD dir:7ec14b81316baa1a31c38c97686a8f030c98cba2035c968412749e33e0c4427e in /root/.data/ ",
      "fileList": [
        {
          "path": "root/.data/tag.sh",
          "typeFlag": 48,
          "linkName": "",
          "size": 917,
          "fileMode": 509,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/.data/test.sh",
          "typeFlag": 48,
          "linkName": "",
          "size": 1270,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/.data",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 11,
      "id": "3d4ad907517a021d86a4102d2764ad2161e4818bbd144e41d019bfc955434181",
      "digestId": "sha256:a4b8f95f266d5c063c9a9473c45f2f85ddc183e37941b5e6b6b9d3c00e8e0457",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /tmp/saved.again1.txt",
      "fileList": [
        {
          "path": "tmp/saved.again1.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "tmp",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2148532735,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 12,
      "id": "81b1b002d4b4c1325a9cad9990b5277e7f29f79e0f24582344c0891178f95905",
      "digestId": "sha256:22a44d45780a541e593a8862d80f3e14cb80b6bf76aa42ce68dc207a35bf3a4a",
      "sizeBytes": 6405,
      "command": "cp /root/saved.txt /root/.data/saved.again2.txt",
      "fileList": [
        {
          "path": "root/.data/saved.again2.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 420,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root/.data",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484141,
          "uid": 0,
          "gid": 0,
          "isDir": true
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    },
    {
      "index": 13,
      "id": "cfb35bb5c127d848739be5ca726057e6e2c77b2849f588e7aebb642c0d3d4b7b",
      "digestId": "sha256:ba689cac6a98c92d121fa5c9716a1bab526b8bb1fd6d43625c575b79e97300c5",
      "sizeBytes": 6405,
      "command": "chmod +x /root/saved.txt",
      "fileList": [
        {
          "path": "root/saved.txt",
          "typeFlag": 48,
          "linkName": "",
          "size": 6405,
          "fileMode": 493,
          "uid": 0,
          "gid": 0,
          "isDir": false
        },
        {
          "path": "root",
          "typeFlag": 53,
          "linkName": "",
          "size": 0,
          "fileMode": 2147484096,
          "uid": 0,
          "gid": 0,
          "isDir": true
        }
      ]
    }
  ],
  "image": {
    "sizeBytes": 1220598,
    "inefficientBytes": 32025,
    "efficiencyScore": 0.9844212134184309,
    "fileReference": [
      {
        "count": 2,
        "sizeBytes": 12810,
        "file": "/root/saved.txt"
      },
      {
        "count": 2,
        "sizeBytes": 12810,
        "file": "/root/example/somefile1.txt"
      },
      {
        "count": 2,
        "sizeBytes": 6405,
        "file": "/root/example/somefile3.txt"
      }
    ]
  }
}`

	actualResult := string(payload)
	if expectedResult != actualResult {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expectedResult, actualResult, false)

		t.Errorf("Test_Export: unexpected export result:\n%v", dmp.DiffPrettyText(diffs))
	}
}
