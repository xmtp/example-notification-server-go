{
  description = "Example notification server in Go";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem =
        { pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs =
              with pkgs;
              [
                go
                mockgen
                moreutils
                protoc-gen-go
                gopls
                buf
                jq
                shellcheck
                golangci-lint
              ]
              ++ lib.optionals pkgs.stdenv.isDarwin [ pkgs.darwin.cctools ];
          };
        };
    };
}
