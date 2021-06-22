module "as3" {
  source     = "./declarations"
  folder     = "AS3"
  validation = true

}

output "bigip1" {
  value = module.as3.declarations.bigip1
}

output "bigip2" {
  value = module.as3.declarations.bigip2
}
