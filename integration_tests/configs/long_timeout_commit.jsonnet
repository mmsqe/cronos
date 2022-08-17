local default = import 'default.jsonnet';

default {
  'cronos_777-1'+: {
    'start-flags': '--trace --log_level debug',
    config+: {
      consensus+: {
        timeout_commit: '15s',
      },
    },
  },
}
