## silta ci image build

Build and push container image

```
silta ci image build [flags]
```

### Options

```
      --branchname string           Branch name (used as an extra tag for image identification)
      --build-path string           Docker image build path
      --dockerfile string           Dockerfile (relative path)
  -h, --help                        help for build
      --image-identifier string     Docker image identifier (i.e. "php")
      --image-repo-host string      (Docker) container image repository url
      --image-repo-project string   (Docker) image repository project (project name, i.e. "silta")
      --image-reuse                 Do not rebuild image if identical image:tag exists in remote (default true)
      --image-tag string            Docker image tag (optional, check '--image-reuse' flag)
      --image-tag-prefix string     Prefix for Docker image tag (optional)
      --namespace string            Project name (namespace, i.e. "drupal-project")
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci image](silta_ci_image.md)	 - CI (docker) image commands

