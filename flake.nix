{
  description = "Development environment for git-cm";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShell = pkgs.mkShell {
          buildInputs = [
            # Go
            pkgs.go_1_24
            pkgs.gopls
            pkgs.delve
            pkgs.golangci-lint
            # Other tools
            pkgs.curl
            pkgs.docker
            pkgs.git
            pkgs.github-cli
            pkgs.gnumake
            pkgs.jq
          ];

          shellHook = ''
            echo "Welcome to the git-cm development shell!"
          '';
        };
      });
}
