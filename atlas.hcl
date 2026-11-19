variable "dev_url" {
  type    = string
  default = coalesce(getenv("ATLAS_DEV_URL"), "docker://postgres/16/dev?search_path=public")
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
