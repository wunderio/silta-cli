package azure

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
}

// OAuth 2 token structure
type tokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    float64 `json:"expires_in"`
	ExtExpiresIn float64 `json:"ext_expires_in"`
}

func GetAzureCredentials(tenantID string) (ctx context.Context, cred *azidentity.DefaultAzureCredential, err error) {

	cred, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{TenantID: tenantID})
	if err != nil {
		return nil, nil, err
	}

	ctx = context.Background()
	return ctx, cred, nil
}

func GetKubeconfig(ctx context.Context, cred azidentity.DefaultAzureCredential, resourceGroupName string, subscriptionID string, clusterName string) (kubeconfig []byte, err error) {

	mcClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, &cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := mcClient.ListClusterAdminCredentials(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Kubeconfigs) > 0 {
		return resp.Kubeconfigs[0].Value, nil
	} else {
		return nil, errors.New("no kubeconfigs are available")
	}

}

// Returns the first subscription ID
// If no IDs are found, 'empty' is set to true
func GetDefaultSubscriptionID(token string) (subscriptionID string, err error) {

	req, err := http.NewRequest(http.MethodGet, "https://management.azure.com/subscriptions?api-version=2020-01-01", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var subscriptions struct {
		Value []Subscription `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return "", err
	}

	if len(subscriptions.Value) == 0 {
		return "", errors.New("no subscriptions found")
	}
	return subscriptions.Value[0].SubscriptionID, nil
}

// Returns access token. Failing that, returns non-nil error
// tenantId - Azure tenant ID
// clientId - Client ID. Can pass Service Principal ID
// clientSecret - Client secret. Cant pass Service Principal password
func GetAuthToken(tenantId string, clientId string, clientSecret string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, "https://login.microsoftonline.com/"+tenantId+"/oauth2/v2.0/token", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	q := req.URL.Query()
	q.Add("grant_type", "client_credentials")
	q.Add("client_id", clientId)
	q.Add("client_secret", clientSecret)
	q.Add("scope", "https://management.azure.com/.default")

	req.Body = io.NopCloser(strings.NewReader(q.Encode()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var token tokenResponse
	if err := json.Unmarshal(str, &token); err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
