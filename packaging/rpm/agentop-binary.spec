# Binary-only RPM for all COPR chroots (Fedora 43/44/rawhide + EPEL 9 / RHEL 9).
# No Source0 — the correct architecture tarball is downloaded at build time in %prep.
# This avoids the single-SRPM / multi-arch source problem entirely.

%global debug_package %{nil}

# Map RPM build arch to GoReleaser archive naming.
%ifarch x86_64
%global binary_arch x86_64
%endif
%ifarch aarch64
%global binary_arch arm64
%endif
%ifarch i386 i686
%global binary_arch i386
%endif

Name:           agentop
Version:        0.1.2
Release:        1%{?dist}
Summary:        Terminal dashboard for AI coding assistant sessions
License:        MIT
URL:            https://github.com/mohitmishra786/agentop

ExclusiveArch:  x86_64 aarch64 i386 i686

BuildRequires:  curl
BuildRequires:  tar

%description
agentop reads AI assistant session data from ~/.claude/projects/ and shows
token usage, cost, and cache efficiency in a clean terminal dashboard.
Supports Claude Code sessions with duf-style grid layout, color-coded token
bars, anomaly detection, and multi-view subcommands (today, daily, monthly,
blocks, session, doctor).

%prep
# Download the pre-built GoReleaser archive for this build architecture.
curl -fL \
  "https://github.com/mohitmishra786/agentop/releases/download/v%{version}/agentop_%{version}_linux_%{binary_arch}.tar.gz" \
  -o "%{_builddir}/agentop.tar.gz"
tar -xzf "%{_builddir}/agentop.tar.gz" -C "%{_builddir}"

%build
# Nothing to compile — pre-built static binary.

%install
cd "%{_builddir}/agentop_%{version}_linux_%{binary_arch}"
install -Dpm 0755 agentop    %{buildroot}%{_bindir}/agentop
install -Dpm 0644 agentop.1  %{buildroot}%{_mandir}/man1/agentop.1
install -Dpm 0644 LICENSE    %{buildroot}%{_datadir}/licenses/%{name}/LICENSE

%files
%{_bindir}/agentop
%{_mandir}/man1/agentop.1*
%{_datadir}/licenses/%{name}/LICENSE

%changelog
* Fri Apr 18 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 0.1.2-1
- Binary-only package for all COPR chroots; curl-downloads arch tarball at build time
