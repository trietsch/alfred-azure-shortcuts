package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"go.deanishe.net/fuzzy"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	azureSubscriptionCacheKey = "azure-subscriptions"
	updateJobName             = "checkForUpdate"
	repo                      = "trietsch/alfred-azure-subscriptions"
)

var (
	logger = log.New(os.Stderr, "[subscriptions] ", log.LstdFlags)
	wf     *aw.Workflow

	doCheck       bool
	iconAvailable = &aw.Icon{Value: "update-available.png"}

	cred *azidentity.AzureCLICredential
	ctx  = context.Background()
)

type Subscription struct {
	Name           string `json:"name"`
	SubscriptionID string `json:"subscriptionId"`
	TenantID       string `json:"tenantId"`
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	updateOpt := update.GitHub(repo)
	wf = aw.New(updateOpt, aw.SortOptions(sopts...))
	flag.BoolVar(&doCheck, "check", false, "check for a new version")

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
				TenantID:       *subscription.TenantID,
			})
		}
	}

	return allSubscriptions, nil
}

func run() {
	wf.Args()
	flag.Parse()
	query := flag.Arg(0)

	if doCheck {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")

		cmd := exec.Command(os.Args[0], "-check")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	logger.Printf("query=%s", query)

	if query == "" && wf.UpdateAvailable() {
		// Turn off UIDs to force this item to the top.
		// If UIDs are enabled, Alfred will apply its "knowledge"
		// to order the results based on your past usage.
		wf.Configure(aw.SuppressUIDs(true))

		// Notify user of update. As this item is invalid (Valid(false)),
		// actioning it expands the query to the Autocomplete value.
		// "workflow:update" triggers the updater Magic Action that
		// is automatically registered when you configure Workflow with
		// an Updater.
		//
		// If executed, the Magic Action downloads the latest version
		// of the workflow and asks Alfred to install it.
		wf.NewItem("Update available!").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconAvailable)
	}

	var subscriptions []Subscription

	if err := wf.Data.LoadOrStoreJSON(azureSubscriptionCacheKey, time.Minute*30, ListAzureSubscriptions, &subscriptions); err != nil {
		wf.FatalError(err)
		return
	}

	for _, s := range subscriptions {
		wf.NewItem(s.Name).
			Arg(s.SubscriptionID).
			Subtitle(s.SubscriptionID).
			UID(s.SubscriptionID).
			Var("tenantId", s.TenantID).
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
