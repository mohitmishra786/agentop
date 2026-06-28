# Binary-only RPM for all COPR chroots (Fedora 43/44/rawhide + EPEL 9 / RHEL 9).
# No Source0 — downloads the correct arch tarball at build time in %prep.
# GoReleaser archives have files at the root (no wrapper dir), extracted to %{_tmppath}.

%global debug_package %{nil}
%global agentop_src   %{_tmppath}/agentop-bin-%{version}

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
Version:        1.0.2
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
mkdir -p %{agentop_src}
curl -fL \
  "https://github.com/mohitmishra786/agentop/releases/download/v%{version}/agentop_%{version}_linux_%{binary_arch}.tar.gz" \
  -o "%{agentop_src}/agentop.tar.gz"
tar -xzf "%{agentop_src}/agentop.tar.gz" -C "%{agentop_src}/"

%build
# Pre-built static binary — nothing to compile.

%install
install -Dpm 0755 "%{agentop_src}/agentop"    %{buildroot}%{_bindir}/agentop
install -Dpm 0644 "%{agentop_src}/agentop.1"  %{buildroot}%{_mandir}/man1/agentop.1
install -Dpm 0644 "%{agentop_src}/LICENSE"    %{buildroot}%{_datadir}/licenses/%{name}/LICENSE

%clean
rm -rf "%{agentop_src}"

%files
%{_bindir}/agentop
%{_mandir}/man1/agentop.1*
%{_datadir}/licenses/%{name}/LICENSE

%changelog
* Sun Jun 28 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 1.0.2-1
- v1.0.2: Fix AggregateSession dropping tokens from 8/12 adapters
- Remove stopReason gate in aggregator — all adapters now count tokens
- Fix codex session_meta JSON tag (session_id -> id)
- Add windsurf implicit/ sibling directory discovery
- Improve claude IsAvailable to detect v2+ empty projects dir

* Sun Jun 28 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 1.0.1-1
- v1.0.1: Devin Desktop adapter, codex nil pointer fix
- Add Devin Desktop (SQLite) adapter for session discovery
- Fix codex adapter nil pointer on oversized lines
- Fix token double-counting in Devin adapter

* Sun Jun 28 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 1.0.0-1
- v1.0.0: multi-agent dashboard supporting 10 AI coding assistants
- New: Grok CLI, OpenCode, Kiro, JetBrains, Continue adapters
- Enhanced: pricing, filtering, UI themes, budget tracking
- Breaking: agentop now aggregates all agents, not just Claude
