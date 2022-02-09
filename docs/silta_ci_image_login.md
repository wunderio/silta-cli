## silta ci image login

Image repository login

### Synopsis

Login to (docker) image repository. 
	
Use either flags or environment variables for authentication. 

Available flags and environment variables:

  * General (required):
    - "--image-repo-host" flag or "IMAGE_REPO_HOST" environment variable: (Docker) container image repository url

  * Google Cloud:
    - "--gcp-key-json" flag or "GCP_KEY_JSON" environment variable: Google Cloud service account key (string value)

  * Amazon Web Services:
    - "--aws-secret-access-key" flag or "AWS_SECRET_ACCESS_KEY" environment variable: Amazon Web Services IAM account key (string value)

  * Azure Services:
    - "--aks-tenant-id" flag or "AKS_TENANT_ID" environment variable: Azure Services tenant id
    - "--aks-sp-app-id" flag or "AKS_SP_APP_ID" environment variable: Azure Services servicePrincipal app id
    - "--aks-sp-password" flag or "AKS_SP_PASSWORD" environment variable: Azure Services servicePrincipal password


```
silta ci image login [flags]
```

### Options

```
      --image-repo-host string         (Docker) container image repository url
      --gcp-key-json string            Google Cloud service account key (plaintext, json)
      --aws-secret-access-key string   Amazon Web Services IAM account key (string value)
      --aks-tenant-id string           Azure Services tenant id
      --aks-sp-app-id string           Azure Services servicePrincipal app id
      --aks-sp-password string         Azure Services servicePrincipal password
  -h, --help                           help for login
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta ci image](silta_ci_image.md)	 - CI (docker) image commands

