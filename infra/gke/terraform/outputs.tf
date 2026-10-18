output "site_ip" {
  value = google_compute_address.site.address
}

output "kube_context" {
  value = "gke_${var.project}_${local.zone}_${google_container_cluster.cka_sim.name}"
}

output "get_credentials" {
  value = "gcloud container clusters get-credentials ${google_container_cluster.cka_sim.name} --zone ${local.zone} --project ${var.project}"
}
