# kit

## About kit

> Common packages and toolkits for golang

## Environment Requirements

These environments and tools must be installed properly.

- [go](https://golang.org/dl/)
- [make](https://docs.gitea.io/zh-cn/make/)

if `cc1.exe: sorry, unimplemented: 64-bit mode not compiled in` error occur on Windows, it is recommended to
use [MinGW-w64 ](https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z/download)

The `GO111MODULE` should be enabled.

```bash
go env -w GO111MODULE=on
```

If you faced with network problem (especially you are in China Mainland), please [setup GOPROXY](https://goproxy.cn/)
