{
  description = "smr-cli";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs-unstable.url = "flake:nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs, flake-utils, nixpkgs-unstable }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
            pkgs = nixpkgs.legacyPackages.${system};
            unstable = nixpkgs-unstable.legacyPackages.${system}; in
        {
          devShells.default = import ./shell.nix { inherit pkgs unstable; };
        }
      );
}