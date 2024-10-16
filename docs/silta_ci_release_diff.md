## silta ci release diff

Diff release resources

### Synopsis

Release diff command is used to compare the resources of a release with the current state of the cluster.
	
	* Chart allows prepending extra configuration (to helm --values line) via 
	"SILTA_<chart_name>_CONFIG_VALUES" environment variable. It has to be a 
	base64 encoded string of a silta configuration yaml file.

	* If IMAGE_PULL_SECRET is set (base64 encoded), it will be added to the 
	release values as imagePullSecret.
	

```
silta ci release diff [flags]
```

### Options

```
      --branchname string               Repository branchname that will be used for release name and environment name creation
      --chart-name string               Chart name
      --chart-repository string         Chart repository (default "https://storage.googleapis.com/charts.wdr.io")
      --chart-version string            Diff a specific chart version
      --cluster-domain string           Base domain for cluster urls (i.e. dev.example.com)
      --cluster-type string             Cluster type (i.e. gke, aws, aks, other)
      --db-root-pass string             Database password for root account
      --db-user-pass string             Database password for user account
      --gitauth-password string         Gitauth server password
      --gitauth-username string         Gitauth server username
      --helm-flags string               Extra flags for helm release
  -h, --help                            help for diff
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

