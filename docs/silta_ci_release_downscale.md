## silta ci release downscale

Downscale a release

### Synopsis

Downscale a release the same way silta-downscaler does it: redirect the release service
to the placeholder upscaler proxy, mark ingress as down, suspend MariaDB resources and cronjobs,
and scale deployments and statefulsets to 0. Use "silta ci release wakeup" to restore it.

```
silta ci release downscale [flags]
```

### Options

```
      --dry-run                                Print changes without applying them
  -h, --help                                   help for downscale
      --namespace string                       Project name (namespace, i.e. "drupal-project")
      --placeholder-proxy-image string         Upscaler proxy image (default: read from silta-downscaler cronjob)
      --placeholder-service-name string        Placeholder upscaler service name (default: read from silta-downscaler cronjob)
      --placeholder-service-namespace string   Placeholder upscaler service namespace (default: read from silta-downscaler cronjob)
      --release-name string                    Release name
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

