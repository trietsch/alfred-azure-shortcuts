package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
	"log"
	"os"
	"time"
)

const (
	azureSubscriptionCacheKey = "azure-subscriptions"
)

var (
	logger = log.New(os.Stderr, "[subscriptions] ", log.LstdFlags)
	wf     *aw.Workflow

	cred *azidentity.AzureCLICredential
	ctx  = context.Background()
)

type Subscription struct {
	Name           string `json:"name"`
	SubscriptionID string `json:"subscriptionId"`
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(aw.SortOptions(sopts...))

	credential, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		wf.FatalError(err)
	}
	cred = credential
}

func ListAzureSubscriptions() (interface{}, error) {
	client, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	allSubscriptions := make([]Subscription, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			fmt.Printf("failed to list subscriptions: %v\n", err)
			os.Exit(1)
		}

		for _, subscription := range page.Value {
			allSubscriptions = append(allSubscriptions, Subscription{
				Name:           *subscription.DisplayName,
				SubscriptionID: *subscription.SubscriptionID,
			})
		}
	}

	return allSubscriptions, nil
}

func run() {
	args := wf.Args()

	var query string
	if len(args) > 0 {
		query = args[0]
	}

	logger.Printf("query=%s", query)

	var subscriptions []Subscription

	if err := wf.Data.LoadOrStoreJSON(azureSubscriptionCacheKey, time.Minute*30, ListAzureSubscriptions, &subscriptions); err != nil {
		wf.FatalError(err)
		return
	}

	for _, s := range subscriptions {
		wf.NewItem(s.Name).Arg(s.SubscriptionID).Subtitle(s.SubscriptionID).UID(s.SubscriptionID).Valid(true)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
