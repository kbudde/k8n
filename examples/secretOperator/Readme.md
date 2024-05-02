# secretOperator

This is a simple example of a secret operator that reads secrets with label `k8n.budd.ee/secret-operator=enabled` from namespace `prod-secrets` and writes it to all namespaces with label `env=prod`.

## implementation

Two watchers are defined in `config.yaml`:

- prodSecrets: Watches for secrets with label `k8n.budd.ee/secret-operator=enabled` in namespace `prod-secrets`.
- namespacesProd: Watches for namespaces with label `env=prod`.

## tests

The tests are defined as `input.XYZ.yaml` and `output.XYZ.yaml` files in the directory with `config.yaml`. The tests are run using the `test` command:

```bash
k8n test
```

The input files will be used with the config and ytt files in the folder. The output files are compared with the expected output files.
