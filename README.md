# tfinline

[![Go CI](https://github.com/wyattfry/tfinline/actions/workflows/test.yaml/badge.svg)](https://github.com/wyattfry/tfinline/actions/workflows/test.yaml)

This is a wrapper around `terraform` to display the output of `init`, `plan`, `apply` and `destroy` with in-line (instead of append) updates.

It also has an 'auto-import' feature, where if it encounters a 'resource already exists' during an apply, it will try to import the resource.

## Vanilla Terraform

```sh
$ terraform destroy -auto-approve
azurerm_resource_group.four: Destroying... [id=.../resourceGroups/acctestwyattrg3]
azurerm_resource_group.three: Destroying... [id=.../resourceGroups/acctestwyattrg3]
azurerm_resource_group.three: Still destroying... [id=.../resourceGroups/acctestwyattrg3, 10s elapsed]
azurerm_resource_group.four: Still destroying... [id=.../resourceGroups/acctestwyattrg3, 10s elapsed]
azurerm_resource_group.three: Destruction complete after 16s
azurerm_resource_group.two: Destroying... [id=.../resourceGroups/acctestwyattrg2]
azurerm_resource_group.four: Destruction complete after 16s
azurerm_resource_group.two: Still destroying... [id=.../resourceGroups/acctestwyattrg2, 10s elapsed]
azurerm_resource_group.two: Destruction complete after 16s
azurerm_resource_group.one: Destroying... [id=.../resourceGroups/acctestwyattrg1]
azurerm_resource_group.one: Still destroying... [id=.../resourceGroups/acctestwyattrg1, 10s elapsed]
azurerm_resource_group.one: Destruction complete after 16s

Destroy complete! Resources: 4 destroyed.
```

## With `tfinline` (in-progress view)

```sh
$ tfinline apply -auto-approve
azurerm_netapp_account.test                                       ✓ Creation complete after 15s
azurerm_netapp_pool.test_secondary                                ✓ Creation complete after 1m8s
azurerm_netapp_pool.test                                          ✓ Creation complete after 1m8s
azurerm_linux_virtual_machine.test                                ✓ Creation complete after 52s
azurerm_linux_virtual_machine.test_secondary                      ✓ Creation complete after 54s
azurerm_netapp_volume_group_sap_hana.test_primary                 ✓ Creation complete after 7m3s
azurerm_netapp_volume_group_sap_hana.test_secondary               ⠧ Creating...
```
