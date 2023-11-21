# Installation from source

1. Verify that you have Go 1.21+ installed

   ```shell
   $ go version
   ```

   If `go` is not installed, follow instructions on [the Go website](https://golang.org/doc/install).

2. Clone this repository

   ```sh
   $ git clone https://github.com/abc-inc/heimdall.git
   $ cd heimdall
   ```

3. Build and install

   #### Unix-like systems

   ```sh
   # installs to '/usr/local' by default; sudo may be required
   $ make install
   
   # or, install to a different location
   $ make install bindir=~/bin
   ```

   #### Windows

   ```pwsh
   # determine the version
   > git describe --tags
   # or
   > git rev-parse --short HEAD
   # build the `bin\heimdall.exe` binary (replace VERSION with the actual version)
   > go build -trimpath -ldflags "-X main.version=VERSION" -o bin ./...
   ```
   There is no install step available on Windows.

4. Run `heimdall version` to check if it worked.

   #### Windows

   Run `bin\heimdall.exe version` to check if it worked.

## Cross-compiling binaries for different platforms

You can use any platform with Go installed to build a binary that is intended for another platform or CPU architecture.
This is achieved by setting environment variables such as `GOOS` and `GOARCH`.

For example, to compile the `heimdall` binary for the Apple Silicon:

```sh
# on a Unix-like system:
$ GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 make
```

Run `go tool dist list` to list all supported values of `GOOS`/`GOARCH`.

Tip: to reduce the size of the resulting binary, you can use `GO_LDFLAGS="-s -w"`.
This omits symbol tables used for debugging.
See the list of [supported linker flags](https://golang.org/cmd/link/).
