{
  description = "A Nix flake for hyde-config";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      goModule = pkgs.buildGoModule {
        pname = "hyde-config";
        version = "0.1.0"; # You might want to get this from go.mod or a tag
        src = ./.;
        vendorHash = "sha256-qQ7rr2Y+AnnuyW/N/ogwzT6lvyixHK31lM77Sv3ziiE="; # Run 'nix build .#hyde-config' to get the correct hash
        proxyVendor = true;
      };
    in
    {
      packages.${system}.default = goModule;
      defaultPackage.${system} = self.packages.${system}.hyde-config;

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          goModule
        ];

        shellHook = ''
          export PATH=$(pwd)/bin:$PATH
          echo "Welcome to the hyde-config development shell!"
        '';
      };
    };
}
