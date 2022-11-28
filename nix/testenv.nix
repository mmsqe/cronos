{ poetry2nix, lib, python310, rocksdb }:
poetry2nix.mkPoetryEnv {
  projectDir = ../integration_tests;
  python = python310;
  overrides = poetry2nix.overrides.withDefaults (lib.composeManyExtensions [
    (self: super:
      let
        buildSystems = {
          eth-bloom = [ "setuptools" ];
          pystarport = [ "poetry" ];
          cprotobuf = [ "setuptools" ];
          durations = [ "setuptools" ];
          multitail2 = [ "setuptools" ];
          pytest-github-actions-annotate-failures = [ "setuptools" ];
          flake8-black = [ "setuptools" ];
          multiaddr = [ "setuptools" ];
          rocksdb = [ "setuptools" "cython" "pkgconfig" ];
          pyroaring = [ "setuptools" "cython" ];
          roaring64 = [ "poetry" ];
        };
      in
      lib.mapAttrs
        (attr: systems: super.${attr}.overridePythonAttrs
          (old: {
            nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ map (a: self.${a}) systems;
          }))
        buildSystems
    )
    (self: super: {
      eth-bloom = super.eth-bloom.overridePythonAttrs {
        preConfigure = ''
          substituteInPlace setup.py --replace \'setuptools-markdown\' ""
        '';
      };
      pyyaml-include = super.pyyaml-include.overridePythonAttrs {
        preConfigure = ''
          substituteInPlace setup.py --replace "setup()" "setup(version=\"1.3\")"
        '';
      };
      rocksdb = super.rocksdb.overridePythonAttrs (
        old: {
          buildInputs = (old.buildInputs or [ ]) ++ [ rocksdb ];
        }
      );
    })
  ]);
}
