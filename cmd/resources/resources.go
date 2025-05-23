package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	aw "github.com/deanishe/awgo"
	"go.deanishe.net/fuzzy"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger = log.New(os.Stderr, "[resources] ", log.LstdFlags)
	wf     *aw.Workflow

	cred *azidentity.AzureCLICredential
	ctx  = context.Background()

	query             string
	subscriptionId    string
	tenantId          string
	resourceGroupName string
)

type Icon struct {
	Type      string `json:"type"`
	ImagePath string `json:"imagePath"`
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
	flag.StringVar(&resourceGroupName, "resource-group", "", "Azure Resource Group name")

	credential, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		wf.FatalError(err)
	}
	cred = credential
}

func ListResourcesForResourceGroup() (interface{}, error) {
	client, err := armresources.NewClient(subscriptionId, cred, nil)
	if err != nil {
		fmt.Printf("failed to create client: %v\n", err)
		os.Exit(1)
	}

	allResources := make([]armresources.GenericResourceExpanded, 0)

	// List all resources in the specified resource group
	pager := client.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			fmt.Printf("failed to list resources: %v\n", err)
			os.Exit(1)
		}

		for _, resource := range page.Value {
			allResources = append(allResources, *resource)
		}
	}

	return allResources, nil
}

func readAzureIcons() (map[string]string, error) {
	f, err := os.Open("./images/azure_icons.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byt, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var icons []Icon
	if err := json.Unmarshal(byt, &icons); err != nil {
		return nil, err
	}

	iconMap := make(map[string]string)
	for _, icon := range icons {
		iconMap[strings.ToLower(icon.Type)] = "images/" + icon.ImagePath
	}

	return iconMap, nil
}

func getAzurePortalURL(resourceId string) string {
	baseURL := "https://portal.azure.com/#@"
	return fmt.Sprintf("%s%s/resource%s", baseURL, tenantId, resourceId)
}

func run() {
	flag.Parse()

	icons, err := readAzureIcons()
	if err != nil {
		wf.FatalError(err)
		return
	}

	azureResourcesCacheKey := fmt.Sprintf("azure-resources-%s-%s", subscriptionId, resourceGroupName)

	logger.Printf("query=%s", query)
	logger.Printf("subscriptionId=%s", subscriptionId)
	logger.Printf("resourceGroupName=%s", resourceGroupName)

	var resources []armresources.GenericResourceExpanded

	if err := wf.Data.LoadOrStoreJSON(azureResourcesCacheKey, time.Minute*30, ListResourcesForResourceGroup, &resources); err != nil {
		wf.FatalError(err)
		return
	}

	for _, r := range resources {
		iconPath, _ := icons[strings.ToLower(*r.Type)]
		icon := aw.Icon{Value: iconPath}

		url := getAzurePortalURL(*r.ID)

		name := fmt.Sprintf("%s (%s)", *r.Name, *r.Type)
		wf.NewItem(name).
			Icon(&icon).
			Arg(url).
			Subtitle(*r.ID).
			UID(*r.ID).
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
