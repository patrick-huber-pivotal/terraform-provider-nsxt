/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"net/http"
)

func resourceNsxtDhcpRelayProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtDhcpRelayProfileCreate,
		Read:   resourceNsxtDhcpRelayProfileRead,
		Update: resourceNsxtDhcpRelayProfileUpdate,
		Delete: resourceNsxtDhcpRelayProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
			"server_addresses": &schema.Schema{
				Type:        schema.TypeSet,
				Description: "Set of dhcp relay server addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
				Required: true,
			},
		},
	}
}

func resourceNsxtDhcpRelayProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	serverAddresses := getStringListFromSchemaSet(d, "server_addresses")
	dhcpRelayProfile := manager.DhcpRelayProfile{
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		ServerAddresses: serverAddresses,
	}

	dhcpRelayProfile, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateDhcpRelayProfile(nsxClient.Context, dhcpRelayProfile)

	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during DhcpRelayProfile create: %v", resp.StatusCode)
	}
	d.SetId(dhcpRelayProfile.Id)

	return resourceNsxtDhcpRelayProfileRead(d, m)
}

func resourceNsxtDhcpRelayProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	dhcpRelayProfile, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelayProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile read: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayProfile %s not found", id)
		d.SetId("")
		return nil
	}

	d.Set("revision", dhcpRelayProfile.Revision)
	d.Set("description", dhcpRelayProfile.Description)
	d.Set("display_name", dhcpRelayProfile.DisplayName)
	setTagsInSchema(d, dhcpRelayProfile.Tags)
	d.Set("server_addresses", dhcpRelayProfile.ServerAddresses)

	return nil
}

func resourceNsxtDhcpRelayProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	serverAddresses := interface2StringList(d.Get("server_addresses").(*schema.Set).List())
	dhcpRelayProfile := manager.DhcpRelayProfile{
		Revision:        revision,
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		ServerAddresses: serverAddresses,
	}

	dhcpRelayProfile, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateDhcpRelayProfile(nsxClient.Context, id, dhcpRelayProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during DhcpRelayProfile update: %v", err)
	}

	return resourceNsxtDhcpRelayProfileRead(d, m)
}

func resourceNsxtDhcpRelayProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteDhcpRelayProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
