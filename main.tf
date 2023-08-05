  // Tanzu Mission Control EKS Cluster Type: AWS EKS clusters.
// Operations supported : Read, Create, Update & Delete

// Create Tanzu Mission Control AWS EKS cluster entry

terraform {
  required_providers {
    tanzu-mission-control = {
      source = "vmware/dev/tanzu-mission-control"
    }
  }
}
terraform {
  backend "local" {
    path = "./terraform.tfstate"
  }
}
provider "tanzu-mission-control" {
}
resource "tanzu-mission-control_ekscluster" "demo_eks_cluster" {
  credential_name = "gargi-cred-721" // Required
  region          = "us-east-2"          // Required
  name            = "gargi-test-tf-1"        // Required

  ready_wait_timeout = "30m" // Wait time for cluster operations to finish (default: 30m).

  meta {
    description = "description of the cluster"
  }

  spec {
    cluster_group = "default" // Default: default

    config {
      role_arn           = "arn:aws:iam::666099245364:role/control-plane.13805125522349961624.eks.tmc.cloud.vmware.com" // Required, forces new
      kubernetes_version = "1.26"                // Required

      kubernetes_network_config {
        service_cidr = "172.31.0.0/16" // Forces new
      }

      logging {
        api_server         = false
        audit              = true
        authenticator      = true
        controller_manager = false
        scheduler          = true
      }

      vpc { // Required
        enable_private_access = true
        enable_public_access  = true
        public_access_cidrs = [
          "0.0.0.0/0",
        ]
        subnet_ids = [ // Forces new
          "subnet-0c4b51095bff4f982",
          "subnet-0ba1bfcb1278604db",
          "subnet-01f7bfbb0545baefa",
          "subnet-082f9488467c5c60f"
        ]
      }
    }
    nodepool {
      info {
        name        = "nodepool-1" // Required
        description = "description of node pool"
      }

      spec {
        role_arn    = "arn:aws:iam::666099245364:role/worker.13805125522349961624.eks.tmc.cloud.vmware.com" // Required
        // tags        = { "<key>" : "<value>" }
        // node_labels = { "<key>" : "<value>" }

        subnet_ids = [ // Forces new
          "subnet-0c4b51095bff4f982",
          "subnet-0ba1bfcb1278604db",
          "subnet-01f7bfbb0545baefa",
          "subnet-082f9488467c5c60f"
        ]

        scaling_config {
          desired_size = 4
          max_size     = 8
          min_size     = 1
        }

        release_version = "1.26.6-20230728"

        update_config {
          max_unavailable_percentage = "12"
        }
      }
    }
  }
}
