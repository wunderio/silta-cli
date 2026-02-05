## silta ci release debug-failed

Debug failed deployment resources

### Synopsis

Debug failed deployment by checking:
- OOMKilled containers
- Failed pods with their events and logs
- Not-ready statefulsets with their events
- Not-ready deployments with their events

This command is typically called when a deployment fails to help diagnose the issue.

```
silta ci release debug-failed [flags]
```

### Options

```
  -h, --help                  help for debug-failed
      --namespace string      Namespace
      --release-name string   Release name
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

