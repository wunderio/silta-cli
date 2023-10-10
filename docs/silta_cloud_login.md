## silta cloud login

Kubernetes cluster login

### Synopsis

Log in to kubernetes cluster using different methods:
	
	* Kubeconfig for custom cluster access 
	Requires:
	  - "--kube-config" flag or "KUBECTL_CONFIG" environment variable
	  
	* Google cloud GKE access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--gcp-key-json" flag or "GCLOUD_KEY_JSON" environment variable
	  - "--gcp-project-name" flag or "GCLOUD_PROJECT_NAME" environment variable
	
	Optional parameters:
	  - "--gcp-compute-region" flag or "GCLOUD_COMPUTE_REGION" environment variable
	  - "--gcp-compute-zone" flag or "GCLOUD_COMPUTE_ZONE" environment variable
	  
	* Amazon Web Services EKS access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--aws-secret-access-key" flag or "AWS_SECRET_ACCESS_KEY" environment variable
	  - "--aws-region" flag or "AWS_REGION" environment variable

	* Azure Services AKS access
	Requires:
	  - "--cluster-name" flag or "CLUSTER_NAME" environment variable
	  - "--aks-resource-group" flag or "AKS_RESOURCE_GROUP" environment variable
	  - "--aks-tenant-id" flag or "AKS_TENANT_ID" environment variable
	  - "--aks-sp-app-id" flag or "AKS_SP_APP_ID" environment variable
	  - "--aks-sp-password" flag or "AKS_SP_PASSWORD" environment variable
	

```
silta cloud login [flags]
```

### Options

```
      --aks-resource-group string      Azure Services resource group (this is not the AKS RG)
      --aks-sp-app-id string           Azure Services servicePrincipal app id
      --aks-sp-password string         Azure Services servicePrincipal password
      --aks-tenant-id string           Azure Services tenant id
      --aws-region string              Amazon Web Services resource region
      --aws-secret-access-key string   Amazon Web Services IAM account key (string value)
      --cluster-name string            Kubernetes cluster name
      --gcp-compute-region string      GCP compute region
      --gcp-compute-zone string        GCP compute zone
      --gcp-key-json string            Google Cloud service account key (plaintext, json)
      --gcp-key-path string            Location of Google Cloud service account key file
      --gcp-project-name string        GCP project name (project id)
  -h, --help                           help for login
      --kubeconfig string              Kubernetes config content (plaintext, base64 encoded)
      --kubeconfigpath string          Kubernetes config path (default "~/.kube/config")
```

### Options inherited from parent commands

```
      --debug     Print variables, do not execute external commands, rather print them
      --use-env   Use environment variables for value assignment (default true)
```

### SEE ALSO

* [silta cloud](silta_cloud.md)	 - Kubernetes cloud related commands

