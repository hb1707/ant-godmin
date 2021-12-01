package main

import (
    "antGodmin/orm"
    "antGodmin/routers"
    "antGodmin/setting"
    "log"
)

func init() {
    orm.OpenDB()
}
func main() {
    var confApp = setting.App
    if confApp.RUNMODE == "dev" {
        router := routers.List(false)
        err := router.Run(":8081")
        if err != nil {
            log.Fatal(err)
        }
    } else {
        router := routers.List(true)
        err := router.Run(":80")
        if err != nil {
            log.Fatal(err)
        }
    }
}
