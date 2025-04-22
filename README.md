# tfinline

This is a wrapper around `terraform` to display the output of `apply` and `destroy` with in-line (instead of append) updates.

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
$ tfinline destroy
azurerm_resource_group.four   ✅ Destruction complete after 16s
azurerm_resource_group.one    ⠹ Refresh complete
azurerm_resource_group.two    ⠙ Destroying...
azurerm_resource_group.three  ✅ Destruction complete after 20s
```

## With `tfinline` (complete view)

```sh
azurerm_resource_group.four   ✅ Destruction complete after 16s
azurerm_resource_group.one    ✅ Destruction complete after 16s
azurerm_resource_group.two    ✅ Destruction complete after 16s
azurerm_resource_group.three  ✅ Destruction complete after 20s
```
