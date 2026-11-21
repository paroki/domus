variable "dev_url" {
  type    = string
  default = getenv("ATLAS_DEV_URL") != "" ? getenv("ATLAS_DEV_URL") : "docker://paradedb/paradedb:latest-pg18/dev?search_path=public"
}

env "local" {
  src = "ent://ent/schema"
  dev = var.dev_url
  migration {
    dir = "file://migrations"
  }
}

env "ci" {
  src = "ent://ent/schema"
  dev = var.dev_url
  migration {
    dir = "file://migrations"
  }
}
