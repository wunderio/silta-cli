## silta ci release info

Print release information

### Synopsis

This will print release information. Required flags: "--release-name" and "--namespace". 

This command will post release information to github when there are following extra parameters provided:
--github-token
--pr-number
--pull-request
--project-organization
--project-reponame

Difference between "--project-reponame" and "--namespace" is that project-reponame can be uppercase, 
but namespace is normalized lowercase version of it.


```
silta ci release info [flags]
```

### Options

```
      --github-token string           Github token for posting release status to PR
  -h, --help                          help for info
      --namespace string              Project name (namespace, i.e. "drupal-project")
      --pr-number string              PR number
      --project-organization string   Repository username / organization
      --project-reponame string       Project repository name (i.e. "drupal-project"
      --pull-request string           Pull request url
      --release-name string           Release name
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci release](silta_ci_release.md)	 - CI release related commands

