## silta ci image url

Calculate container image url based on build content

```
silta ci image url [flags]
```

### Options

```
      --build-path string           Docker image build path
      --dockerfile string           Dockerfile (relative path)
  -h, --help                        help for url
      --image-identifier string     Docker image identifier (i.e. "php")
      --image-repo-host string      (Docker) container image repository url
      --image-repo-project string   (Docker) image repository project (project name, i.e. "silta")
      --image-tag string            Docker image tag (optional)
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

