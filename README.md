# az: Azure Portal shortcuts for Alfred

Open Azure resources under specific subscriptions and resource groups in the Azure Portal via Alfred.

## Usage

![workflow](.github/readme/workflow.gif)

## Installation

Download the latest release [here](https://github.com/trietsch/alfred-azure-shortcuts/releases), and make sure that you
meet the requirements.

## Requirements

Make sure you have
the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli-macos?view=azure-cli-latest) installed.
After installation, run the following command to log in:

```bash
az login
```

This only needs to happen once, because this Alfred workflow uses the same credentials as the Azure CLI.

### Configuration

**hotkey**

Changing the variable `hotkey` from `az` to `azure` results in commands like `azure <query>`.

**always_show_tenant_selection**

By default, when you only have access to a single Azure tenant, the workflow skips the tenant
selection step and goes straight to listing subscriptions. Enable the "Always show tenant
selection" checkbox in the workflow's configuration (or set the `always_show_tenant_selection`
variable to `1`) to always show the tenant selection step, even with a single tenant.

## Acknowledgments

* [alfred-gcloud-shortcuts](https://github.com/jarlefosen/alfred-gcloud-shortcuts) has been used to bootstrap this
  Alfred workflow.
* [maskati/azure-icons](https://github.com/maskati/azure-icons) is used for getting up to date Azure icons.
