{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = [ 
    pkgs.go
    pkgs.gotools
    pkgs.gopls
    pkgs.go-outline
    pkgs.gocode
    pkgs.gopkgs
    pkgs.gocode-gomod
    pkgs.godef
    pkgs.golint
    pkgs.goose
    pkgs.cobra-cli
    pkgs.cowsay
  ];

  inputsFrom = [];

  shellHook = ''
  cowsay "Savvy CLI!"
  '';
} 

