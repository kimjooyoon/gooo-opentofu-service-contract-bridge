resource "gooo_service" "orders" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "profile" {
  service_type = "http"
}

resource "gooo_service" "reports" {
  service_type = "http"
  scope        = "prod"
}
