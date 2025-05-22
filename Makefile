.PHONY: build


TARGET_DIR := target
WORKFLOW_FILE := $(TARGET_DIR)/alfred-azure-shortcuts.alfredworkflow


target/:
	mkdir target

clean:
	@[ -d $(TARGET_DIR) ] && rm -r $(TARGET_DIR) || true

workflow: build clean target/
	zip $(WORKFLOW_FILE) \
	info.plist \
	icon.png \
	bin/subscriptions \
	bin/resources \
	bin/resource-groups

install: build
	@echo "Installing workflow to Alfred..."
	@cp -r bin/* /Users/robintrietsch/dotfiles/preferences/alfred/Alfred.alfredpreferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/bin
	@cp info.plist /Users/robintrietsch/dotfiles/preferences/alfred/Alfred.alfredpreferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/info.plist
	@echo "Workflow installed to Alfred."

build-subscriptions:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/subscriptions cmd/subscriptions/*.go

build-resources:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/resources cmd/resources/*.go

build-resource-groups:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/resource-groups cmd/resource_groups/*.go

build: build-subscriptions build-resources build-resource-groups
