package linode

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/linode/linodego"
)

func TestAccLinodeSSHKey_basic(t *testing.T) {
	t.Parallel()

	resName := "linode_sshkey.foobar"
	var sshkeyName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeSSHKeyConfigBasic(sshkeyName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeSSHKeyExists,
					resource.TestCheckResourceAttr(resName, "label", sshkeyName),
					resource.TestCheckResourceAttr(resName, "ssh_key", publicKeyMaterial),
					resource.TestCheckResourceAttrSet(resName, "created"),
				),
			},

			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLinodeSSHKey_update(t *testing.T) {
	t.Parallel()
	resName := "linode_sshkey.foobar"
	var sshkeyName = acctest.RandomWithPrefix("tf_test")
	publicKeyMaterial, _, err := acctest.RandSSHKeyPair("linode@ssh-acceptance-test")
	if err != nil {
		t.Fatalf("Cannot generate test SSH key pair: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeSSHKeyConfigBasic(sshkeyName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeSSHKeyExists,
					resource.TestCheckResourceAttr(resName, "label", sshkeyName),
					resource.TestCheckResourceAttr(resName, "ssh_key", publicKeyMaterial),
					resource.TestCheckResourceAttrSet(resName, "created"),
				),
			},
			{
				Config: testAccCheckLinodeSSHKeyConfigUpdates(sshkeyName, publicKeyMaterial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeSSHKeyExists,
					resource.TestCheckResourceAttr(resName, "label", fmt.Sprintf("%s_renamed", sshkeyName)),
					resource.TestCheckResourceAttr(resName, "ssh_key", publicKeyMaterial),
					resource.TestCheckResourceAttrSet(resName, "created"),
				),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckLinodeSSHKeyExists(s *terraform.State) error {
	client := testAccProvider.Meta().(linodego.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_sshkey" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.ID)
		}

		_, err = client.GetSSHKey(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Error retrieving state of SSHKey %s: %s", rs.Primary.Attributes["label"], err)
		}
	}

	return nil
}

func testAccCheckLinodeSSHKeyDestroy(s *terraform.State) error {
	client, ok := testAccProvider.Meta().(linodego.Client)
	if !ok {
		return fmt.Errorf("Error getting Linode client")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_sshkey" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing %v to int", rs.Primary.ID)
		}
		if id == 0 {
			return fmt.Errorf("Would have considered %v as %d", rs.Primary.ID, id)

		}

		_, err = client.GetSSHKey(context.Background(), id)

		if err == nil {
			return fmt.Errorf("Linode SSH Key with id %d still exists", id)
		}

		if apiErr, ok := err.(*linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Error requesting Linode SSH Key with id %d", id)
		}
	}

	return nil
}

func testAccCheckLinodeSSHKeyConfigBasic(label, sshkey string) string {
	return fmt.Sprintf(`
resource "linode_sshkey" "foobar" {
	label = "%s"
	ssh_key = "%s"
}`, label, sshkey)
}

func testAccCheckLinodeSSHKeyConfigUpdates(label, sshkey string) string {
	return fmt.Sprintf(`
resource "linode_sshkey" "foobar" {
	label = "%s_renamed"
	ssh_key = "%s"
}`, label, sshkey)
}
