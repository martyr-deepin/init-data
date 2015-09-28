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
			fmt.Printf("image format not support :%q\n", path.Join(dir, f.Name()))
		}
	}
	return r
}

var client = &http.Client{}

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

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("name", t.NameEN)
	w.WriteField("description", t.DescriptionEN)

	w.WriteField("name:en_US", t.NameEN)
	w.WriteField("description:en_US", t.DescriptionEN)

	w.WriteField("name:zh_CN", t.NameZH)
	w.WriteField("description:zh_CN", t.DescriptionZH)
	w.Close()

	err := DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

	if err != nil {
		return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
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

	// Upload  screenshots
	for _, lang := range []string{"en_US", "zh_CN", ""} {
		imageDir := path.Join(baseDir, "screenshots", t.Id, lang)
		for i, fpath := range FindImagePath(imageDir) {

			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			field := fmt.Sprintf("s%d", i)
			if lang != "" {
				w.WriteField("screenshots:"+lang, field)
			} else {
				w.WriteField("screenshots", field)
			}
			err := writeImage(w, field, fpath)
			fmt.Println("try upload image... ", fpath, field)
			if err != nil {
				return fmt.Errorf("Upload.WriteImage(%q) failed: %v\n", fpath, err)
			}
			w.Close()

			err = DoRequest(serverURL, t.Id, &b, w.FormDataContentType())

			if err != nil {
				return fmt.Errorf("Upload.DoRequest(%q) failed: %v\n", t.Id, err)
			}
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
