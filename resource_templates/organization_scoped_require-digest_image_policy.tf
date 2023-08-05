// Create/ Delete/ Update Tanzu Mission Control organization scoped require-digest image policy entry
resource "tanzu-mission-control_image_policy" "organization_scoped_require-digest_image_policy" {
  name = "<image-policy-name>"

  scope {
    organization {
      organization = "<organization-id>" // Required
    }
  }

  spec {
    input {
      require_digest {
        audit = false // Default: false
      }
    }

    namespace_selector {
      match_expressions {
        key      = "<label-selector-requirement-key-1>"
        operator = "<label-selector-requirement-operator>"
        values = [
          "<label-selector-requirement-value-1>",
          "<label-selector-requirement-value-2>"
        ]
      }
      match_expressions {
        key      = "<label-selector-requirement-key-2>"
        operator = "<label-selector-requirement-operator>"
        values   = []
      }
    }
  }
}
