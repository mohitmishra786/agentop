# COPR build hook — https://docs.pagure.org/copr.copr/user_documentation.html#copr-makefile
#
# The binary spec has no Source0 — the correct arch tarball is downloaded
# at binary build time via curl in %prep. So rpmbuild -bs just packages
# the spec file, which always works regardless of chroot or architecture.

SPEC := packaging/rpm/agentop-binary.spec

srpm:
	@echo "Building SRPM from $(SPEC)"
	@rpmbuild -bs \
	    --nodeps \
	    --define "_sourcedir $(PWD)" \
	    --define "_srcrpmdir $(outdir)" \
	    $(SPEC)

.PHONY: srpm
