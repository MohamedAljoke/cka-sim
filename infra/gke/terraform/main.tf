terraform {
  required_version = ">= 1.11"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.7"
    }
  }
}

provider "google" {
  project = var.project
  region  = var.region
}

locals {
  zone = "${var.region}-a"
}

resource "google_container_cluster" "cka_sim" {
  name     = "cka-sim"
  location = local.zone

  # The pools are their own resources below, so the default one goes as soon as the cluster exists.
  remove_default_node_pool = true
  initial_node_count       = 1

  # Dataplane v2 enforces the NetworkPolicy that keeps the lab reachable only from the app.
  datapath_provider   = "ADVANCED_DATAPATH"
  deletion_protection = false
}

resource "google_container_node_pool" "system" {
  name     = "system"
  cluster  = google_container_cluster.cka_sim.id
  location = local.zone

  autoscaling {
    min_node_count = 1
    max_node_count = 2
  }

  node_config {
    machine_type = "e2-medium"
    disk_size_gb = 30
  }
}

# The lab pod's node: made when Start lab creates the pod, gone about 10 minutes after Stop lab. The taint keeps
# everything else off it, and the pod tolerates it and asks for pool=labs.
resource "google_container_node_pool" "labs" {
  name               = "labs"
  cluster            = google_container_cluster.cka_sim.id
  location           = local.zone
  initial_node_count = 0

  autoscaling {
    min_node_count = 0
    max_node_count = 1
  }

  node_config {
    machine_type = var.lab_machine_type
    # kind in the lab pod needs the kernel modules and cgroups of a regular Linux, not Container-Optimized OS.
    image_type   = "UBUNTU_CONTAINERD"
    disk_size_gb = 50
    labels       = { pool = "labs" }

    taint {
      key    = "lab"
      value  = "true"
      effect = "NO_SCHEDULE"
    }
  }
}

resource "google_compute_address" "site" {
  name   = "cka-sim-site"
  region = var.region
}
