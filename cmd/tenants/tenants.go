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
	azureTenantCacheKey = "azure-tenants"
	updateJobName       = "checkForUpdate"
	repo                = "trietsch/alfred-azure-shortcuts"
)

var (
	logger = log.New(os.Stderr, "[tenants] ", log.LstdFlags)
	wf     *aw.Workflow

	doCheck       bool
	iconAvailable = &aw.Icon{Value: "update-available.png"}

	cred *azidentity.AzureCLICredential
	ctx  = context.Background()
)

type Tenant struct {
	TenantID    string `json:"tenantId"`
	DisplayName string `json:"displayName"`
	CountryCode string `json:"countryCode,omitempty"`
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

func ListAzureTenants() (interface{}, error) {
	client, err := armsubscriptions.NewTenantsClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	allTenants := make([]Tenant, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tenants: %w", err)
		}

		for _, tenant := range page.Value {
			displayName := ""
			if tenant.DisplayName != nil {
				displayName = *tenant.DisplayName
			}
			if displayName == "" {
				displayName = *tenant.TenantID
			}

			countryCode := ""
			if tenant.CountryCode != nil {
				countryCode = *tenant.CountryCode
			}

			allTenants = append(allTenants, Tenant{
				TenantID:    *tenant.TenantID,
				DisplayName: displayName,
				CountryCode: countryCode,
			})
		}
	}

	return allTenants, nil
}

func run() {
	wf.Args()
	flag.Parse()
	query := flag.Arg(0)

	var tenants []Tenant

	if err := wf.Data.LoadOrStoreJSON(azureTenantCacheKey, time.Minute*30, ListAzureTenants, &tenants); err != nil {
		wf.NewWarningItem("Failed to list tenants.", "Try 'az login' or check network. Error: "+err.Error()).
			Icon(aw.IconWarning).
			Valid(false)
		wf.FatalError(err)
		return
	}

	// If only one tenant is available, automatically proceed with that tenant
	if len(tenants) == 1 && query == "" {
		// Call subscriptions directly with the single tenant
		exec.Command("./bin/subscriptions", "-tenant-id", tenants[0].TenantID, "-single-tenant", "true")
		return
	}

	for _, t := range tenants {
		subtitle := t.TenantID
		if t.CountryCode != "" {
			subtitle += " (" + t.CountryCode + ")"
		}

		wf.NewItem(t.DisplayName).
			Arg(t.TenantID).
			Subtitle(subtitle).
			UID(t.TenantID).
			Var("selectedTenantId", t.TenantID).
			Valid(true)
	}

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
		wf.NewItem("Update available").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconAvailable)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
