# az: Azure Portal shortcuts for Alfred

Open Azure resources under specific subscriptions and resource groups in the Azure Portal via Alfred.

## Usage

[//]: # (TODO add GIF / screenshot)

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

## Acknowledgments

* [alfred-gcloud-shortcuts](https://github.com/jarlefosen/alfred-gcloud-shortcuts) has been used to bootstrap this
  Alfred
  workflow. 
