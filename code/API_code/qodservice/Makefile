COMPNAME = qodservice

VERSION ?= $(shell git describe --tags) # Top level makefile might set a common version for all comps
BUILD_TIME ?= $(shell date) # # Top level makefile might set a common version for all comps
COMMIT_HASH = $(shell git log --oneline -1 | awk '{print $$1}')
LDFLAGS = -X 'github.com/sfnuser/qodservice/util.Version=$(VERSION)' \
		  -X 'github.com/sfnuser/qodservice/util.BuildTime=$(BUILD_TIME)' \
		  -X 'github.com/sfnuser/qodservice/util.CommitHash=$(COMMIT_HASH)' \

.PHONY: $(COMPNAME) clean

all: $(COMPNAME)

$(COMPNAME):
	go build -ldflags="$(LDFLAGS)" -o $(COMPNAME)

clean:
	rm -f $(COMPNAME)