# Look up a Provider by its ID
data "customerconnect_provider" "example" {
  id = "00000000-0000-0000-0000-000000000001"
}

# Reference computed attributes from the data source
output "provider_name" {
  value = data.customerconnect_provider.example.name
}

output "provider_counts" {
  value = data.customerconnect_provider.example.counts
}
