variable "namespace" {
    type = string
}

# e.g.: bhojpur.net
variable "dns_zone_domain" {
    type = string
}

# e.g.: my-branch.staging.bhojpur.net
variable "domain" {
    type = string
}

# e.g.: ["", "*.", "*.ws."]
variable "subdomains" {
    type = list(string)
}

variable "public_ip" {
    type = string
}

variable "cert_namespace" {
    type = string
    default = "certs"
}