namespace go idgenerator
namespace java idgenerator

service IdGenerator {
  i64 getWorkerId()
  i64 getTimestamp()
  i64 getId(1:string scope)
  i64 getDatacenterId()
  list<string> getScopes()
}
