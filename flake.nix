{
  description = "Terminal dashboard for AI coding assistant sessions";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "agentop";
            version = self.shortRev or "dev";

            __structuredAttrs = true;

            src = self;

            vendorHash = "sha256-AR1IpxAlwRK1uH3iBXMstIZJ/eIV9yeSYtYY2Q29rLg=";

            ldflags = [
              "-s"
              "-w"
              "-X github.com/agentop-dev/agentop/cmd.Version=${self.shortRev or "dev"}"
            ];

            meta = {
              description = "Terminal dashboard for AI coding assistant sessions — token usage, cost, and cache efficiency";
              homepage = "https://github.com/mohitmishra786/agentop";
              license = pkgs.lib.licenses.mit;
              mainProgram = "agentop";
            };
          };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = [ pkgs.go ];
          };
        }
      );
    };
}
