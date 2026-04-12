# nixpkgs derivation for agentop
# Submit to: https://github.com/NixOS/nixpkgs
# Path in nixpkgs: pkgs/by-name/ag/agentop/package.nix
#
# To get the vendorHash:
#   1. Set vendorHash to lib.fakeHash
#   2. Run: nix-build -A agentop (it will fail with the correct hash)
#   3. Replace lib.fakeHash with the actual hash
{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "agentop";
  version = "0.1.1";

  src = fetchFromGitHub {
    owner = "mohitmishra786";
    repo = "agentop";
    rev = "v${version}";
    hash = "sha256-aXMxjtEjBQUy+81zQHqM83NQKe1BjuE1UMzAlY2WAMI=";
  };

  vendorHash = null;

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
  ];

  meta = {
    description = "Terminal dashboard for AI coding assistant sessions — token usage, cost, and cache efficiency";
    homepage = "https://github.com/mohitmishra786/agentop";
    license = lib.licenses.mit;
    maintainers = [ ];
    mainProgram = "agentop";
  };
}
