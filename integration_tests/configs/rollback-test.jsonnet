local config = import 'default.jsonnet';

config {
  'cronos_777-1'+: {
    validators: super.validators[0:1] + [{
      name: 'fullnode',
    }],
  },
}
