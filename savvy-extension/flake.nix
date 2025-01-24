{
  description = "Chrome extension development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/de1864217bfa9b5845f465e771e0ecb48b30e02d";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Node.js 23
            nodejs_23
            # TypeScript and development tools
            nodePackages.typescript
            nodePackages.typescript-language-server
            #nodePackages.vite

            # Additional useful tools
            nodePackages.npm
          ];

          shellHook = ''
            # Only run the initialization once
            if [ -z "$IN_NIX_SHELL_INIT" ]; then
              export IN_NIX_SHELL_INIT=1
              
              echo "Chrome Extension Development Environment"
              echo "Available tools:"
              echo "- Node.js $(node --version)"
              echo "- npm $(npm --version)"
              echo "- TypeScript $(tsc --version)"
              echo "- Vite $(vite --version)"

              # Use system zsh with user's existing configuration
              if [ -x "$(command -v zsh)" ]; then
                exec zsh
              else
                echo "zsh not found, falling back to default shell"
              fi
            fi
          '';
        };
      }
    );
}
