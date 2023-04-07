---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netapp-ontap_protocols_nfs_export_policy_rule_resource Resource - terraform-provider-netapp-ontap"
subcategory: "nasnas"
description: |-
  Export policy rule resource
---

# netapp-ontap_protocols_nfs_export_policy_rule_resource (Resource)

Export policy rule resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `clients_match` (List of String) List of Client Match Hostnames, IP Addresses, Netgroups, or Domains
- `cx_profile_name` (String) Connection profile name
- `export_policy_name` (String) Export policy name
- `vserver` (String) Name of the vserver to use

### Optional

- `allow_device_creation` (Boolean) Allow Creation of Devices
- `allow_suid` (Boolean) Honor SetUID Bits in SETATTR
- `anonymous_user` (String) User ID To Which Anonymous Users Are Mapped
- `chown_mode` (String) Specifies who is authorized to change the ownership mode of a file
- `ntfs_unix_security` (String) NTFS export UNIX security options
- `protocols` (List of String) Access Protocol
- `ro_rule` (List of String) RO Access Rule
- `rw_rule` (List of String) RW Access Rule
- `superuser` (List of String) Superuser Security Types

### Read-Only

- `export_policy_id` (String) Export policy identifier
- `index` (Number) rule index

