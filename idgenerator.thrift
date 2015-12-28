namespace go idgenerator

service IdGenerator {
  i64 get_worker_id()
  i64 get_timestamp()
  i64 get_id(1:string scope)
  i64 get_datacenter_id()
}
