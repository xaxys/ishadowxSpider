package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	ConfigFile = "config.json"
)

var (
	URL    string
	SSPath string
	Mode   string
	list   []*Server
)

type Server struct {
	IP       string
	Port     string
	Password string
	Method   string
}

func (server *Server) IsComplete() bool {
	return server.IP != "" && server.Port != "" && server.Password != "" && server.Method != ""
}

func (server *Server) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"server":      server.IP,
		"server_port": server.Port,
		"password":    server.Password,
		"method":      server.Method,
		"plugin":      "",
		"plugin_opts": "",
		"plugin_args": "",
		"remarks":     "",
		"timeout":     5,
	}
}

func getConfig() {
	change := false

	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		os.Create(ConfigFile)
	}

	confByte, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		panic("无法打开配置文件！")
	}
	conf := gjson.ParseBytes(confByte)

	url := conf.Get("URL")
	if !url.Exists() {
		URL = getInput("请输入ishadowx的网址")
		confByte, _ = sjson.SetBytes(confByte, "URL", URL)
		change = true
	} else {
		URL = url.String()
	}

	mode := conf.Get("Mode")
	if !mode.Exists() {
		switch Mode {
		case "cover", "add", "none":
		default:
			Mode = getInput("请输入运行模式: [cover/add/none] (覆盖SS配置/在SS配置中追加/不修改SS配置)")
		}
		confByte, _ = sjson.SetBytes(confByte, "Mode", Mode)
		change = true
	} else {
		Mode = mode.String()
	}

	if Mode != "none" {
		sspath := conf.Get("SSPath")
		if !sspath.Exists() {
			SSPath = getInput("请输入SS的文件夹路径")
			confByte, _ = sjson.SetBytes(confByte, "SSPath", SSPath)
			change = true
		} else {
			SSPath = sspath.String()
		}
	}

	if change {
		fmt.Println("正在写入config.json...")
		ioutil.WriteFile(ConfigFile, confByte, 0777)
		fmt.Println("写入完成")
	}

}

func getInput(info string) string {
	var s string
	fmt.Println("----------", info, "----------")
	fmt.Scanln(&s)
	return s
}

func showConfigInfo() {
	fmt.Println("----------[CONFIG]----------")
	fmt.Println("[URL]", URL)
	fmt.Println("[Mode]", Mode)
	fmt.Println("[ShadowsocksPath]", SSPath)
}

func runSpider() {
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		fmt.Println("[ERROR] 获取URL: ", URL, "失败！")
		confByte, err := ioutil.ReadFile(ConfigFile)
		if err != nil {
			panic("无法打开配置文件！")
		}
		confByte, _ = sjson.DeleteBytes(confByte, "URL")
		ioutil.WriteFile(ConfigFile, confByte, 0777)
		panic(err)
	}
	infos := doc.Find("div.portfolio-item div.hover-text")
	infos.Each(func(i int, sel *goquery.Selection) {
		server := &Server{}
		if ip := sel.Find("[id^='ip']").Text(); ip != "" {
			server.IP = ip
		}
		if port := sel.Find("[id^='port']").Text(); port != "" {
			port = strings.Trim(port, "\n")
			server.Port = port
		}
		if pw := sel.Find("[id^='pw']").Text(); pw != "" {
			pw = strings.Trim(pw, "\n")
			server.Password = pw
		}
		sel.Find("h4").Each(func(i int, sel *goquery.Selection) {
			ctx := sel.Text()
			if strings.HasPrefix(ctx, "Method:") {
				mtd := strings.TrimLeft(ctx, "Method:")
				server.Method = mtd
			}
		})

		if server.IsComplete() {
			list = append(list, server)
		}
	})
}

func showServerInfo() {
	for i, server := range list {
		fmt.Println("----------[", i, "]----------")
		fmt.Println("[IP]", server.IP)
		fmt.Println("[Port]", server.Port)
		fmt.Println("[Password]", server.Password)
		fmt.Println("[Method]", server.Method)
	}
}

func editSSConfig() {
	if Mode == "none" {
		return
	}
	GUI_CONFIG := path.Join(SSPath, "gui-config.json")
	confByte, err := ioutil.ReadFile(GUI_CONFIG)
	if err != nil {
		panic("无法打开SS配置文件！")
		confByte, err := ioutil.ReadFile(ConfigFile)
		confByte, _ = sjson.DeleteBytes(confByte, "SSPath")
		ioutil.WriteFile(ConfigFile, confByte, 0777)
		panic(err)
	}
	if Mode == "cover" {
		confByte, _ = sjson.SetBytes(confByte, "configs", []interface{}{})
	}
	for _, server := range list {
		confByte, _ = sjson.SetBytes(confByte, "configs.-1", server.ConvertToMap())
	}
	fmt.Println("正在写入gui-config.json...")
	ioutil.WriteFile(GUI_CONFIG, confByte, 0777)
	fmt.Println("写入完成")
}

func main() {
	getConfig()
	showConfigInfo()
	runSpider()
	showServerInfo()
	editSSConfig()
	getInput("按任意键退出")
}
