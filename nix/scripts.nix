{ pkgs
, config
, cronos ? (import ../. { inherit pkgs; })
}: rec {
  start-cronos = pkgs.writeShellScriptBin "start-cronos" ''
    # rely on environment to provide cronosd
    ${../scripts/start-cronos} ${config.cronos-config} ${config.dotenv} $@
  '';
  start-geth = pkgs.writeShellScriptBin "start-geth" ''
    source ${config.dotenv}
    ${../scripts/start-geth} ${config.geth-genesis} $@
  '';
  start-scripts = pkgs.symlinkJoin {
    name = "start-scripts";
    paths = [ start-cronos start-geth ];
  };
}
