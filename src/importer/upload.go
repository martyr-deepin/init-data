package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

// FindImagePath return supported image path
func FindImagePath(dir string) []string {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	var r []string
	// 1. find default images.
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(path.Ext(f.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png":
			r = append(r, path.Join(dir, f.Name()))
		default:
			fmt.Printf("image format not support :%q\n", f.Name())
		}
	}
	return r
}

var client = &http.Client{}

// zh_CN: name, descriptions and screenshots(if has any)
func UploadEnglish(w *multipart.Writer, t Item) {
	w.WriteField("lang", "en_US")
	w.WriteField("name", t.NameEN)
	w.WriteField("description", t.DescriptionEN)
}

// zh_CN: name, descriptions and screenshots(if has any)
func UploadChinese(w *multipart.Writer, t Item) {
	w.WriteField("lang", "zh_CN")
	w.WriteField("name", t.NameZH)
	w.WriteField("description", t.DescriptionZH)
}

func writeImage(w *multipart.Writer, field string, fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	fw, err := w.CreateFormFile(field, fpath)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, f)
	if err != nil {
		return err
	}
	return nil
}

func Upload(serverURL string, t Item, baseDir string) error {
	//TODO: split this
	iconPath := path.Join(baseDir, "icons", t.Id+".svg")
	if _, err := os.Stat(iconPath); err != nil {
		return fmt.Errorf("Can't find the icon for %q(%q)\n", t.Id, iconPath)
	}

	fns := []func(*multipart.Writer, Item){
		UploadEnglish,
		UploadChinese,
	}
	for _, fn := range fns {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fn(w, t)
		w.Close()

		err := DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

		if err != nil {
			return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
		}
	}

	// Upload Icon
	{
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("icons", "i0")
		err := writeImage(w, "i0", iconPath)
		if err != nil {
			return fmt.Errorf("Upload.WriteImage(%q) failed: %v\n", iconPath, err)
		}
		w.Close()

		err = DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

		if err != nil {
			return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
		}
	}

	// Upload English screenshots
	enImgs := FindImagePath(path.Join(baseDir, "screenshots", t.Id, "en_US"))
	if len(enImgs) == 0 {
		enImgs = FindImagePath(path.Join(baseDir, "screenshots", t.Id, "en"))
	}
	for i, fpath := range enImgs {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("lang", "en_US")
		field := fmt.Sprintf("s%d", i)
		w.WriteField("screenshots", field)
		err := writeImage(w, field, fpath)
		if err != nil {
			return fmt.Errorf("Upload.WriteImage(%q) failed: %v\n", fpath, err)
		}
		w.Close()

		err = DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

		if err != nil {
			return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
		}
	}

	zhImgs := FindImagePath(path.Join(baseDir, "screenshots", t.Id, "zh_CN"))
	if len(enImgs) == 0 {
		zhImgs = FindImagePath(path.Join(baseDir, "screenshots", t.Id, "zh"))
	}
	for i, fpath := range zhImgs {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("lang", "zh_CN")
		field := fmt.Sprintf("s%d", i)
		w.WriteField("screenshots", field)
		err := writeImage(w, field, fpath)
		if err != nil {
			return fmt.Errorf("Upload.WriteImage(%q) failed: %v\n", fpath, err)
		}
		w.Close()

		err = DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

		if err != nil {
			return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
		}
	}
	for i, fpath := range FindImagePath(path.Join(baseDir, "screenshots", t.Id)) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		w.WriteField("lang", "any")
		field := fmt.Sprintf("s%d", i)
		w.WriteField("screenshots", field)
		err := writeImage(w, field, fpath)
		if err != nil {
			return fmt.Errorf("Upload.WriteImage(%q) failed: %v\n", fpath, err)
		}
		w.Close()

		err = DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

		if err != nil {
			return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
		}
	}
	return nil
}

func DoRequest(serverURL string, id string, r io.Reader, formBoundary string) error {
	req, err := http.NewRequest("POST", serverURL+"/metadata/"+id, r)
	req.Header.Set("Content-Type", formBoundary)
	if err != nil {
		return fmt.Errorf("failed when http.NewRequest:%v\n", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed when http.Client.Do: %v\n", err)
	}
	if res.StatusCode != http.StatusOK {
		fmt.Printf("http.StatusCode not ok: %v\n\t", res.Status)
		io.Copy(os.Stderr, res.Body)
		res.Body.Close()
		return nil
	}
	return nil
}
