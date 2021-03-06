/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"net/http"
)

func dataSourceNsxtNsGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtNsGroupRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Unique ID of this resource",
				Optional:    true,
				Computed:    true,
			},
			"display_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The display name of this resource",
				Optional:    true,
				Computed:    true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtNsGroupRead(d *schema.ResourceData, m interface{}) error {
	// Read NS Group by name or id
	nsxClient := m.(*api.APIClient)
	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj manager.NsGroup
	if objID != "" {
		// Get by id
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["populateReferences"] = true
		objGet, resp, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, objID, localVarOptionals)

		if err != nil {
			return fmt.Errorf("Error while reading ns group %s: %v", objID, err)
		}
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("NS group %s was not found", objID)
		}
		obj = objGet
	} else if objName != "" {
		// Get by full name
		// TODO use 2nd parameter localVarOptionals for paging
		objList, _, err := nsxClient.GroupingObjectsApi.ListNSGroups(nsxClient.Context, nil)
		if err != nil {
			return fmt.Errorf("Error while reading NS groups: %v", err)
		}
		// go over the list to find the correct one
		found := false
		for _, objInList := range objList.Results {
			if objInList.DisplayName == objName {
				if found {
					return fmt.Errorf("Found multiple NS groups with name '%s'", objName)
				}
				obj = objInList
				found = true
			}
		}
		if !found {
			return fmt.Errorf("NS group '%s' was not found out of %d groups", objName, len(objList.Results))
		}
	} else {
		return fmt.Errorf("Error obtaining NS group ID or name during read")
	}

	d.SetId(obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
