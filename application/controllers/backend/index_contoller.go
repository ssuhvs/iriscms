package backend

import (
	"iriscms/application/models"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"iriscms/common/helper"
)

type IndexController struct {
	Ctx     iris.Context
	Orm     *xorm.Engine
	Session *sessions.Session
}

func (c *IndexController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("ANY", "/index/index", "Index")
	b.Handle("ANY", "/index/menu", "Menu")
	b.Handle("ANY", "/index/main", "Main")
	b.Handle("ANY", "/index/sessionlife", "Sessionlife")
}

func (this *IndexController) Index() {
	roleid, _ := this.Ctx.Values().GetInt64("roleid")
	if roleid == -1 {
		this.Ctx.Redirect("/b/login/index")
		return
	}
	menus := models.NewMenuModel(this.Orm).GetMenu(0, roleid) //读取一级菜单
	this.Ctx.ViewData("menus", menus)
	this.Ctx.ViewData("username", this.Session.Get("username").(string))
	this.Ctx.View("backend/index_index.html")
}

func (this *IndexController) Main() {
	//统计尺寸
	us, _ := disk.Usage(helper.GetRootPath())

	vm, _ := mem.VirtualMemory()

	formatMem := func(mem uint64) string {
		fm := map[int64]string{
			1024:                      "K",
			1024 * 1024:               "MB",
			1024 * 1024 * 1024:        "GB",
			1024 * 1024 * 1024 * 1024: "TB",
		}
		for b, s := range fm {
			res := float64(mem) / float64(b)
			if res > float64(1024) || res < 1 {
				continue
			}
			return strconv.FormatFloat(res, 'f', 2, 64) + s
		}
		return strconv.FormatFloat(float64(mem), 'f', 2, 64) + "Bit"
	}

	//要转换的值，fmt方式，切割长度如果为-1则显示最大长度，64是float64
	siteSize := formatMem(uint64(us.Total))

	this.Ctx.ViewData("SiteSize", siteSize)
	this.Ctx.ViewData("NumCPU", runtime.NumCPU())
	this.Ctx.ViewData("GoVersion", "Version " + strings.ToUpper(runtime.Version()))
	this.Ctx.ViewData("IrisVersion", "Version " + iris.Version)
	this.Ctx.ViewData("Goos", strings.ToUpper(runtime.GOOS))
	this.Ctx.ViewData("Grountues", runtime.NumGoroutine())
	if vm != nil {
		this.Ctx.ViewData("Mem", "总内存:"+formatMem(vm.Total)+",已使用:"+formatMem(vm.Used))
	} else {
		this.Ctx.ViewData("Mem", "未获得内存情况")
	}
	this.Ctx.View("backend/index_main.html")
}

func (this *IndexController) Menu() {
	meid, err := strconv.Atoi(this.Ctx.PostValue("menuid"))
	if err != nil {
		meid = 0
	}
	roleid := this.Ctx.Values().Get("roleid").(int64)
	menus := models.NewMenuModel(this.Orm).GetMenu(int64(meid), roleid) //获取menuid内容
	menujs := []map[string]interface{}{}                                //要返回json的对象
	for _, v := range menus {
		menu := models.NewMenuModel(this.Orm).GetMenu(v.Id, roleid)
		if len(menu) == 0 {
			continue
		}
		sonmenu := []map[string]interface{}{}
		for _, son := range menu {
			sonmenu = append(sonmenu, map[string]interface{}{
				"text": son.Name,
				"id":   son.Id,
				"url":  "/b/" + son.C + "/" + son.A + "?menuid=" + strconv.Itoa(int(son.Id)) + "&" + son.Data,
			})
		}
		menujs = append(menujs, map[string]interface{}{
			"name": v.Name,
			"son":  sonmenu,
		})
	}
	this.Ctx.JSON(menujs)
}

//维持session不过期
func (this *IndexController) Sessionlife() {
	//维持session防止过期
	this.Ctx.WriteString("1")
}
