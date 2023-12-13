{ pkgs, unstable }:

pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    unstable.go_1_21
    unstable.golangci-lint
  ];
}
