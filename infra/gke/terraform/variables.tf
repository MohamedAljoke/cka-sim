variable "project" {
  type    = string
  default = "cka-sim-gke-maljoke"
}

variable "region" {
  type    = string
  default = "southamerica-east1"
}

variable "lab_machine_type" {
  type    = string
  default = "e2-standard-4"
}

# Each lab takes a node of its own, so this is also how many students can have a lab at once.
variable "max_labs" {
  type    = number
  default = 2
}
