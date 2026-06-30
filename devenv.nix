{ pkgs, ... }:

{
  packages = [
    pkgs.gcc
  ];

  languages.go = {
    enable = true;
    package = pkgs.go;
  };
}
