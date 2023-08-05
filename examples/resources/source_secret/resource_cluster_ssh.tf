# Create Tanzu Mission Control source secret with attached set as default value.
resource "tanzu-mission-control_repository_credential" "create_cluster_source_secret_ssh" {
  name = "tf-secret" # Required

  scope {
    cluster {
      cluster_name            = "testcluster" # Required
      provisioner_name        = "attached"    # Default: attached
      management_cluster_name = "attached"    # Default: attached
    }
  }

  meta {
    description = "Create namespace through terraform"
    labels      = { "key" : "value" }
  }

  spec {
    ssh_key {
      identity    = "testidentity"    # Required
      known_hosts = "testknown_hosts" # Required
    }
  }
}
