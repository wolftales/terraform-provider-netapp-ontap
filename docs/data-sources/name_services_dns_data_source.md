---
page_title: "ONTAP: DNS"
subcategory: "name-services"
description: |-
  DNS data source
---

# Data Source DNS

Retrieves the DNS Configuration of an SVM

## Example Usage
```terraform
data "netapp-ontap_name_services_dns_data_source" "name_services_dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm_name = "ansibleSVM_cifs"
}
```


<!-- schema generated by tfplugindocs -->
## Argument Reference

### Required

- `cx_profile_name` (String) Connection profile name

### Optional

- `svm_name` (String) IPInterface vserver name

### Read-Only

- `dns_domains` (List of String) List of DNS domains such as 'sales.bar.com'. The first domain is the one that the Vserver belongs to
- `name_servers` (List of String) List of IPv4 addresses of name servers such as '123.123.123.123'.

