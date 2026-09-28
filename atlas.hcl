env "local" {
  src = "ent://ent/schema"
  dev = "docker://mysql/8/dev"
  url = getenv("ATLAS_DB_URL")

  migration {
    dir = "file://migrations"
  }
}
