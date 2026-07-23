# Local build / install helpers for probo-agent.
#
# macOS (primary for browser-enroll testing):
#   make install     build signed PKG, uninstall leftovers, install PKG
#   make uninstall   remove all system artifacts (idempotent)
#   make clean       uninstall + wipe local build caches
#
# Requires on Darwin: CODESIGN_IDENTITY, APPLE_TEAM_ID
# Optional: INSTALLER_IDENTITY; notarize via APPLE_ID+APPLE_ID_PASSWORD
#   (stored into a keychain profile; submit uses --keychain-profile)
#
# CLI-only (no app / helper), any Unix:
#   make install-cli
#   make run

CP ?=		cp
MKDIR ?=	mkdir -p
RMRF ?=		rm -rf
SUDO ?=		sudo

REPO_ROOT=		$(abspath ../..)
PROBO_AGENT_BIN=	$(REPO_ROOT)/bin/probo-agent
VERSION=		$(shell cat VERSION)

STATE_DIR=		$(HOME)/.local/share/probo-agent-dev
CACHE_ROOT=		$(HOME)/.cache/probo-agent-dev
DEV_TAG=		probo-agent/dev
BINARY=			/usr/local/bin/probo-agent
RELEASE_DIR=		$(CACHE_ROOT)/release/$(DEV_TAG)
INSTALL_SCRIPT=		installer/install.sh
MACOS_BUILD_SCRIPT=	installer/macos/build.sh
MACOS_UNINSTALL_SCRIPT=	installer/macos/uninstall.sh
MACOS_REINSTALL_SCRIPT=	installer/macos/reinstall.sh
ENROLL_UI_BUILD=	installer/macos/enroll-ui/.build

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
BUILD_ARCH=	amd64
else ifeq ($(UNAME_M),amd64)
ARCH_LABEL=	x86_64
BUILD_ARCH=	amd64
else ifeq ($(UNAME_M),arm64)
ARCH_LABEL=	arm64
BUILD_ARCH=	arm64
else ifeq ($(UNAME_M),aarch64)
ARCH_LABEL=	arm64
BUILD_ARCH=	arm64
else
ARCH_LABEL=
BUILD_ARCH=
endif

AGENT_DIR=		probo-agent_$(OS_LABEL)_$(ARCH_LABEL)
ARCHIVE_NAME=		$(AGENT_DIR).tar.gz
ARCHIVE_PATH=		$(RELEASE_DIR)/$(ARCHIVE_NAME)
STAGING_DIR=		$(CACHE_ROOT)/staging/$(AGENT_DIR)
BUILD_BINARY=		$(STAGING_DIR)/probo-agent

# Fat binary required by installer/macos/build.sh (arm64 + x86_64).
FAT_BINARY=		$(REPO_ROOT)/dist/probo-agent_fat
THIN_ARM64=		$(REPO_ROOT)/dist/probo-agent_arm64
THIN_AMD64=		$(REPO_ROOT)/dist/probo-agent_amd64
PKG=			$(REPO_ROOT)/dist/probo-agent_$(VERSION)_darwin.pkg
AGENT_LDFLAGS=		-ldflags "-X 'main.version=$(VERSION)'"

INSTALL_ARGS?=		--skip-service --dir "$(STATE_DIR)"
INSTALL_ENV=		PROBO_AGENT_RELEASE_TAG="$(DEV_TAG)" \
			PROBO_AGENT_RELEASE_BASE="file://$(abspath $(RELEASE_DIR))" \
			PROBO_AGENT_SKIP_CHECKSUM_VERIFY=true \
			PROBO_AGENT_STATE_DIR="$(STATE_DIR)" \
			PROBO_SERVER_URL="$(PROBO_SERVER_URL)" \
			PROBO_ENROLLMENT_TOKEN="$(PROBO_ENROLLMENT_TOKEN)"

.PHONY: help all pkg install uninstall clean clean-build install-cli run
.PHONY: $(PROBO_AGENT_BIN) $(FAT_BINARY)

all: help

