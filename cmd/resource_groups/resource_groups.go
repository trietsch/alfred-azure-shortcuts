package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"go.deanishe.net/fuzzy"
	"log"
	"os"
	"time"

	aw "github.com/deanishe/awgo"
)

var (
	logger = log.New(os.Stderr, "[resource-groups] ", log.LstdFlags)
	wf     *aw.Workflow

	cred *azidentity.AzureCLICredential
	ctx  = context.Background()

	query          string
	subscriptionId string
	tenantId       string
)

type ResourceGroup struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.SortOptions(sopts...))

	flag.StringVar(&query, "query", "", "Query input by the user, to filter the list on")
	flag.StringVar(&subscriptionId, "subscription", "", "Azure Subscription ID")
	flag.StringVar(&tenantId, "tenant-id", "", "Azure Tenant ID")

	credential, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		wf.FatalError(err)
	}
	cred = credential
}

func ListResourceGroupsForSubscription() (interface{}, error) {
	client, err := armresources.NewResourceGroupsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	allResourceGroups := make([]ResourceGroup, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list resource groups: %w", err)
		}

		for _, rg := range page.Value {
			allResourceGroups = append(allResourceGroups, ResourceGroup{
				Name: *rg.Name,
				Id:   *rg.ID,
			})
		}
	}

	return allResourceGroups, nil
}

func run() {
	flag.Parse()

	azureResourceGroupsCacheKey := "azure-resource-groups-" + subscriptionId

	logger.Printf("query=%s", query)

	var resourceGroups []ResourceGroup

	if err := wf.Data.LoadOrStoreJSON(azureResourceGroupsCacheKey, time.Minute*30, ListResourceGroupsForSubscription, &resourceGroups); err != nil {
		wf.NewWarningItem("Failed to list resource groups.", "Try 'az login' or check network. Error: "+err.Error()).
			Icon(aw.IconWarning).
			Valid(false)
		// wf.FatalError(err) // Original line
		wf.SendFeedback() // Send the warning
		return           // Exit after sending warning
	}

	for _, rg := range resourceGroups {
		logger.Printf("%+v", rg)
		wf.NewItem(rg.Name).
			Arg(rg.Name).
			Subtitle(rg.Id).
			UID(rg.Id).
			Var("subscription", subscriptionId).
			Var("tenantId", tenantId).
			Valid(true)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
