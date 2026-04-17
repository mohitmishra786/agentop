# Binary-only RPM for EPEL/RHEL 9 and other distros where Go ≥ 1.25 is not
# available. Packages the pre-built binary produced by GoReleaser instead of
# building from source. Used by the chessman/agentop COPR project for
# epel-9-x86_64 and epel-9-aarch64 chroots.
#
# To build locally:
#   rpmbuild -bb --define "_sourcedir ." agentop-binary.spec
#   (download the tarball first: see Source0 URL below)

%global debug_package %{nil}

Name:           agentop
Version:        0.1.2
Release:        1%{?dist}
Summary:        Terminal dashboard for AI coding assistant sessions

License:        MIT
URL:            https://github.com/mohitmishra786/agentop

# Map RPM %_arch to GoReleaser archive arch names.
%ifarch x86_64
%global binary_arch x86_64
%endif
%ifarch aarch64
%global binary_arch arm64
%endif
%ifarch i386 i686
%global binary_arch i386
%endif

# The pre-built tarball produced by GoReleaser — no Go toolchain required.
Source0: https://github.com/mohitmishra786/%{name}/releases/download/v%{version}/%{name}_%{version}_linux_%{binary_arch}.tar.gz

# Explicitly list supported architectures so COPR skips unsupported chroots.
ExclusiveArch: x86_64 aarch64 i386 i686

# Runtime — agentop is a fully static binary (CGO_ENABLED=0).
Requires: /bin/sh

%description
agentop reads AI assistant session data from ~/.claude/projects/ and shows
token usage, cost, and cache efficiency in a clean terminal dashboard.
Supports Claude Code sessions with a duf-style grid layout, color-coded
token bars, anomaly detection, and multi-view subcommands (today, daily,
monthly, blocks, session, doctor).

%prep
# GoReleaser tarballs wrap contents in a directory named after the archive.
%setup -q -n %{name}_%{version}_linux_%{binary_arch}

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}
# Man page is included in the GoReleaser archive.
install -Dpm 0644 %{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1*

%changelog
* Fri Apr 18 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 0.1.2-1
- Binary-only package for EPEL/RHEL 9 compatibility (no Go toolchain required)
