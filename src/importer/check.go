package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

// 1. 所有id以info.xlsx文件中为主
// 2. 检测是否所有pid都有对应的screenshot, icons, name, description等
// 3. 检测是否有多余的icons
// 4. 检测是否有多余的screenshots
// 5. 检测是否有pid在仓库中没有
type Checker struct {
	items     []string
	serverURL string
	validIDs  map[string]bool
	targetIDs map[string]bool
}

func (c *Checker) loadDSC() {
	resp, err := http.Get(c.serverURL + "/dsc")
	if err != nil {
		fmt.Printf("无法获得服务器状态: %s\n", c.serverURL)
		os.Exit(1)
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	var playload struct {
		StatusCode int      `json:"status_code"`
		Data       []string `json:"data"`
	}
	d.Decode(&playload)
	if playload.StatusCode != 0 {
		fmt.Printf("服务器状态错误: %q %v\n", c.serverURL, playload)
		os.Exit(1)
	}
	for _, id := range playload.Data {
		c.validIDs[id] = true
	}
}

func NewChecker(serverURL string) *Checker {
	c := &Checker{
		serverURL: serverURL,
		validIDs:  make(map[string]bool),
	}
	c.loadDSC()
	return c
}

func (c *Checker) setupTargets(items []Item) {
	c.targetIDs = make(map[string]bool)
	for _, t := range items {
		c.targetIDs[t.Id] = true
	}
}

// HitDSC return true if this id is in repository
func (c Checker) HitDSC(id string) bool {
	return c.validIDs[id]
}

// HitTarget return true if this id is in target set of c.Filter
func (c Checker) HitTarget(id string) bool {
	return c.targetIDs[id]
}

func (c Checker) FilterLost(items []Item, baseDir string) ([]Item, int) {
	//TODO: split this
	var r []Item
	var invalidIds []string
	var lostIcons []string
	var lostScreenshots []string
	var lostCovers []string
	for _, t := range items {
		invalid := false
		if !c.HitDSC(t.Id) {
			invalidIds = append(invalidIds, t.Id)
			invalid = true
		}

		iconPath := path.Join(baseDir, "icons", t.Id+".svg")
		if _, err := os.Stat(iconPath); err != nil {
			lostIcons = append(lostIcons, t.Id)
			invalid = true
		}

		coverPath := path.Join(baseDir, "covers", t.Id+".jpg")
		if _, err := os.Stat(coverPath); err != nil {
			lostCovers = append(lostCovers, t.Id)
			invalid = true
		}

		screenDir := path.Join(baseDir, "screenshots", t.Id)
		if _, err := os.Stat(screenDir); err != nil {
			lostScreenshots = append(lostScreenshots, t.Id)

			invalid = true
		}
		if !invalid {
			r = append(r, t)
		}
	}

	if len(invalidIds) != 0 {
		fmt.Printf("共%d个应用在仓库中无对应软件包, 这些应用信息不会被自动导入到服务器\n%v\n\n",
			len(invalidIds), invalidIds)
	}

	if len(lostIcons) != 0 {
		fmt.Printf("共%d个应用没有找到的图标文件, 这些应用信息不会自动导入到服务器\n%v\n\n",
			len(lostIcons), lostIcons)
	}

	if len(lostScreenshots) != 0 {
		fmt.Printf("共%d个应用没有找到截图目录, 这些应用信息不会自动导入到服务器\n%v\n\n",
			len(lostScreenshots), lostScreenshots)
	}
	if len(lostCovers) != 0 {
		fmt.Printf("共%d个应用没有找到封面图目录, 这些应用信息不会自动导入到服务器\n%v\n\n",
			len(lostCovers), lostCovers)
	}

	return r, len(invalidIds) + len(lostIcons) + len(lostScreenshots) + len(lostCovers)
}

func (c *Checker) Filter(items []Item, baseDir string) []Item {
	c.setupTargets(items)

	r, n1 := c.FilterLost(items, baseDir)
	n2 := c.checkUseless(baseDir)
	ShowCow(n1, n2)

	return r
}

func (c *Checker) checkUseless(baseDir string) int {
	var uselessIcons []string
	{
		fs, err := ioutil.ReadDir(path.Join(baseDir, "icons"))
		if err != nil {
			fmt.Println("无法找到图标目录")
		}
		for _, f := range fs {
			name := f.Name()
			id := name[:len(name)-len(path.Ext(name))]

			if !c.HitTarget(id) {
				uselessIcons = append(uselessIcons, f.Name())
			}
		}
		if len(uselessIcons) != 0 {
			fmt.Printf("共%d个多余图标:\n%v\n\n", len(uselessIcons), uselessIcons)
		}
	}

	var uselessScreenshot []string
	{
		fs, err := ioutil.ReadDir(path.Join(baseDir, "screenshots"))
		if err != nil {
			fmt.Println("无法找到截图目录")
		}
		for _, f := range fs {
			if !c.HitTarget(f.Name()) {
				uselessScreenshot = append(uselessScreenshot, f.Name())
			} else {
				c.WarningScreenshotLang(path.Join(baseDir, "screenshots", f.Name()))
			}
		}
		if len(uselessScreenshot) != 0 {
			fmt.Printf("共%d个多余截图目录:\n%v\n\n", len(uselessScreenshot), uselessScreenshot)
		}
	}

	var uselessCovers []string
	{
		fs, err := ioutil.ReadDir(path.Join(baseDir, "covers"))
		if err != nil {
			fmt.Println("无法找到封面目录")
		}
		for _, f := range fs {
			name := f.Name()
			id := name[:len(name)-len(path.Ext(name))]
			if !c.HitTarget(id) {
				uselessCovers = append(uselessCovers, f.Name())
			}
		}
		if len(uselessCovers) != 0 {
			fmt.Printf("共%d个多余封面图目录:\n%v\n\n", len(uselessCovers), uselessCovers)
		}
	}

	return len(uselessScreenshot) + len(uselessIcons) + len(uselessCovers)
}

func (c *Checker) WarningScreenshotLang(imgDir string) {
	defaultN := len(FindImagePath(imgDir))
	enN := len(FindImagePath(path.Join(imgDir, "en")))
	enUSN := len(FindImagePath(path.Join(imgDir, "en_US")))
	zhN := len(FindImagePath(path.Join(imgDir, "zh")))
	zhCNN := len(FindImagePath(path.Join(imgDir, "zh_CN")))

	any := false
	if defaultN == 0 {
		fmt.Printf("%q中无默认图片\n", imgDir)
		any = true
	}
	if enN != 0 && enUSN == 0 {
		fmt.Printf("%q中发现en目录，请使用en_US\n", imgDir)
		any = true
	}
	if zhN != 0 && zhCNN == 0 {
		fmt.Printf("%q中发现zh目录，请使用zh_CN\n", imgDir)
		any = true
	}
	if defaultN+enN+enUSN+zhN+zhCNN == 0 {
		fmt.Printf("%q中没有任何有效的截图文件\n", imgDir)
		any = true
	}
	if any {
		fmt.Println("")
	}
}

func (t Item) Valid() bool {
	return true
}

func validSVG(fpath string) bool {
	bs, err := exec.Command("inkscape", "-z", fpath).CombinedOutput()
	if err != nil {
		//can't find the inkscape command
		return true
	}
	return !bytes.Contains(bs, ([]byte)("error"))
}

// 1. move all files into targetDir
// 2. use inkscape to fix invalid svg file
func FixSVG(file string, backupDir string) error {
	os.MkdirAll(backupDir, 0755)
	newFile := path.Join(backupDir, path.Base(file))
	err := os.Rename(file, newFile)
	if err != nil {
		return err
	}
	cmd := exec.Command("inkscape", "-z", "-l", file, newFile)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func CheckIcons(baseDir string) bool {
	baseDir = path.Join(baseDir, "icons")
	fs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic("Can't find icons directory " + baseDir)
	}

	var invalidSVG []string
	for _, finfo := range fs {
		if strings.ToLower(path.Ext(finfo.Name())) == ".svg" {
			fpath := path.Join(baseDir, finfo.Name())
			if !validSVG(fpath) {
				invalidSVG = append(invalidSVG, fpath)
				fmt.Println("InvalidSVG:", fpath)
			}
		}
	}
	if len(invalidSVG) != 0 {
		fmt.Printf("发现%d个有问题的svg文件,尝试修复中\n", len(invalidSVG))
		for _, fpath := range invalidSVG {
			backupDir := path.Join(baseDir, "invalids")
			err := FixSVG(fpath, backupDir)
			fmt.Printf("修复%s中...旧文件备份到%s\n", fpath, path.Join(backupDir, path.Base(fpath)))
			if err != nil {
				fmt.Printf("修复%s失败:%v..\n", fpath, err)
			}
		}
	}
	return true
}
