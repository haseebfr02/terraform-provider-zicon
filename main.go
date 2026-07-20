package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/haseebfr02/terraform-provider-zicon/internal/provider"
)

// version is set via -ldflags at release build time; it defaults to "dev"
// for local builds.
var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/haseebfr02/zicon",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
