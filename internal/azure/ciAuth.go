package azure

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

func GetAzureCredentials(tenantID string) (context.Context, *azidentity.AzureCLICredential) {

	cred, err := azidentity.NewAzureCLICredential(&azidentity.AzureCLICredentialOptions{TenantID: tenantID})

	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	return ctx, cred
}

func GetKubeconfig(ctx context.Context, cred azidentity.AzureCLICredential, resourceGroupName string, subscriptionID string, clusterName string) (kubeconfig []byte) {

	mcClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, &cred, nil)
	if err != nil {
		log.Fatal(err) //remake to return nil on error
	}

	resp, err := mcClient.ListClusterAdminCredentials(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Kubeconfigs) > 0 {
		return resp.Kubeconfigs[0].Value
	} else {
		log.Fatal("Error: No Kubeconfigs are available")
		return
	}

}
