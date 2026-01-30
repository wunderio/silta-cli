## silta ci release delete-resources

Delete orphaned release resources

### Synopsis

Deletes release resources based on labels ("release", "app.kubernetes.io/instance" and "app=<release-name>-es" (for Elasticsearch storage))
		This command can be used to clean up resources when helm release configmaps are absent.
	

```
silta ci release delete-resources [flags]
```

### Options

```
      --delete-pvcs           Delete PVCs (default: true) (default true)
      --dry-run               Dry run (default: true) (default true)
  -h, --help                  help for delete-resources
      --namespace string      Project name (namespace, i.e. "drupal-project")
      --release-name string   Release name
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

