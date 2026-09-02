resource "gooo_service" "checkout" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "catalog" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "billing" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "search" {
  service_type = "http"
  scope        = "prod"
}
