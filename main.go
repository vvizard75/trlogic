package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/labstack/echo"
	"html/template"
	"image/gif"
	"image/jpeg"
	"io/ioutil"
	"mime"
	"net/http"
	"os/signal"
	"time"
)
import (
	"fmt"
	"github.com/labstack/echo/middleware"
	"github.com/nfnt/resize"
	"github.com/rs/xid"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const thumbSuffix = "_100x100"
const storePath = "./images"
const thumbPath = "./thumbs"
const port = "8081"

type Template struct {
	templates *template.Template
}

type Img struct {
	Img string
	Tmb string
}

func main() {
	initEnv()
	t := &Template{templates: template.Must(template.ParseGlob("./html/*.html"))}
	server := echo.New()
	server.Renderer = t
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	server.Static("/images", "images")
	server.Static("/thumbs", "thumbs")

	server.POST("/upload", chooseHandler)
	server.GET("/", gallery)

	server.Logger.Fatal(server.Start(":" + port))

	go func() {
		if err := server.Start(":" + port); err != nil {
			server.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		server.Logger.Fatal(err)
	}

}
func gallery(c echo.Context) error {
	gallery := listDir(storePath)
	return c.Render(http.StatusOK, "gallery", gallery)
}

func listDir(path string) []Img {
	var gallery []Img
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("Error list gallery: %s\n", err)
	}
	exts := map[string]bool{".gif": true, ".png": true, ".jpg": true, ".jpeg": true}
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && exts[ext] {
			img := new(Img)
			img.Img = storePath + "/" + file.Name()
			fname := strings.TrimSuffix(file.Name(), ext)
			img.Tmb = thumbPath + "/" + fname + thumbSuffix + ext
			gallery = append(gallery, *img)
		}

	}
	return gallery
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func chooseHandler(c echo.Context) error {
	ct := c.Request().Header.Get(echo.HeaderContentType)
	ct, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return c.HTML(http.StatusBadRequest, "Error parse media type")
	}
	switch ct {
	default:
		return c.HTML(http.StatusBadRequest, "Error parse Content-Type")
	case echo.MIMEMultipartForm:
		getFilesFromForm(c)
	case echo.MIMEApplicationJSON:
		getFileFromJson(c)
	case echo.MIMEApplicationForm:
		getFileFromPath(c)

	}
	c.Redirect(http.StatusOK, "/")
	return nil
}

func getFilesFromForm(c echo.Context) {

	var files []*multipart.FileHeader

	form, err := c.MultipartForm()
	if err == nil {
		files = form.File["files"]
	} else {

		fmt.Printf("Error parse multipartform: %s\n", err)
	}

	file, err2 := c.FormFile("file")
	if err2 == nil {
		files = append(files, file)
	} else {

		fmt.Printf("Error parse multipartform: %s\n", err2)
	}
	if err != nil && err2 != nil {
		result := new(Result)
		result.Error = true
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		return
	}

	var results []Result
	for _, f := range files {
		result := new(Result)
		id := xid.New().String()
		filename := filepath.Base(f.Filename)
		result.Filename, err = makePicFromFileHeader(f, id)

		if err != nil {
			result.Error = true
		} else {
			result.Error = false
		}
		result.Name = filename
		//result.Filename=id+"_"+result.Name
		results = append(results, *result)
	}
	c.JSON(http.StatusOK, results)
}
func getFileFromJson(c echo.Context) {

	type Image struct {
		Name    string `json:"name"`
		ImgData string `json:"img_data"`
	}
	result := new(Result)
	result.Error = true

	img := new(Image)
	err := c.Bind(img)
	if err != nil {
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		fmt.Printf("Error parse JSON: %s\n", err)
		return
	}
	result.Name = filepath.Base(img.Name)

	rawData, err := base64.StdEncoding.DecodeString(img.ImgData)
	if err != nil {
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		fmt.Printf("Error encode BASE64: %s\n", err)
		return
	}
	reader := bytes.NewReader(rawData)
	id := xid.New().String()
	result.Filename, err = makePic(reader, img.Name, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		fmt.Printf("Error encode  image base64: %s\n", err)
		return
	}
	//result.Filename=id+"_"+result.Name
	result.Error = false
	c.JSON(http.StatusOK, [1]Result{*result})
}
func getFileFromPath(c echo.Context) {

	path := c.FormValue("path")

	result := new(Result)
	result.Name = filepath.Base(path)
	result.Error = true
	resp, err := http.Get(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		fmt.Printf("Error download image from path: %s\n", err)
		return
	}
	defer resp.Body.Close()
	id := xid.New().String()
	result.Filename, err = makePic(resp.Body, result.Name, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, [1]Result{*result})
		fmt.Printf("Error encode  image path: %s\n", err)
		return
	}
	result.Error = false
	//result.Filename=id+"_"+result.Name
	c.JSON(http.StatusOK, [1]Result{*result})
}

func makePicFromFileHeader(f *multipart.FileHeader, id string) (string, error) {
	file, err := f.Open()
	if err != nil {
		fmt.Printf("Error open file from form: %s\n", err)
		return "", err
	}
	defer file.Close()
	return makePic(file, filepath.Base(f.Filename), id)
}

func makePic(r io.Reader, fileName string, id string) (string, error) {
	_fileName := id + "_" + fileName
	var buf bytes.Buffer
	r = io.TeeReader(r, &buf)

	img, t, err := image.Decode(r)
	if err != nil {
		fmt.Printf("Error decode image: %s\n", err)
		return "", err
	}

	ext := filepath.Ext(_fileName)
	if ext == "" {
		ext = "." + t
		_fileName = _fileName + ext

	}
	name := strings.TrimSuffix(_fileName, ext)

	file, err := os.Create(storePath + "/" + _fileName)
	if err != nil {
		fmt.Printf("Error create image file: %s\n", err)
		return "", err
	}
	defer file.Close()
	if _, err = io.Copy(file, &buf); err != nil {
		fmt.Printf("Error copy image file: %s\n", err)
		return "", err
	}

	tb := resize.Thumbnail(100, 100, img, resize.Lanczos3)

	fileTb, err := os.Create(thumbPath + "/" + name + thumbSuffix + ext)
	if err != nil {
		fmt.Printf("Error create thumbnail file: %s\n", err)
		return "", err
	}
	defer fileTb.Close()
	switch t {
	default:
		png.Encode(fileTb, tb)
	case "jpeg":
		jpeg.Encode(fileTb, tb, nil)
	case "png":
		png.Encode(fileTb, tb)
	case "gif":
		gif.Encode(fileTb, tb, nil)
	}
	return _fileName, nil

}

func createDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func initEnv() {
	createDir(storePath)
	createDir(thumbPath)
}
