package helper

//避免循环调用错误,公用非依赖变量以函数方式返回
import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"reflect"

	"github.com/kataras/iris/context"
	"github.com/nfnt/resize"
	"golang.org/x/image/draw"
)

//GetRootPath 获取IRIS项目根目录 (即 main.go的所在位置)
func GetRootPath() string {
	pwd, _ := os.Getwd()
	return pwd
}

var location, _ = time.LoadLocation("Local")

//Krand 随机字符串
func Krand(size int, kind int) []byte {
	ikind, kinds, result := kind, [][]int{{10, 48}, {26, 97}, {26, 65}}, make([]byte, size)
	isAll := kind > 2 || kind < 0
	//随机种子 (如果不以时间戳作为时间种子, 可能每次生成的随机数每次都相同)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if isAll {
			// random ikind
			ikind = rand.Intn(3)
		}
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return result
}

//GetMd5 md5加密字符串
func GetMd5(str string) string {
	md := md5.New()
	md.Write([]byte(str))
	return hex.EncodeToString(md.Sum(nil))
}

//Ajax Ajax返回数据给前端
func Ajax(errmsg interface{}, errcode int64, this context.Context) {
	this.JSON(map[string]interface{}{
		"errcode": errcode,
		"errmsg":  errmsg,
	})
}

//GetTimeStamp 获取时间戳
func GetTimeStamp() int {
	timestamp := time.Now().In(location).Unix()
	return int(timestamp)
}

//NowDate 当前时间 Y m d H:i:s
func NowDate(str string) string {
	t := time.Now().In(location)
	str = strings.Replace(str, "Y", "2006", 1)
	str = strings.Replace(str, "m", "01", 1)
	str = strings.Replace(str, "d", "02", 1)
	str = strings.Replace(str, "H", "13", 1)
	str = strings.Replace(str, "i", "04", 1)
	str = strings.Replace(str, "s", "05", 1)
	return t.Format(str)
}

//FormatTime 时间戳格式化时间
func FormatTime(timestamp int64) string {
	t := time.Unix(timestamp, 0).In(location)
	str := "Y-m-d H:i:s"
	str = strings.Replace(str, "Y", "2006", 1)
	str = strings.Replace(str, "m", "01", 1)
	str = strings.Replace(str, "d", "02", 1)
	str = strings.Replace(str, "H", "13", 1)
	str = strings.Replace(str, "i", "04", 1)
	str = strings.Replace(str, "s", "05", 1)
	return t.Format(str)
}

//GetImg 根据图片路径生成图片,待优化函数
func GetImg(path, waterPath string) {
	fmt.Println("开始处理图片")
	f, err := os.Open(path) //打开文件
	if err != nil {
		fmt.Println("打开文件失败:", err.Error())
		return
	}
	defer f.Close()
	filename := strings.Split(f.Name(), ".")
	if len(filename) != 2 || (filename[1] != "jpg" && filename[1] != "jpeg" && filename[1] != "gif" && filename[1] != "png") {
		return
	}
	fmt.Println("解析文件信息：", filename)

	var imager image.Image
	if filename[1] == "jpg" {
		imager, err = jpeg.Decode(f)
	} else if filename[1] == "jpeg" {
		imager, err = jpeg.Decode(f)
	} else if filename[1] == "gif" {
		imager, err = gif.Decode(f)
	} else if filename[1] == "png" {
		imager, err = png.Decode(f)
	}
	if err != nil {
		fmt.Println("打开文件失败:", err.Error())
		return
	}

	fmt.Println("解码文件:", imager)

	//获取图片缩略图
	thumbnail := resize.Thumbnail(120, 120, imager, resize.Lanczos3)
	fileThumb, err := os.Create(filename[0] + strconv.Itoa(int(time.Now().Unix())) + "_thmub.jpg")
	if err == nil {
		jpeg.Encode(fileThumb, thumbnail, &jpeg.Options{
			Quality: 80})
		fileThumb.Close()
		fmt.Println("生成缩略图片成功")
	}
	rectangler := imager.Bounds()
	fmt.Println("获取图片的0点和尾点:", rectangler)
	//创建画布
	newWidth := 200
	m := image.NewRGBA(image.Rect(0, 0, newWidth, newWidth*rectangler.Dy()/rectangler.Dx()))
	//在画布上绘制图片 m画布 m.bounds画布参数, imager 要参照打开的图片信息 image.Point 图片绘制的其实地址 绘制资源
	draw.Draw(m, m.Bounds(), imager, image.Point{100, 100}, draw.Src)
	//绘制水印图

	//必须是PNG图片

	warter, wterr := os.Open(waterPath)
	if wterr == nil {
		//无错误的时候解码
		watermark, dewaerr := png.Decode(warter)
		if dewaerr == nil {
			//无错误的时候添加水印
			draw.Draw(m, watermark.Bounds().Add(image.Pt(30, 30)), watermark, image.ZP, draw.Over)
		} else {
			fmt.Println("水印图片解码失败")
		}
	} else {
		fmt.Println("水印图片打开失败")
	}
	toimg, _ := os.Create(filename[0] + strconv.Itoa(GetTimeStamp()) + "-120-80.jpg") //创建文件系统
	defer toimg.Close()
	//toimg 保存的名称 要参照的画布，图片选项。默认透明图
	jpeg.Encode(toimg, m, &jpeg.Options{
		Quality: jpeg.DefaultQuality}) //保存为jpeg图片
}

