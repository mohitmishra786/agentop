# Disable debugsource package — binary is stripped (-s linker flag)
%global debug_package %{nil}

Name:           agentop
Version:        1.0.0
Release:        1%{?dist}
Summary:        Terminal dashboard for AI coding assistant sessions

License:        MIT
URL:            https://github.com/mohitmishra786/agentop
Source0:        https://github.com/mohitmishra786/%{name}/archive/refs/tags/v%{version}.tar.gz

BuildRequires:  golang >= 1.25
BuildRequires:  git

%description
agentop aggregates session data from all major AI coding assistants
(Claude Code, Codex CLI, Cursor, Copilot CLI, Gemini CLI, Kiro,
OpenCode, Continue, JetBrains, and Grok CLI) and renders token usage,
cost, and cache efficiency in a clean terminal dashboard.

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
* Sun Jun 28 2026 Mohit Mishra <mohitmishra786687@gmail.com> - 1.0.0-1
- v1.0.0: multi-agent dashboard with 10 adapter support, budget tracking, pricing config
