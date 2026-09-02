resource "gooo_service" "events" {
  service_type = "queue"
  scope        = "prod"
}

resource "gooo_service" "events-post" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "payments" {
  service_type = "http"
  scope        = "prod"
}

resource "gooo_service" "admin" {
  service_type = "http"
  scope        = "prod"
}
