{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "agentop";
  version = "0.1.2";

  src = fetchFromGitHub {
    owner = "mohitmishra786";
    repo = "agentop";
    rev = "v${version}";
    # TODO: run `nix-prefetch-url` after v0.1.2 tag is pushed to get the real hash
    hash = "sha256-0000000000000000000000000000000000000000000000000000";
  };

  # TODO: run `gomod2nix` or `nix build` and copy the vendorHash from the error
  vendorHash = "sha256-0000000000000000000000000000000000000000000000000000";

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
  ];

  meta = {
    description = "Terminal dashboard for AI coding assistant sessions — token usage, cost, and cache efficiency";
    homepage = "https://github.com/mohitmishra786/agentop";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "agentop";
  };
}
