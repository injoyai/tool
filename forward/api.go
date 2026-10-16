package forward

import (
	_ "embed"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/frame/fiber"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
)

//go:embed index.html
var index []byte

func Run(port int, filename string, use ...fiber.Middle) error {
	if err := initDB(filename); err != nil {
		return err
	}

	s := fiber.Default()
	s.SetPort(port)
	s.Use(use...)
	s.GET("/", func(c fiber.Ctx) { c.Html200(index) })
	s.Group("/api", func(g fiber.Grouper) {
		g.GET("/forward/all", GetForwardAll)      //列表
		g.POST("/forward", PostForward)           //新建
		g.PUT("/forward", PutForward)             //修改
		g.DELETE("/forward", DelForward)          //删除
		g.PUT("/forward/enable", EnableForward)   //启用
		g.PUT("/forward/disable", DisableForward) //禁用
	})

	return s.Run()
}

var (
	DB    *xorms.Engine
	Cache = maps.NewGeneric[int64, *Forward]()
)

func initDB(filename string) error {
	var err error
	DB, err = sqlite.NewXorm(filename)
	if err != nil {
		return err
	}

	err = DB.Sync(&Forward{})
	if err != nil {
		return err
	}

	all := []*Forward(nil)
	err = DB.Find(&all)
	if err != nil {
		return err
	}

	for _, f := range all {
		f.init()
		Cache.Set(f.ID, f)
	}
	return nil
}

func GetForwardAll(c fiber.Ctx) {
	ls := []*Forward{}
	co, err := DB.Desc("ID").FindAndCount(&ls)
	c.CheckErr(err)
	for _, v := range ls {
		v.resp()
	}
	c.Succ(ls, co)
}

func PostForward(c fiber.Ctx) {
	f := new(Forward)
	c.Parse(f)
	_, err := DB.Insert(f)
	c.CheckErr(err)
	f.init()
	Cache.Set(f.ID, f)
	c.Succ(nil)
}

func PutForward(c fiber.Ctx) {
	f := new(Forward)
	c.Parse(f)
	_, err := DB.Where("ID=?", f.ID).Cols("Name", "Port", "Address").Update(f)
	c.CheckErr(err)
	val, ok := Cache.Get(f.ID)
	if ok {
		val.r.Disable()
		f.Enable = val.Enable
	}
	f.init()
	Cache.Set(f.ID, f)
	c.Succ(nil)
}

func DelForward(c fiber.Ctx) {
	id := c.GetInt64("id")
	_, err := DB.ID(id).Delete(&Forward{})
	c.CheckErr(err)
	val, ok := Cache.GetAndDel(id)
	if ok {
		val.r.Disable()
	}
	c.Succ(nil)
}

func EnableForward(c fiber.Ctx) {
	id := c.GetInt64("id")
	_, err := DB.ID(id).Cols("Enable").Update(&Forward{Enable: true})
	c.CheckErr(err)
	f, ok := Cache.Get(id)
	if ok {
		f.enable()
	}
	c.Succ(nil)
}

func DisableForward(c fiber.Ctx) {
	id := c.GetInt64("id")
	_, err := DB.ID(id).Cols("Enable").Update(&Forward{Enable: false})
	c.CheckErr(err)
	f, ok := Cache.Get(id)
	if ok {
		f.disable()
	}
	c.Succ(nil)
}
