package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"go.deanishe.net/fuzzy"
	"log"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
)

const (
	azureSubscriptionCacheKey = "azure-subscriptions"
)

var (
	logger = log.New(os.Stderr, "[subscriptions] ", log.LstdFlags)
	wf     *aw.Workflow

	argRefreshSubscriptions bool
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

	flag.BoolVar(&argRefreshSubscriptions, "refresh", false, "refresh subscriptions")
}

func ListAzureSubscriptions(ctx context.Context, cred *azidentity.AzureCLICredential) ([]Subscription, error) {
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
	flag.Parse()
	ctx := context.Background()

	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		wf.FatalError(err)
	}

	if argRefreshSubscriptions {
		wf.Configure(aw.TextErrors(true))
		logger.Printf("refreshing subscriptions")
		projects, err := ListAzureSubscriptions(ctx, cred)
		if err != nil {
			wf.FatalError(err)
		}
		err = wf.Data.StoreJSON(azureSubscriptionCacheKey, projects)
		if err != nil {
			wf.FatalError(err)
		}
		logger.Printf("refresh done")
		wf.SendFeedback()
		return
	}

	var query string
	if len(args) > 0 {
		query = args[0]
	}

	if strings.HasPrefix(query, "-") {
		wf.NewItem("Refresh subscriptions").Arg("-refresh").Autocomplete("-refresh").Valid(false)
		wf.SendFeedback()
		return
	}

	logger.Printf("query=%s", query)
	var subscriptions []Subscription
	if !wf.Data.Exists(azureSubscriptionCacheKey) {
		wf.Fatal(`No subscriptions cached, please run "-refresh"`)
	}
	if err := wf.Data.LoadJSON(azureSubscriptionCacheKey, &subscriptions); err != nil {
		wf.FatalError(err)
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
