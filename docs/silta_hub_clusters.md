## silta hub clusters

List the clusters accessible from your Silta account

### Synopsis

List the clusters the logged-in Silta user may access, with the
kubeconfig context name, assigned namespace and (when reported by cluster
inventory) the Kubernetes version.

Use --json for machine-readable output (exit code 0 on success, 1 on
failure).

```
silta hub clusters [flags]
```

### Options

```
  -h, --help   help for clusters
      --json   Output clusters as JSON
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta hub](silta_hub.md)	 - Interact with a Silta hub

