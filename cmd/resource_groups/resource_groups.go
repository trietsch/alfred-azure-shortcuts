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

const (
	azureResourceGroupsCacheKey = "azure-resource-groups"
)

var (
	logger = log.New(os.Stderr, "[resource-groups] ", log.LstdFlags)
	wf     *aw.Workflow

	cred           *azidentity.AzureCLICredential
	ctx            = context.Background()
	subscriptionId string
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

	flag.StringVar(&subscriptionId, "query", "", "Query input by the user, to filter the list on")
	flag.StringVar(&subscriptionId, "subscription", "", "Azure Subscription ID")

	credential, err := azidentity.NewAzureCLICredential(nil)
	cred = credential
	if err != nil {
		wf.FatalError(err)
	}
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
			log.Fatalf("failed to list resource groups: %v", err)
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
	args := wf.Args()
	flag.Parse()

	var query string
	if len(args) > 0 {
		query = args[0]
	}
	logger.Printf("query=%s", query)

	var resourceGroups []ResourceGroup

	if err := wf.Data.LoadOrStoreJSON(azureResourceGroupsCacheKey, time.Minute*30, ListResourceGroupsForSubscription, &resourceGroups); err != nil {
		wf.FatalError(err)
		return
	}

	for _, rg := range resourceGroups {
		logger.Printf("%+v", rg)
		wf.NewItem(rg.Name).Arg(rg.Id).Subtitle(rg.Id).UID(rg.Id).Valid(true)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
