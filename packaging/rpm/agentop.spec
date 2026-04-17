# Disable debugsource package — binary is stripped (-s linker flag)
%global debug_package %{nil}

Name:           agentop
Version:        0.1.2
Release:        1%{?dist}
Summary:        Terminal dashboard for AI coding assistant sessions

License:        MIT
URL:            https://github.com/mohitmishra786/agentop
Source0:        https://github.com/mohitmishra786/%{name}/archive/refs/tags/v%{version}.tar.gz

BuildRequires:  golang >= 1.23
BuildRequires:  git

%description
agentop reads AI assistant session data from ~/.claude/projects/ and shows
token usage, cost, and cache efficiency in a clean terminal dashboard.
Supports Claude Code sessions with duf-style grid layout, color-coded
token bars, empty session filtering, and anomaly detection.

%prep
%autosetup

%build
export CGO_ENABLED=0
go build \
    -trimpath \
    -ldflags "-s -w -X main.Version=%{version}" \
    -o %{name} \
    .

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}
install -Dpm 0644 %{name}.1 %{buildroot}%{_mandir}/man1/%{name}.1

%check
go test ./...

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1*

%changelog
* Mon Apr 13 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 0.1.2-1
- Initial package
