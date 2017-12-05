package db

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	redis "gopkg.in/redis.v5"
)

var (
	DataSource  orm.Ormer
	AppWish     orm.Ormer
	RedisClient *redis.Client
)

func Setup() {

	RedisClient = createClient()
	local, err := time.LoadLocation("Asia/Shanghai")

	if err != nil {
		panic(err)
	}
	time.Local = local
	if beego.BConfig.RunMode == "dev" {
		err := orm.RegisterDataBase("tdefault", "mysql", "root:123456@tcp(192.168.12.137:3306)/source?charset=utf8&parseTime=true&loc=Asia%2FShanghai", 30, 30)
		if err != nil {
			panic(err)
		}
		err = orm.RegisterDataBase("default", "mysql", "root:123456@tcp(192.168.12.137:3306)/app_wish?charset=utf8&parseTime=true&loc=Asia%2FShanghai", 30, 30)
		if err != nil {
			panic(err)
		}
	} else {
		err := orm.RegisterDataBase("tdefault", "mysql", "root:123456@tcp(127.0.0.1:3306)/source?charset=utf8&parseTime=true&loc=Asia%2FShanghai", 30, 30)
		if err != nil {
			panic(err)
		}
		err = orm.RegisterDataBase("default", "mysql", "root:123456@tcp(127.0.0.1:3306)/app_wish?charset=utf8&parseTime=true&loc=Asia%2FShanghai", 30, 30)
		if err != nil {
			panic(err)
		}
	}
	DataSource = orm.NewOrm()
	DataSource.Using("tdefault")
	AppWish = orm.NewOrm()
	AppWish.Using("default")
}

func createClient() *redis.Client {
	var client *redis.Client
	if beego.BConfig.RunMode == "dev" {
		client = redis.NewClient(&redis.Options{
			Addr:     "192.168.12.137:6379",
			Password: "",
			DB:       0,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	_, err := client.Ping().Result()
	if err != nil {
		if err != nil {
			panic(err)
		}
	}

	return client
}
