std.manifestYamlDoc({
  services: {
    ['testplan-' + i]: {
      image: 'cronos-testground:9d77x43srq7pczib02l7hay1g2kvpihr',
      command: 'stateless-testcase run',
      container_name: 'testplan-' + i,
      volumes: [
        std.extVar('outputs') + ':/outputs',
      ],
      environment: {
        JOB_COMPLETION_INDEX: i,
      },
    }
    for i in std.range(0, std.extVar('nodes') - 1)
  },
})
