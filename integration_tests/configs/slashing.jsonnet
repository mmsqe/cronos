local config = import 'default.jsonnet';

config {
  'cronos_777-1'+: {
    validators: super.validators + [
      {
        coins: '1000000000000000000stake,10000000000000000000000basetcro',
        staked: '1000000000000000000stake',
        client_config: {
          'broadcast-mode': 'sync',
        },
      }
      for i in std.range(1, 2)
    ],
    genesis+: {
      app_state+: {
        staking+: {
          params: {
            unbonding_time: '10s',
          },
        },
        slashing+: {
          params: {
            signed_blocks_window: '10',
            slash_fraction_downtime: '0.01',
            downtime_jail_duration: '60s',
          },
        },
      },
    },
  },
}
