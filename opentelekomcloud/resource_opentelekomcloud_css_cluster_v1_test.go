package opentelekomcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/opentelekomcloud/gophertelekomcloud"
)

func TestAccCssClusterV1_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCssClusterV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCssClusterV1_basic(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCssClusterV1Exists(),
				),
			},
		},
	})
}

func testAccCssClusterV1_basic(val string) string {
	return fmt.Sprintf(`
resource "opentelekomcloud_networking_secgroup_v2" "secgroup" {
  name = "terraform_test_security_group%s"
  description = "terraform security group acceptance test"
}

resource "opentelekomcloud_css_cluster_v1" "cluster" {
  expect_node_num = 1
  name = "terraform_test_cluster%s"
  node_config {
    flavor = "css.medium.8"
    network_info {
      security_group_id = opentelekomcloud_networking_secgroup_v2.secgroup.id
      network_id = "%s"
      vpc_id = "%s"
    }
    volume {
      volume_type = "COMMON"
      size = 40
    }
    availability_zone = "%s"
  }
}
	`, val, val, OS_NETWORK_ID, OS_VPC_ID, OS_AVAILABILITY_ZONE)
}

func testAccCheckCssClusterV1Destroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	client, err := config.cssV1Client(OS_REGION_NAME)
	if err != nil {
		return fmt.Errorf("Error creating sdk client, err=%s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opentelekomcloud_css_cluster_v1" {
			continue
		}

		url, err := replaceVarsForTest(rs, "clusters/{id}")
		if err != nil {
			return err
		}
		url = client.ServiceURL(url)

		_, err = client.Get(url, nil, &golangsdk.RequestOpts{
			MoreHeaders: map[string]string{"Content-Type": "application/json"}})
		if err == nil {
			return fmt.Errorf("opentelekomcloud_css_cluster_v1 still exists at %s", url)
		}
	}

	return nil
}

func testAccCheckCssClusterV1Exists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		client, err := config.cssV1Client(OS_REGION_NAME)
		if err != nil {
			return fmt.Errorf("Error creating sdk client, err=%s", err)
		}

		rs, ok := s.RootModule().Resources["opentelekomcloud_css_cluster_v1.cluster"]
		if !ok {
			return fmt.Errorf("Error checking opentelekomcloud_css_cluster_v1.cluster exist, err=not found this resource")
		}

		url, err := replaceVarsForTest(rs, "clusters/{id}")
		if err != nil {
			return fmt.Errorf("Error checking opentelekomcloud_css_cluster_v1.cluster exist, err=building url failed: %s", err)
		}
		url = client.ServiceURL(url)

		_, err = client.Get(url, nil, &golangsdk.RequestOpts{
			MoreHeaders: map[string]string{"Content-Type": "application/json"}})
		if err != nil {
			if _, ok := err.(golangsdk.ErrDefault404); ok {
				return fmt.Errorf("opentelekomcloud_css_cluster_v1.cluster is not exist")
			}
			return fmt.Errorf("Error checking opentelekomcloud_css_cluster_v1.cluster exist, err=send request failed: %s", err)
		}
		return nil
	}
}
