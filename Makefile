# Docs
docs_fmt:
	terraform fmt -recursive ./examples/

docs: docs_fmt
	go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.4
	go mod tidy
	tfplugindocs --tf-version=1.7.0 --provider-name=roxywi