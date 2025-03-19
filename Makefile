export CGO_ENABLED ?= 0
GOFLAGS += -buildvcs -trimpath
LDFLAGS += -X main.version=$(VERSION)
INSTALL ?= install
INSTALL_PROGRAM ?= $(INSTALL)

prefix = /usr/local
bindir ?= $(prefix)/bin

builddir ?= bin
distdir ?= dist
tmpdir ?= tmp

all: test build

build: generate
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$(builddir)/" ./cmd/...

check:
	golangci-lint run ./...

clean:
	find . -type f -name '*_gen.go' -delete
	rm -rf "$(builddir)" "$(distdir)" "$(tmpdir)"

dist:
	$(MAKE) bindir="$(distdir)/$(notdir $(CURDIR))" install
	tar -C $(distdir) -cvf "$(distdir)/$(notdir $(CURDIR)).tar.gz" "$(notdir $(CURDIR))"

generate:
	@mkdir -p "$(builddir)"
	GOOS= GOARCH= go generate -x ./...

install: all
	@$(INSTALL_PROGRAM) -d "$(DESTDIR)$(bindir)"
	@$(INSTALL_PROGRAM) -v "$(builddir)"/* "$(DESTDIR)$(bindir)"

test:
	@mkdir -p "$(tmpdir)/reports"
	go test $(GOFLAGS) -ldflags "$(LDFLAGS)" -coverprofile "$(tmpdir)/reports/coverage.out" ./...
	go tool cover -html "$(tmpdir)/reports/coverage.out" -o "$(tmpdir)/reports/coverage.html"
	go tool cover -func "$(tmpdir)/reports/coverage.out" -o "$(tmpdir)/reports/coverage.txt"

uninstall:
	rm -fv "$(bindir)/$(notdir $(CURDIR))"

.PHONY: all build check clean dist install test uninstall
