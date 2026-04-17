# COPR build hook — https://docs.pagure.org/copr.copr/user_documentation.html#copr-makefile
#
# Fedora chroots build from source (Go toolchain available).
# EPEL/RHEL 9 chroots use the binary spec (no Go ≥ 1.25 in base repos).
#
# To add epel-9-x86_64 and epel-9-aarch64 to the COPR project:
#   1. Go to https://copr.fedorainfracloud.org/coprs/chessman/agentop/edit/
#   2. Under "Chroots", enable:
#        epel-9-x86_64
#        epel-9-aarch64
#   3. Save. COPR will use this Makefile's srpm target on the next build.

VERSION := 0.1.2

# Detect whether we are on an EPEL/RHEL build host.
# COPR passes %{?dist} which contains ".el9" for EPEL 9 targets.
SPEC_SOURCE := $(shell \
  if echo "$(COPR_CHROOT)" | grep -q "epel\|rhel"; then \
    echo packaging/rpm/agentop-binary.spec; \
  else \
    echo packaging/rpm/agentop.spec; \
  fi)

srpm:
	@echo "Building SRPM for chroot=$(COPR_CHROOT) using $(SPEC_SOURCE)"
	@spectool -g -R $(SPEC_SOURCE)
	@rpmbuild -bs \
	    --define "_sourcedir $(PWD)" \
	    --define "_srcrpmdir $(outdir)" \
	    $(SPEC_SOURCE)

.PHONY: srpm
