## silta ci release deploy

Deploy release

```
silta ci release deploy [flags]
```

### Options

```
      --branchname string               Repository branchname that will be used for release name and environment name creation
      --chart-name string               Chart name
      --chart-repository string         Chart repository
      --chart-version string            Deploy a specific chart version
      --cluster-domain string           Base domain for cluster urls (i.e. silta.wdr.io)
      --cluster-type string             Cluster type (i.e. gke, aws, aks, other)
      --db-root-pass string             Database password for root account
      --db-user-pass string             Database password for user account
      --deployment-timeout string       Helm deployment timeout
      --gitauth-password string         Gitauth server password
      --gitauth-username string         Gitauth server username
  -h, --help                            help for deploy
      --namespace string                Project name (namespace, i.e. "drupal-project")
      --nginx-image-url string          PHP image url
      --php-image-url string            PHP image url
      --release-name string             Release name
      --release-suffix string           Release name suffix for environment name creation
      --repository-url string           Repository url (i.e. git@github.com:wunderio/silta.git)
      --shell-image-url string          PHP image url
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

