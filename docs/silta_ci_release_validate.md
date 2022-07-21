## silta ci release validate

Validate release

```
silta ci release validate [flags]
```

### Options

```
      --branchname string               Repository branchname that will be used for release name and environment name creation
      --chart-name string               Chart name
      --chart-repository string         Chart repository
      --chart-version string            Deploy a specific chart version
      --cluster-type string             Cluster type (i.e. gke, aws, aks, other)
  -h, --help                            help for validate
      --namespace string                Project name (namespace, i.e. "drupal-project")
      --release-name string             Release name
      --silta-config string             Silta release helm chart values
      --silta-environment-name string   Environment name override based on branchname and release-suffix. Used in some helm charts.
      --vpc-native string               VPC-native cluster (GKE specific)
      --vpn-ip string                   VPN IP for basic auth allow list
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