help: ## Show targets
	@printf '%s\n' \
		'Targets:' \
		'  install       macOS: build signed PKG, wipe previous install, install PKG' \
		'                other: same as install-cli' \
		'  uninstall     Remove system install (macOS: full PKG/helper/app teardown)' \
		'  clean         uninstall + remove local build caches' \
		'  clean-build   Remove local caches only (no sudo)' \
		'  pkg           Build signed macOS PKG to dist/ (Darwin only)' \
		'  install-cli   Install binary via installer/install.sh (dev state dir)' \
		'  run           Foreground run against install-cli state dir' \
		'' \
		'Darwin install requires CODESIGN_IDENTITY and APPLE_TEAM_ID.'

# --- primary local test loop -------------------------------------------------

install: ## Install for local testing (PKG on macOS)
ifeq ($(UNAME_S),Darwin)
	$(MAKE) pkg
	$(SUDO) "$(MACOS_REINSTALL_SCRIPT)" "$(PKG)"
else
	$(MAKE) install-cli
endif

uninstall: ## Remove all system artifacts
ifeq ($(UNAME_S),Darwin)
	$(SUDO) "$(MACOS_UNINSTALL_SCRIPT)"
else
	@if [ -x "$(BINARY)" ]; then \
		$(SUDO) "$(BINARY)" uninstall || true; \
	fi
	$(SUDO) $(RMRF) "$(BINARY)"
endif

clean: uninstall clean-build ## uninstall + wipe build caches

clean-build: ## Wipe local build caches (no sudo)
	$(RMRF) "$(STATE_DIR)" "$(CACHE_ROOT)" "$(ENROLL_UI_BUILD)"
	-$(RMRF) "$(PKG)" "$(FAT_BINARY)" "$(THIN_ARM64)" "$(THIN_AMD64)"
	@# Native binary under /usr/local is owned by root after install; leave it
	@# to `uninstall`. Never remove it here without sudo.

# --- macOS PKG ---------------------------------------------------------------

pkg: $(PKG) ## Build signed .pkg into dist/

$(FAT_BINARY):
ifeq ($(UNAME_S),Darwin)
	@$(MKDIR) "$(dir $(FAT_BINARY))"
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(AGENT_LDFLAGS) \
		-o "$(THIN_ARM64)" "$(REPO_ROOT)/cmd/probo-agent"
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(AGENT_LDFLAGS) \
		-o "$(THIN_AMD64)" "$(REPO_ROOT)/cmd/probo-agent"
	lipo -create "$(THIN_ARM64)" "$(THIN_AMD64)" -output "$(FAT_BINARY)"
else
	@echo 'error: fat binary target is macOS-only' >&2
	@exit 1
endif

$(PKG): $(FAT_BINARY)
ifeq ($(UNAME_S),Darwin)
	@if [ -z "$(CODESIGN_IDENTITY)" ]; then \
		echo 'error: CODESIGN_IDENTITY is required' >&2; exit 2; \
	fi
	@if [ -z "$(APPLE_TEAM_ID)" ]; then \
		echo 'error: APPLE_TEAM_ID is required' >&2; exit 2; \
	fi
	@$(MKDIR) "$(dir $(PKG))"
	@CODESIGN_IDENTITY="$(CODESIGN_IDENTITY)" \
	APPLE_TEAM_ID="$(APPLE_TEAM_ID)" \
	INSTALLER_IDENTITY="$(INSTALLER_IDENTITY)" \
	APPLE_ID="$(APPLE_ID)" \
	APPLE_ID_PASSWORD="$(APPLE_ID_PASSWORD)" \
	NOTARYTOOL_KEYCHAIN_PROFILE="$(NOTARYTOOL_KEYCHAIN_PROFILE)" \
	sh "$(MACOS_BUILD_SCRIPT)" \
		--binary "$(FAT_BINARY)" \
		--version "$(VERSION)" \
		--output "$(PKG)"
else
	@echo 'error: pkg target is macOS-only' >&2
	@exit 1
endif

# --- CLI-only (no Probo Agent.app / helper) ----------------------------------

install-cli: $(ARCHIVE_PATH) ## Install binary only via install.sh
	@$(SUDO) $(INSTALL_ENV) sh "$(INSTALL_SCRIPT)" $(INSTALL_ARGS)

run: ## Run agent in foreground (install-cli state dir)
	$(SUDO) "$(BINARY)" run --dir "$(STATE_DIR)"

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
