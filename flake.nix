{
  description = "Universal calendar and task converter";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "salja";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
          subPackages = [ "cmd/salja" ];
          meta = with pkgs.lib; {
            description = "Universal calendar and task converter";
            homepage = "https://github.com/gongahkia/salja";
            license = licenses.mit;
            mainProgram = "salja";
          };
        };
      }
    );
}
