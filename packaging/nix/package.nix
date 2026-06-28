{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "agentop";
  version = "1.0.2";

  src = fetchFromGitHub {
    owner = "mohitmishra786";
    repo = "agentop";
    rev = "v${version}";
    hash = "sha256-FaYT7IhWmbxG5c0slp7cZQfsmFCw0YGfCDySWR4f2js=";
  };

  vendorHash = "sha256-AR1IpxAlwRK1uH3iBXMstIZJ/eIV9yeSYtYY2Q29rLg=";

  ldflags = [
    "-s"
    "-w"
    "-X github.com/agentop-dev/agentop/cmd.Version=${version}"
  ];

  meta = {
    description = "Terminal dashboard for AI coding assistant sessions — token usage, cost, and cache efficiency";
    homepage = "https://github.com/mohitmishra786/agentop";
    license = lib.licenses.mit;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "agentop";
  };
}
