# Packaging agentop

This document is for maintainers who want to add agentop to package repositories.

**Note**: Submission processes vary - some are direct uploads, some require PRs, and some involve complex review processes.

## Quick Reference

| Repository | Process | Difficulty | Files Ready |
|------------|---------|------------|-------------|
| **AUR** | Direct git push | Easy | [READY] PKGBUILD |
| **Scoop** | GitHub PR | Easy | [READY] manifest |
| **Chocolatey** | Direct push OR PR | Medium | [READY] installer |
| **Homebrew** | [READY] Already in tap | Done | [READY] Formula |
| **Debian** | Mentor sponsorship | Hard | [TODO] Need full debian/ |
| **Fedora** | Bugzilla + fedpkg | Hard | [TODO] Need .spec |
| **Alpine** | GitLab MR | Medium | [TODO] Need APKBUILD |
| **Nixpkgs** | GitHub PR | Medium | [TODO] Need default.nix |

---

## v0.1.0 Release Artifacts

## v0.1.0 Release Artifacts

Release artifacts for v0.1.0 are available at: https://github.com/mohitmishra786/agentop/releases/tag/v0.1.0

### Checksums (SHA256)

**Source Tarball:**
```
264c71d270847cca8a7c8b589683670f836ab9a95e6cffd3e5380e64704d1ec7  agentop-0.1.0.tar.gz
```

**Linux Binaries:**
```
d513df1ae82e0d4cf4b8b8a9135db30f3be614c8ac036675502cd363ed02fa74  agentop_v0.1.0_linux_x86_64.tar.gz
0db620ee653fa14437855a19da455dbc1b8b4e6a5e5f3cdc54754b4a4826fe98  agentop_v0.1.0_linux_arm64.tar.gz
1b5c55d6f3ced9182ccd13073a4955406e1c4834698d4487b4632c959e63fac9  agentop_v0.1.0_linux_i386.tar.gz
```

**macOS Binaries:**
```
24dabae9a19be1f55eb013d9a611295a54a606925e93ea4c399e80d936b931f8  agentop_v0.1.0_darwin_x86_64.tar.gz
57db3ec72970e3db6a3ac5d23ce2a0e02e0fa8ed4eed891dec078f05d77bfa25  agentop_v0.1.0_darwin_arm64.tar.gz
```

**Windows Binaries:**
```
f770744f131f04e90d199a2fac7c9e4d740e2f5bb06a2a538374ebd286cc9747  agentop_v0.1.0_windows_x86_64.zip
42b4c5a6859c37cb24ecdea8fb48c6c7331c2dd76f7c2d30b9f73b014c6a8f04  agentop_v0.1.0_windows_arm64.zip
```

**Linux Packages:**
```
ca1b18bb3875e49eec2308f6468387ae15e4f2694f55d68948af696a65c14af5  agentop_v0.1.0_linux_amd64.deb
6ae73e76a9993bb6a7486cc7b0e6eb2e7d383c29dce3a649e45c0743c4bab86d  agentop_v0.1.0_linux_arm64.deb
416cf66cffc354de851b40965b0fcd8648def6490b73b02199cd2819c3b64489  agentop_v0.1.0_linux_amd64.rpm
e66b5f972ab72866f6ce867e102746184665e436f1e7f17aa82591b74fbccb54  agentop_v0.1.0_linux_arm64.rpm
ae7a5114aa58f8367095db660cb27ae44fd2852f74354fc434250c5f978c51a5  agentop_v0.1.0_linux_amd64.apk
6d06442a3f2a4fe3b2bdbb5596431c25d56df7cc738f34849d4407f4ed0ec05a  agentop_v0.1.0_linux_arm64.apk
```

## Package Repository Submission Guide

### Arch Linux (AUR)

1. Clone the AUR package:
   ```bash
   git clone ssh://aur@aur.archlinux.org/agentop.git
   cd agentop
   ```

2. Copy the PKGBUILD from `packaging/aur/PKGBUILD`

