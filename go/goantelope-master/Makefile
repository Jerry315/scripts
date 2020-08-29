PREFIX=.
DESTDIR=
GOFLAGS=
BINDIR=${PREFIX}/bin
COVDIR=${PREFIX}/cov

OS =
ARCH = amd64

CLIS = 
LIBS = 
all: lint $(LIBS) $(CLIS) 

$(BINDIR)/%:
	@mkdir -p $(dir $@)
	gox -verbose \
		-os "$(OS)" -arch "$(ARCH)" \
		${GOFLAGS} \
		-output "$@-{{.OS}}-{{.Arch}}" ./cmd/$*

$(CLIS): %: $(BINDIR)/%

.PHONY: tools
tools:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/kisielk/errcheck
	go get github.com/golang/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/matm/gocov-html
	go get github.com/tools/godep
	go get github.com/mitchellh/gox
	go get github.com/stretchr/testify

.PHONY: lint
lint:
	for pkg in $$(go list ./... | grep -v /vendor/) ; do \
		golint $$pkg ; \
	done
	goimports -l -w `find . -type f -name '*.go' | grep -v /vendor/`
	errcheck -verbose -ignoretests -blank `go list ./... | grep -v /vendor/`
	go vet -v `go list ./... | grep -v /vendor/`
	go fmt `go list ./... | grep -v /vendor/`

.PHONY: test
test:
	go test -v `go list ./... | grep -v /vendor/`

.PHONY: coverage
coverage:
	gocov test -v `go list ./... | grep -v /vendor/ ` > $(COVDIR)/coverage.out 2>/dev/null
	gocov report $(COVDIR)/coverage.out
	@if test -z "$$CI"; then \
		gocov-html $(COVDIR)/coverage.out > $(COVDIR)/coverage.html; \
		if which open &>/dev/null; then \
		    open $(COVDIR)/coverage.html; \
		fi; \
	fi
