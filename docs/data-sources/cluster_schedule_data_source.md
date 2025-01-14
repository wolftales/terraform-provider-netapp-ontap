---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ontap_cluster_schedule_data_source Data Source - terraform-provider-netapp-ontap"
subcategory: "cluster"
description: |-
  Retrieves Cluster Schedule configuration of SVMs.
---

# Data Source cluster_schedule

Retrieves Cluster Schedule configuration of SVMs.

## Example Usage
```terraform
data "netapp-ontap_cluster_schedule_data_source" "cluster_schedule" {
  cx_profile_name = "cluster4"
  # name = "Application Templates ASUP Dump"
  name = "Balanced Placement Model Cache Update"
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cx_profile_name` (String) Connection profile name
- `name` (String) Schedule name

### Read-Only

- `cron` (Attributes) (see [below for nested schema](#nestedatt--cron))
- `interval` (String) Cluster schedule interval
- `scope` (String) Cluster schedule scope
- `type` (String) Cluster schdeule type
- `id` (String) Cluster schedule UUID

<a id="nestedatt--cron"></a>
### Nested Schema for `cron`

Read-Only:

- `days` (List of Number) List of cluster schedule days
- `hours` (List of Number) List of cluster schedule hours
- `minutes` (List of Number) List of cluster schedule minutes
- `months` (List of Number) List of cluster schedule months
- `weekdays` (List of Number) List of cluster schedule weekdays


