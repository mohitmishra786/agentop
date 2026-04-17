# COPR build hook — https://docs.pagure.org/copr.copr/user_documentation.html#copr-makefile
#
# Fedora chroots build from source (Go toolchain available).
# EPEL/RHEL 9 chroots use the binary spec (no Go ≥ 1.25 in base repos).

VERSION := 0.1.2

# Detect whether we are on an EPEL/RHEL build host.
SPEC_SOURCE := $(shell \
  if echo "$(COPR_CHROOT)" | grep -q "epel\|rhel"; then \
    echo packaging/rpm/agentop-binary.spec; \
  else \
    echo packaging/rpm/agentop.spec; \
  fi)

srpm:
	@echo "Building SRPM for chroot=$(COPR_CHROOT) using $(SPEC_SOURCE)"
	@# Download sources — use curl directly; spectool is not guaranteed in mock chroot
	@grep -E "^Source[0-9]*:" $(SPEC_SOURCE) | awk '{print $$2}' | while read url; do \
	  fname=$$(basename "$$url"); \
	  if [ ! -f "$$fname" ]; then \
	    echo "Downloading $$url"; \
	    curl -fL "$$url" -o "$$fname"; \
	  fi; \
	done
	@rpmbuild -bs \
	    --define "_sourcedir $(PWD)" \
	    --define "_srcrpmdir $(outdir)" \
	    $(SPEC_SOURCE)

.PHONY: srpm
