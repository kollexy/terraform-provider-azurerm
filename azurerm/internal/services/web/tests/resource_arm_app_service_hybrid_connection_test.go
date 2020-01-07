package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMAppServiceHybridConnection_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_app_service_hybrid_connection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMAppServiceHybridConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMAppServiceHybridConnection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMAppServiceHybridConnectionExists(data.ResourceName),
				),
			},
		},
	})
}

func testCheckAzureRMAppServiceHybridConnectionDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Web.AppServicesClient

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_app_service_hybrid_connection" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resGroup := rs.Primary.Attributes["resource_group_name"]
		namespaceName := rs.Primary.Attributes["namespace_name"]
		relayName := rs.Primary.Attributes["relay_name"]
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext
		resp, err := client.GetHybridConnection(ctx, resGroup, name, namespaceName, relayName)

		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return nil
	}
	return nil
}

func testCheckAzureRMAppServiceHybridConnectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		resGroup := rs.Primary.Attributes["resource_group_name"]
		namespaceName := rs.Primary.Attributes["namespace_name"]
		relayName := rs.Primary.Attributes["relay_name"]

		conn := acceptance.AzureProvider.Meta().(*clients.Client).Web.AppServicesClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext
		resp, err := conn.GetHybridConnection(ctx, resGroup, name, namespaceName, relayName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: App Service Hybrid Connection %q (resource group: %q) does not exist", name, resGroup)
			}

			return fmt.Errorf("Bad: GetHybridConnection on appServicesClient: %+v", err)
		}

		return nil
	}
}

func testAccAzureRMAppServiceHybridConnection_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_app_service_plan" "test" {
  name                = "acctestASP-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  sku {
    tier = "Standard"
    size = "S1"
  }
}

resource "azurerm_app_service" "test" {
  name                = "acctestAS-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  app_service_plan_id = azurerm_app_service_plan.test.id
}

resource "azurerm_relay_namespace" "test" {
  name                = "acctestrn-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  sku_name = "Standard"
}

resource "azurerm_relay_hybrid_connection" "test" {
	name                 = "acctestrnhc-%d"
	resource_group_name  = azurerm_resource_group.test.name
	relay_namespace_name = azurerm_relay_namespace.test.name
	user_metadata        = "metadatatest"
  }

resource "azurerm_servicebus_namespace" "test" {
  name                = "acctestservicebusnamespace-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "basic"
}

resource "azurerm_app_service_hybrid_connection" "test" {
	app_service_name      = azurerm_app_service.test.name
	resource_group_name   = azurerm_resource_group.test.name
	namespace_name        = azurerm_relay_hybrid_connection.test.name
	relay_name            = azurerm_relay_hybrid_connection.test.name
	relay_arm_uri         = azurerm_relay_hybrid_connection.test.id
	hostname              = "testhostname.azuretest"
	port                  = 80
	service_bus_namespace = azurerm_servicebus_namespace.test.name
	send_key_name         = "RootManageSharedAccessKey"
	send_key_value        = azurerm_servicebus_namespace.test.default_primary_key	
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}
