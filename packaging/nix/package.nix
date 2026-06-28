{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "agentop";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "mohitmishra786";
    repo = "agentop";
    rev = "v${version}";
    hash = "sha256-MTu32h3sL0fAavW/JGJaZg9yl3JuBjYi7ktxjN78Q9w=";
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
