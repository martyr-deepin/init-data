package main

import (
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"os"
	"path"
	"strings"
)

type Item struct {
	Id            string `json:"id"`
	Category      string `json:"category"`
	NameZH        string `json:"name:zh_CN"`
	NameEN        string `json:"name:en_US"`
	SloganZH      string `json:"slogan:zh_CN"`
	SloganEN      string `json:"slogan:en_US"`
	DescriptionZH string `json:"description:zh_CN"`
	DescriptionEN string `json:"description:en_US"`
}

func NewItem(cs []*xlsx.Cell) *Item {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Invalid Row ", cs)
		}
	}()
	t := &Item{
		Id:            cs[1].String(),
		Category:      strings.ToLower(cs[0].String()),
		NameZH:        cs[2].String(),
		NameEN:        cs[3].String(),
		SloganZH:      cs[4].String(),
		SloganEN:      cs[5].String(),
		DescriptionZH: cs[6].String(),
		DescriptionEN: cs[7].String(),
	}
	return t
}

// Convert 将xlsx格式文件转换到JSON格式
func Convert(xl string) ([]Item, error) {
	xlFile, err := xlsx.OpenFile(xl)
	if err != nil {
		return nil, fmt.Errorf("解析xlsx文件出错:%v", err)
	}
	sheet := xlFile.Sheets[0]

	var r []Item
	for _, row := range sheet.Rows[1:] {
		t := NewItem(row.Cells)
		if t != nil {
			r = append(r, *t)
		}
	}
	return r, nil
}

func main() {
	var dataDir = flag.String("i", "", "需要转换的数据目录,包含info.xlsx, screenshots等目录")
	var upload = flag.Bool("upload", false, "上传当前有效的数据. 默认情况下仅进行数据有效性的检测")
	var server = flag.String("server", "", "仓库api的服务器, 用来检测数据有效性以及上传数据")
	var fixSVG = flag.Bool("fix", false, "尝试修复无效的图标")
	var _ids = flag.String("ids", "", "只上传指定的包列表")
	flag.Parse()

	if *dataDir == "" {
		return
	}

	if *fixSVG {
		CheckIcons(*dataDir)
		return
	}

	if *server == "" {
		flag.Usage()
		return
	}

	ts, err := Convert(path.Join(*dataDir, "info.xlsx"))
	if ts == nil || err != nil {
		fmt.Printf("无法读取应用信息:%v\n", err)
		os.Exit(1)
	}

	if *_ids != "" {
		ids := strings.Split(*_ids, ",")
		var r []Item
		for _, t := range ts {
			for _, id := range ids {
				if id == t.Id {
					r = append(r, t)
				}
			}
		}
		ts = r
	} else {
		c := NewChecker(*server)

		ts = c.Filter(ts, *dataDir)

		err = WriteToJSON(ts, path.Join(*dataDir, "info.json"))
		if err != nil {
			fmt.Println("Can't write info.json", err)
		}
	}

	if !*upload {
		return
	}

	for _, t := range ts {
		fmt.Printf("Uploading... %q.......", t.Id)
		err := Upload(*server, t, *dataDir)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
		} else {
			fmt.Printf("successful\n")
		}
	}
}
