# COPR build hook — https://docs.pagure.org/copr.copr/user_documentation.html#copr-makefile
#
# Uses the binary spec for ALL chroots (Fedora 43/44/rawhide + EPEL 9).
# The binary spec downloads the pre-built GoReleaser tarball, so no Go toolchain
# is required. This works on every chroot without per-chroot routing.
#
# COPR_CHROOT is NOT available during make_srpm (only during binary builds),
# so per-chroot spec selection is not possible here.

VERSION := 0.1.2
SPEC    := packaging/rpm/agentop-binary.spec

srpm:
	@echo "Building SRPM using $(SPEC) for version $(VERSION)"
	@# Create SRPM without downloading sources — each chroot downloads its own
	@# architecture-specific binary via the Source0 URL at build time.
	@rpmbuild -bs \
	    --nodeps \
	    --define "_sourcedir $(PWD)" \
	    --define "_srcrpmdir $(outdir)" \
	    --define "_specdir   $(PWD)/packaging/rpm" \
	    $(SPEC)

.PHONY: srpm
