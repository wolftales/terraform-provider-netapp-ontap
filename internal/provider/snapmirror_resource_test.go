package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSnapmirrorResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test non existant Vol
			{
				Config:      testAccSnapmirrorResourceBasicConfig("snapmirror_dest_svm:testme", "snapmirror_source_svm:testme"),
				ExpectError: regexp.MustCompile("6619337"),
			},
			// Create snapmirror and read
			{
				Config: testAccSnapmirrorResourceBasicConfig("snapmirror_dest_svm:snap_dest", "snapmirror_source_svm:snap"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_resource.example", "destination_endpoint.path", "snapmirror_source_svm:snap"),
				),
			},
		},
	})
}

func testAccSnapmirrorResourceBasicConfig(sourceEndpoint string, destinationEndpoint string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST3")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST3, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_snapmirror_resource" "example" {
  cx_profile_name = "cluster4"
  source_endpoint = {
    path = "%s"
  }
  destination_endpoint = {
    path = "%s"
  }
}`, host, admin, password, sourceEndpoint, destinationEndpoint)
}
