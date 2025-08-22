.PHONY: build vet


TARGET_DIR := target
WORKFLOW_FILE := $(TARGET_DIR)/alfred-azure-shortcuts.alfredworkflow


target/:
	mkdir target

clean:
	@rm -rf $(TARGET_DIR)

workflow: build clean target/
	zip -r $(WORKFLOW_FILE) \
	info.plist \
	icon.png \
	update-available.png \
	images \
	bin

install: build
	@echo "Installing workflow to Alfred..."
	@cp -r bin/* $$alfred_preferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/bin
	@cp info.plist $$alfred_preferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/info.plist
	@cp -r images $$alfred_preferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/images
	@cp icon.png $$alfred_preferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/icon.png
	@cp update-available.png $$alfred_preferences/workflows/user.workflow.27E38029-F5F4-4617-A991-F4123720B2CD/update-available.png
	@echo "Workflow installed to Alfred."

build-subscriptions:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/subscriptions cmd/subscriptions/*.go

build-resources:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/resources cmd/resources/*.go

build-resource-groups:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/resource-groups cmd/resource_groups/*.go

build-tenants:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o bin/tenants cmd/tenants/*.go

build: build-subscriptions build-resources build-resource-groups build-tenants
vet:
	go vet ./...

