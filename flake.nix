{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default";
  };

  outputs = {
    nixpkgs,
    systems,
    ...
  }: let
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
  in {
    devShells =
      forEachSystem
      (system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        default = pkgs.mkShell {
          packages = [
            pkgs.git
            pkgs.go
            pkgs.gopls
            pkgs.gh
            pkgs.bubblewrap
            pkgs.python313
            pkgs.python3Packages.pyyaml
          ];

          shellHook = ''
            echo "MCP Task Manager Development Environment"
            echo -n "Go version: "
            go version
          '';
        };
      });
  };
}
