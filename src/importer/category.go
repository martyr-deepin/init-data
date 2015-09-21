package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var category = `
[{"name":"internet","id":"internet","locales":{"zh_CN":{"name":"网络应用"}}},{"name":"office","id":"office","locales":{"zh_CN":{"name":"办公软件"}}},{"name":"development","id":"development","locales":{"zh_CN":{"name":"编程开发"}}},{"name":"reading","id":"reading","locales":{"zh_CN":{"name":"翻译阅读"}}},{"name":"graphics","id":"graphics","locales":{"zh_CN":{"name":"图形图像"}}},{"name":"game","id":"game","locales":{"zh_CN":{"name":"游戏娱乐"}}},{"name":"music","id":"music","locales":{"zh_CN":{"name":"音乐软件"}}},{"name":"system","id":"system","locales":{"zh_CN":{"name":"系统工具"}}},{"name":"video","id":"video","locales":{"zh_CN":{"name":"视频软件"}}},{"name":"chat","id":"chat","locales":{"zh_CN":{"name":"聊天软件"}}},{"name":"others","id":"others","locales":{"zh_CN":{"name":"其他软件"}}}]
`

type Category struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

var Categories map[string]Category = func() map[string]Category {
	var cs []Category
	r := make(map[string]Category)
	json.Unmarshal(([]byte)(category), &cs)
	for _, c := range cs {
		r[c.Id] = c
	}
	return r
}()

func WriteToJSON(items []Item, fpath string) error {
	for _, t := range items {
		_, ok := Categories[t.Category]
		if !ok {
			fmt.Printf("Invalid Category %q in %q\n", t.Category, t.Id)
		}
	}

	f, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("WriteToJSON: %v\n", err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	err = e.Encode(items)
	if err != nil {
		return fmt.Errorf("WriteToJSON.Encode: %v\n", err)
	}
	return nil
}
