/* Copyright © 2018 VMware, Inc. All Rights Reserved.
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

func resourceNsxtSpoofGuardSwitchingProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtSpoofGuardSwitchingProfileCreate,
		Read:   resourceNsxtSpoofGuardSwitchingProfileRead,
		Update: resourceNsxtSpoofGuardSwitchingProfileUpdate,
		Delete: resourceNsxtSpoofGuardSwitchingProfileDelete,
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
			"address_binding_whitelist_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "When true, this profile overrides the default system wide settings for Spoof Guard when assigned to ports",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceNsxtSpoofGuardSwitchingProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	whiteListProviders := []string{}
	if d.Get("address_binding_whitelist_enabled").(bool) {
		whiteListProviders = append(whiteListProviders, "LPORT_BINDINGS")
	}

	sgSwitchingProfile := manager.SpoofGuardSwitchingProfile{
		Description:        description,
		DisplayName:        displayName,
		Tags:               tags,
		WhiteListProviders: whiteListProviders,
	}

	sgSwitchingProfile, resp, err := nsxClient.LogicalSwitchingApi.CreateSpoofGuardSwitchingProfile(nsxClient.Context, sgSwitchingProfile)

	if err != nil {
		return fmt.Errorf("Error during SpoofGuardSwitchingProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during SpoofGuardSwitchingProfile create: %v", resp.StatusCode)
	}
	d.SetId(sgSwitchingProfile.Id)

	return resourceNsxtSpoofGuardSwitchingProfileRead(d, m)
}

func resourceNsxtSpoofGuardSwitchingProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	sgSwitchingProfile, resp, err := nsxClient.LogicalSwitchingApi.GetSpoofGuardSwitchingProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during SpoofGuardSwitchingProfile read: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] SpoofGuardSwitchingProfile %s not found", id)
		d.SetId("")
		return nil
	}
	d.Set("revision", sgSwitchingProfile.Revision)
	d.Set("description", sgSwitchingProfile.Description)
	d.Set("display_name", sgSwitchingProfile.DisplayName)
	setTagsInSchema(d, sgSwitchingProfile.Tags)
	if len(sgSwitchingProfile.WhiteListProviders) == 1 && sgSwitchingProfile.WhiteListProviders[0] == "LPORT_BINDINGS" {
		d.Set("address_binding_whitelist_enabled", true)
	} else {
		d.Set("address_binding_whitelist_enabled", false)
	}

	return nil
}

func resourceNsxtSpoofGuardSwitchingProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	whiteListProviders := []string{}
	if d.Get("address_binding_whitelist_enabled").(bool) {
		whiteListProviders = append(whiteListProviders, "LPORT_BINDINGS")
	}

	sgSwitchingProfile := manager.SpoofGuardSwitchingProfile{
		Description:        description,
		DisplayName:        displayName,
		Tags:               tags,
		WhiteListProviders: whiteListProviders,
	}

	sgSwitchingProfile, resp, err := nsxClient.LogicalSwitchingApi.UpdateSpoofGuardSwitchingProfile(nsxClient.Context, id, sgSwitchingProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during SpoofGuardSwitchingProfile update: %v", err)
	}

	return resourceNsxtSpoofGuardSwitchingProfileRead(d, m)
}

func resourceNsxtSpoofGuardSwitchingProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.LogicalSwitchingApi.DeleteSwitchingProfile(nsxClient.Context, id, nil)
	if err != nil {
		return fmt.Errorf("Error during SpoofGuardSwitchingProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] SpoofGuardSwitchingProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