//MultiUpload 多图上传生成html内容
func MultiUpload(data []string) string {
	upload := "/public/upload"
	box := ""
	if len(data) > 0 {
		for _, v := range data {
			box += `<div class="imgbox">
					<input
					  class="imgbox_inputBtn"
					  type="image"
					  onclick="return doUpload(this)"
					  src="` + v + `" alt="点击上传" />
					<input type="hidden" value="` + v + `" name="info[thumb][]" />
				</div>`
		}
	} else {
		box += `<div class="imgbox">
				<input
				class="imgbox_inputBtn"
				type="image"
				onclick="return doUpload(this)"
				src=""
				alt="点击上传" />
				<input type="hidden" value="" name="info[thumb][]" />
		    	</div>`
	}

	str := `
		<style>
		.imgbox{
			display: inline-block;
			margin-right: 10px;
		}
		.imgbox_inputBtn {
			width:50px;
			height:50px;
			display:block;
			border:1px solid #ddd;
			padding:2px
		}
		.add_imgbox_inputBtn{
			width:50px;
			height:56px;
			display:inline;
			border:1px solid #ddd;
			padding:2px;
			font-size: 54px;
			outline: none
		}
		</style>
		` + box + `
		<div class="imgbox" onclick="return createHtml(this)">
		  <input
		  class="add_imgbox_inputBtn"
		  type="button"
		  value="+" alt="点击添加上传框"/>
		</div>


<script >
    function doUpload(obj){
            $.upload({
                url:"` + upload + `",
                fileName: 'formfile',
                params: {},
                dataType: 'json',
                onSend: function() {
                    return true;
                },
                onSubmit: function(){
                },
                onComplate: function(data) {
                    if (data.errcode == 1) {
			layer.msg(data.msg)
		    }else{
			$(obj).next().val(data.msg);
			$(obj).attr('src',data.msg);
		    }	
                }
            });
            return false;
    }

    function createHtml(obj){
        var str="<div class='imgbox'>\
                  <input class='imgbox_inputBtn' type='image' onclick='return doUpload(this)' src='' alt='点击上传'/>\
                  <input type='hidden' name='info[thumb][]' />\
            </div>";
            $(obj).before(str);
        return false;
    }

</script>
`
	return str
}

//SiginUpload 单图上传
func SiginUpload(data, field string) string {
	upload := "/public/upload"
	html := `<input onclick="` + field + `doUpload(this)" type="image" src="` + data + `" onerror='this.src="https://placeholdit.imgix.net/~text?txtsize=12&txt=暂无图片&w=100&h=100"' alt="点击上传" style="width:100px;height:100px;display:block;border:1px solid #ddd;padding:2px;float:left;" />
			 <input type="hidden"  value="` + data + `" name="` + field + `" />
			 <span style='color:#fff;display:inline-block;width:15px;height:15px;font-size:15px;line-height:15px;text-align:center;background:rgba(0,0,0,0.5);font-weight:normal;cursor:pointer' onclick="` + field + `DelImg(this)">×</span>
			 `
	html += `
	<script>
		function ` + field + `doUpload(obj) {
			$.upload({
				url:"` + upload + `",
				fileName: 'formfile',
				dataType: 'json',
				onSend: function (){return true;},
				onSubmit:function(){return true;},
				onComplate: function(data) {
					if (data.errcode == 1) {
						layer.msg(data.errmsg);
					}else{
						$(obj).next().val(data.errmsg);
						$(obj).attr('src',data.errmsg);
					}
				}
			});
		}

		function ` + field + `DelImg(obj) {
		 layer.confirm('要删除此图片吗？', {
			  btn: ['确定','取消']
		 }, function(){
		  $(obj).prev().val("");
		  $(obj).prev().prev().attr('src',"https://placeholdit.imgix.net/~text?txtsize=12&txt=暂无图片&w=100&h=100");
		  layer.closeAll();
		 }, function(){});
		}
	</script>
	`
	return html
}

//Password 生成密码
func Password(password, encrypt string) string {
	return GetMd5(GetMd5(password) + encrypt)
}

//IsFalse 检测字段是否为 空 0 nil
func IsFalse(args ...interface{}) bool {
	for _, v := range args {
		switch v.(type) {
		case string:
			if v == nil || v == "" {
				return false
			}
		case int, int64, int8, int32:
			if v == nil || v == 0 {
				return false
			}
		default:
			if v == nil {
				return false
			}

		}
	}
	return false
}

//IsError 检测是否有Error
func IsError(args ...error) bool {
	for _, v := range args {
		if v != nil {
			return true
		}
	}
	return false
}

//IsAjax 判断是否是ajax
func IsAjax(this context.Context) bool {
	return this.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

//Struct2Map 结构体转map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj) //得到变量的类型
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); /*取得字段长度*/ i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}
