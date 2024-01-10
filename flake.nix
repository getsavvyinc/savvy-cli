{
  description = "Savvy";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        {
          devShells.default = pkgs.mkShell {
            buildInputs = [
              pkgs.go
              pkgs.vhs
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
              pkgs.git
            ];

        shellHook = ''
            cowsay "Savvy CLI!"
        '';
        };
    }
    );
}
