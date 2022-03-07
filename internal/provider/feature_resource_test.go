package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFeatureResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccFeatureResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("devcycle_feature.test", "project_id", "622112634cabe0e9fbaf974d"),
				),
			},
			{
				Config: testAccFeatureResourceConfigEdit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("devcycle_feature.test", "project_id", "622112634cabe0e9fbaf974d"),
					resource.TestCheckResourceAttr("devcycle_feature.test", "description", "Terraform acceptance testing edited"),
				),
			},
			{
				Config:  testAccFeatureResourceConfig,
				Destroy: true,
			},
		},
	})
}

var testAccFeatureResourceConfig = `
resource "devcycle_feature" "test" {
  project_id = "622112634cabe0e9fbaf974d"
  name = "TerraformAccTest` + randString + `"
  key = "terraform-acceptance-testing` + randString + `"
  description = "Terraform acceptance testing"
  type = "experiment"
  tags = ["acceptance-testing"]
}
`

var testAccFeatureResourceConfigEdit = `
resource "devcycle_feature" "test" {
  project_id = "622112634cabe0e9fbaf974d"
  name = "TerraformAccTest` + randString + `"
  key = "terraform-acceptance-testing` + randString + `"
  description = "Terraform acceptance testing edited"
  type = "experiment"
  tags = ["acceptance-testing"]
}
`
