CP ?=		cp
MKDIR ?=	mkdir -p
SUDO ?=		sudo

STATE_DIR=		$(HOME)/.local/share/probo-agent-dev
CACHE_ROOT=		$(HOME)/.cache/probo-agent-dev
DEV_TAG=		probo-agent/dev
BINARY=			/usr/local/bin/probo-agent
RELEASE_DIR=		$(CACHE_ROOT)/release/$(DEV_TAG)
INSTALL_SCRIPT=		installer/install.sh
REPO_ROOT=		../..
PROBO_AGENT_BIN=	$(REPO_ROOT)/bin/probo-agent

UNAME_S:=		$(shell uname -s)
UNAME_M:=		$(shell uname -m)

ifeq ($(UNAME_S),Darwin)
OS_LABEL=	Darwin
else ifeq ($(UNAME_S),Linux)
OS_LABEL=	Linux
else ifeq ($(UNAME_S),FreeBSD)
OS_LABEL=	Freebsd
else
OS_LABEL=
endif

ifeq ($(UNAME_M),x86_64)
ARCH_LABEL=	x86_64
else ifeq ($(UNAME_M),arm64)
ARCH_LABEL=	arm64
else ifeq ($(UNAME_M),aarch64)
ARCH_LABEL=	arm64
else
ARCH_LABEL=
endif

AGENT_DIR=		probo-agent_$(OS_LABEL)_$(ARCH_LABEL)
ARCHIVE_NAME=		$(AGENT_DIR).tar.gz
ARCHIVE_PATH=		$(RELEASE_DIR)/$(ARCHIVE_NAME)
STAGING_DIR=		$(CACHE_ROOT)/staging/$(AGENT_DIR)
BUILD_BINARY=		$(STAGING_DIR)/probo-agent

INSTALL_ARGS?=		--skip-service --dir "$(STATE_DIR)"
INSTALL_ENV=		PROBO_AGENT_RELEASE_TAG="$(DEV_TAG)" \
                    PROBO_AGENT_RELEASE_BASE="file://$(abspath $(RELEASE_DIR))" \
                    PROBO_AGENT_SKIP_CHECKSUM_VERIFY=true \
                    PROBO_AGENT_STATE_DIR="$(STATE_DIR)" \
                    PROBO_SERVER_URL="$(PROBO_SERVER_URL)" \
                    PROBO_ENROLLMENT_TOKEN="$(PROBO_ENROLLMENT_TOKEN)"

.PHONY: all install run clean
all: install

install: $(ARCHIVE_PATH)
	$(SUDO) $(INSTALL_ENV) sh "$(INSTALL_SCRIPT)" $(INSTALL_ARGS)

run:
	$(SUDO) "$(BINARY)" run --dir "$(STATE_DIR)"

clean:
	rm -rf "$(STATE_DIR)" "$(CACHE_ROOT)" "$(BINARY)"

$(ARCHIVE_PATH): $(BUILD_BINARY)
	$(MKDIR) "$(RELEASE_DIR)"
	tar -czf "$(ARCHIVE_PATH)" -C "$(CACHE_ROOT)/staging" "$(AGENT_DIR)"

$(PROBO_AGENT_BIN):
	$(MAKE) -C "$(REPO_ROOT)" bin/probo-agent

$(BUILD_BINARY): $(PROBO_AGENT_BIN)
	$(MKDIR) "$(STAGING_DIR)"
	$(CP) "$(PROBO_AGENT_BIN)" "$(BUILD_BINARY)"
	$(CP) "$(REPO_ROOT)/README.md" "$(REPO_ROOT)/LICENSE" "$(STAGING_DIR)/"
	@if [ -f CHANGELOG.md ]; then $(CP) CHANGELOG.md "$(STAGING_DIR)/"; fi
