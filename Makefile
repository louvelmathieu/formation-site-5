GO=go

args = `arg="$(filter-out $@,$(MAKECMDGOALS))" && echo $${arg:-${1}}`

deploy_current:
	eb deploy --staged

build_aws: ## Build for beanstalk
	GOARCH=amd64 GOOS=linux $(GO) build -o bin/application cmd/*.go