3. Update checksum in PKGBUILD:
   ```bash
   sha256sums=('264c71d270847cca8a7c8b589683670f836ab9a95e6cffd3e5380e64704d1ec7')
   ```

4. Test build:
   ```bash
   makepkg -si
   ```

5. Commit and push:
   ```bash
   git add PKGBUILD
   git commit -m "upgpkg: agentop 0.1.0"
   git push
   ```

### Scoop (Windows)

1. Fork the Scoop bucket: https://github.com/ScoopInstaller/Main

2. Add manifest from `packaging/scoop/agentop.json`

3. Update checksum:
   ```json
   "hash": "f770744f131f04e90d199a2fac7c9e4d740e2f5bb06a2a538374ebd286cc9747"
   ```

4. Test installation:
   ```bash
   scoop install agentop.json
   ```

5. Submit PR to https://github.com/ScoopInstaller/Main

### Chocolatey (Windows)

1. Fork the Chocolatey repository: https://github.com/chocolatey/chocolatey-coreteampackages

2. Create package directory: `automatic/agentop`

3. Copy files from `packaging/chocolatey/`

4. Update checksums in chocolateyinstall.ps1:
   ```powershell
   $checksum64 = 'f770744f131f04e90d199a2fac7c9e4d740e2f5bb06a2a538374ebd286cc9747'
   $checksumarm64 = '42b4c5a6859c37cb24ecdea8fb48c6c7331c2dd76f7c2d30b9f73b014c6a8f04'
   ```

5. Test package:
   ```bash
   choco pack
   choco install agentop -source .
   ```

6. Submit PR to https://github.com/chocolatey/chocolatey-coreteampackages

### Debian (MENTORS)

1. Clone the packaging repository:
   ```bash
   git clone https://salsa.debian.org/go-team/packages/dpkg-golang.git
   ```

2. Create new package directory and add Debian packaging files

3. Import upstream tarball:
   ```bash
   uscan --download-current-version
   ```

4. Build package:
   ```bash
   dpkg-buildpackage -us -uc
   ```

5. Test and upload to mentors.debian.net

6. Request sponsorship for upload to unstable

### Fedora

1. Clone the Fedora packages repository:
   ```bash
   fedpkg clone agentop
   cd agentop
   ```

2. Create .spec file based on `packaging/` templates

3. Build locally:
   ```bash
   fedpkg local
   ```

4. Test and submit:
   ```bash
   fedpkg new-sources agentop_v0.1.0_linux_x86_64.tar.gz
   git add .
   git commit -m "Update to 0.1.0"
   fedpkg push
   ```

5. Create Bugzilla review request and follow Fedora package review process

### Alpine Linux

1. Clone the aports repository:
   ```bash
   git clone https://git.alpinelinux.org/aports
   cd aports/community/agentop
   ```

2. Create APKBUILD based on `packaging/` templates

3. Test build:
   ```bash
   abuild -r
   ```

4. Submit merge request to https://git.alpinelinux.org/aports

### Nixpkgs (NixOS)

1. Clone nixpkgs:
   ```bash
   git clone https://github.com/NixOS/nixpkgs
   cd nixpkgs
   ```

2. Add package to `pkgs/applications/networking/agentop/default.n`:
   ```nix
   { lib, buildGoModule, fetchFromGitHub }:
   buildGoModule rec {
     pname = "agentop";
     version = "0.1.0";
     src = fetchFromGitHub {
       owner = "mohitmishra786";
       repo = "agentop";
       rev = "v${version}";
       sha256 = "264c71d270847cca8a7c8b589683670f836ab9a95e6cffd3e5380e64704d1ec7";
     };
     vendorHash = null; # Update after build
     subPackages = [ "." ];
     meta = with lib; {
       description = "Terminal dashboard for AI coding assistant sessions";
       homepage = "https://github.com/mohitmishra786/agentop";
       license = licenses.mit;
       maintainers = [ maintainers.mohitmishra786 ];
     };
   }
   ```

3. Test build:
   ```bash
   nix-build -A agentop
   ```

