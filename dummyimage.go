package main

import (
	"fmt"
	"github.com/golang/freetype"
	"github.com/jie123108/glog"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
)

func ParseSize(url string) (width, height int) {
	re := regexp.MustCompile(`(\d+)x(\d+)`)
	matched := re.FindStringSubmatch(url)
	width = 400
	height = 300
	if len(matched) == 3 {
		i, err := strconv.Atoi(matched[1])
		if err == nil {
			width = i
		}
		i, err = strconv.Atoi(matched[2])
		if err == nil {
			height = i
		}
	}
	return
}

func ParseColor(color string) (r, g, b int) {
	fmt.Sscanf(color, "%2x%2x%2x", &r, &g, &b)
	return
}

func DrawImage(rw io.Writer, width, height int, bgcolor string,
	title string, title_size int, content string, content_size int, fgcolor string, fmt string) error {
	bg_r, bg_g, bg_b := 0, 100, 0
	if len(bgcolor) > 0 {
		bg_r, bg_g, bg_b = ParseColor(bgcolor)
	}
	bg := color.RGBA{uint8(bg_r), uint8(bg_g), uint8(bg_b), 255}

	//创建一个图像
	img := image.NewRGBA(image.Rect(0, 0, width, height)) //*NRGBA (image.Image interface)
	// 填充蓝色,并把其写入到img
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	// 解析前景色
	fg_r, fg_g, fg_b := 100, 0, 0
	if len(fgcolor) > 0 {
		fg_r, fg_g, fg_b = ParseColor(fgcolor)
	}

	src := &image.Uniform{color.RGBA{uint8(fg_r), uint8(fg_g), uint8(fg_b), 255}}

	font_file := "fonts/FZFSJW.TTF"
	data, err := ioutil.ReadFile(font_file)
	if err != nil {
		glog.Errorf("ReadFile(%s) failed! err: %s", font_file, err)
		return err
	}
	f, err := freetype.ParseFont(data)
	if err != nil {
		glog.Errorf("ParseFont(%s) failed! err: %s", font_file, err)
		return err
	}

	fontDPI := 72.0
	c := freetype.NewContext()
	c.SetDPI(fontDPI)
	c.SetDst(img)
	c.SetClip(img.Bounds())
	c.SetSrc(src)
	c.SetFont(f)
	c.SetFontSize(float64(title_size))

	// x := 10 + int(c.PointToFixed(float64(title_size))>>8)
	// glog.Infof("x: %d", x)
	pt := freetype.Pt(20, title_size)
	_, err = c.DrawString(title, pt)
	if err != nil {
		glog.Errorf("DrawString(%s) failed! err:", title, err)
		return err
	}

	if len(content) > 0 {
		c.SetFontSize(float64(content_size))
		_, err = c.DrawString(content, freetype.Pt(20, title_size+content_size+height/16))
		if err != nil {
			glog.Errorf("DrawString(%s) failed! err:", content, err)
			return err
		}
	}

	if fmt == "jpg" || fmt == "jpeg" {
		q := &jpeg.Options{100}
		jpeg.Encode(rw, img, q) //Encode writes the Image img to w in PNG format.
	} else {
		png.Encode(rw, img)
	}
	glog.Infof("Draw Image(size: %dx%d, title: %s, content: %s",
		width, height, title, content)
	return nil
}

func args_get(args url.Values, argname, defvalue string) string {
	value := args.Get(argname)
	if value == "" {
		value = defvalue
	}
	return value
}

func args_get_int(args url.Values, argname string, defvalue int) int {
	str := args_get(args, argname, "")
	value := defvalue
	var err error
	if str != "" {
		value, err = strconv.Atoi(str)
		if err != nil {
			glog.Errorf("invalid params %s's [%s]", argname, str)
			value = defvalue
		}
	}
	return value
}

func DrawImageHandler(rw http.ResponseWriter, req *http.Request) {
	url := req.URL
	if url.Path == "/favicon.ico" {
		return
	}

	args := url.Query()
	// 解析宽高
	ext := "png"
	req_ext := path.Ext(url.Path)
	if len(req_ext) > 0 {
		ext = req_ext[1:]
	}
	width, height := ParseSize(url.Path)
	// 解析背景颜色,前景颜色
	bgcolor := args_get(args, "bgcolor", "006400")
	fontcolor := args_get(args, "fontcolor", "242424")
	title := args_get(args, "title", fmt.Sprintf("%dx%d", width, height))
	title_size := args_get_int(args, "title_size", 60)
	content := args_get(args, "content", url.Path)
	content_size := args_get_int(args, "content_size", 30)

	DrawImage(rw, width, height, bgcolor, title, title_size,
		content, content_size, fontcolor, ext)
}

// 大家可以查看这个网址看看这个image包的使用方法 http://golang.org/doc/articles/image_draw.html
func main() {
	http.HandleFunc("/", DrawImageHandler)
	http.ListenAndServe(":8421", nil)
}
