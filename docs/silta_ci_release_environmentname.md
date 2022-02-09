## silta ci release environmentname

Return environment name

### Synopsis

Generate enviornment name based on branchname and release-suffix. 
		This is used in some helm charts

```
silta ci release environmentname [flags]
```

### Options

```
      --branchname string       Repository branchname that will be used for release name
  -h, --help                    help for environmentname
      --release-suffix string   Release name suffix
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