4. Submit PR to https://github.com/NixOS/nixpkgs

## Installation Commands

### Homebrew

```ruby
class Agentop < Formula
  desc "Terminal dashboard for AI coding assistant sessions"
  homepage "https://github.com/mohitmishra786/agentop"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/mohitmishra786/agentop/releases/download/v#{version}/agentop_#{version}_darwin_amd64.tar.gz"
      sha256 "<intel-sha256>"
    end
    on_arm do
      url "https://github.com/mohitmishra786/agentop/releases/download/v#{version}/agentop_#{version}_darwin_arm64.tar.gz"
      sha256 "<arm-sha256>"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/mohitmishra786/agentop/releases/download/v#{version}/agentop_#{version}_linux_amd64.tar.gz"
      sha256 "<intel-sha256>"
    end
    on_arm do
      url "https://github.com/mohitmishra786/agentop/releases/download/v#{version}/agentop_#{version}_linux_arm64.tar.gz"
      sha256 "<arm-sha256>"
    end
  end

  def install
    bin.install "agentop"
    man1.install "agentop.1"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/agentop --version")
  end
end
```

### Arch Linux (AUR)

Create `PKGBUILD`:

```bash
pkgname=agentop
pkgver=0.1.0
pkgrel=1
pkgdesc="Terminal dashboard for AI coding assistant sessions"
arch=("x86_64" "aarch64" "armv7h")
url="https://github.com/mohitmishra786/agentop/archive/refs/tags/v$pkgver.tar.gz"
sha256sums=("<tarball-sha256>")
license=("MIT")
makedepends=("go")
depends=("glibc")

build() {
  cd "$pkgname-$pkgver"
  go build -o agentop .
}

package() {
  install -Dm755 agentop "$pkgdir/usr/bin/agentop"
  install -Dm644 agentop.1 "$pkgdir/usr/share/man/man1/agentop.1"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
```

### Debian/Ubuntu

The following files are generated in each release:
- `agentop_<version>_linux_amd64.deb`
- `agentop_<version>_linux_arm64.deb`

Install with:

```bash
sudo dpkg -i agentop_<version>_linux_amd64.deb
```

### Fedora/RHEL

The following RPM files are generated in each release:
- `agentop_<version>_linux_amd64.rpm`
- `agentop_<version>_linux_arm64.rpm`

Install with:

```bash
sudo dnf install agentop_<version>_linux_amd64.rpm
```

### Alpine Linux

The following APK file is generated in each release:
- `agentop_<version>_linux_x86_64.apk`

Install with:

```bash
sudo apk add agentop_<version>_linux_x86_64.apk
```

### Windows (Scoop)

Create `bucket/agentop.json`:

```json
{
  "version": "0.1.0",
  "description": "Terminal dashboard for AI coding assistant sessions",
  "homepage": "https://github.com/mohitmishra786/agentop",
  "license": "MIT",
  "url": "https://github.com/mohitmishra786/agentop/releases/download/v0.1.0/agentop_0.1.0_windows_amd64.zip",
  "hash": "<zip-sha256>",
  "extract_dir": "agentop",
  "bin": "agentop.exe"
}
```

### Windows (Chocolatey)

Create `agentop.nuspec`:

```xml
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/developer/2004/01/01/package.xsd">
  <metadata>
    <id>agentop</id>
    <version>0.1.0</version>
    <title>agentop</title>
    <authors>Mohit Mishra</authors>
    <owners>Mohit Mishra</owners>
    <description>Terminal dashboard for AI coding assistant sessions</description>
    <licenseUrl>https://github.com/mohitmishra786/agentop/blob/main/LICENSE</licenseUrl>
    <projectUrl>https://github.com/mohitmishra786/agentop</projectUrl>
    <tags>cli developer-tools ai</tags>
  </metadata>
</package>
```

## Verification

After installation, verify with:

```bash
agentop --version
```

Expected output format:

```
agentop 0.1.0 (abc123f) (built on 2026-04-08)
```
