local config = import 'default.jsonnet';

config {
  'cronos_777-1'+: {
    'app-config'+: {
      'app-db-backend': 'rocksdb',
      'minimum-gas-prices': '100000000000basetcro',
      store:: super.store,
      streamers:: super.streamers,
      'iavl-lazy-loading':: super['iavl-lazy-loading'],
    },
    validators: [super.validators[0] {
      'app-config':: super['app-config'],
    }] + super.validators[1:],
    genesis+: {
      app_state+: {
        cronos:: super.cronos,
        feemarket:: super.feemarket,
        gov: {
          voting_params: {
            voting_period: '10s',
          },
          deposit_params: {
            max_deposit_period: '10s',
            min_deposit: [
              {
                denom: 'basetcro',
                amount: '1',
              },
            ],
          },
        },
      },
    },
  },
}
