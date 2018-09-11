package main

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/rs/xid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetFileFromPath_badPath(t *testing.T) {
	e := echo.New()
	fv := make(url.Values)
	fv.Set("path", "./test/testdata/test_pic3.jpg")

	req := httptest.NewRequest(echo.POST, "/upload", strings.NewReader(fv.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Error get file from path!\n Response: %d; %s\n", rec.Code, rec.Body)
	} else {
		results, err := getResults(rec, t)
		if err != nil {
			t.Errorf("Error parse json response %s\n", err)
			t.Fail()
		}
		if len(results) == 1 && !results[0].Error {
			t.Errorf("Found downloaded image by bad path!")
		}
	}

}

func getResults(rec *httptest.ResponseRecorder, t *testing.T) ([]*Result, error) {
	t.Helper()
	var results []*Result
	err := json.Unmarshal(rec.Body.Bytes(), &results)
	if err != nil {
		t.Errorf("Error binding: %s\n", err)
	}
	return results, err
}

func TestInitEnv(t *testing.T) {
	initEnv()
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Errorf("Directory for store pictures not found!")
	}
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Errorf("Directory for store thumbnails not found!")
	}
}

func TestMakePic(t *testing.T) {
	const path = "test/testdata/test_pic1."

	id := xid.New().String()

	ext := [...]string{"jpg", "png", "gif"}

	for _, sfx := range ext {
		filePic, err := os.Open(path + sfx)
		if err != nil {
			t.Errorf("Error open test file %s", path+sfx)
		}

		makePic(filePic, "test_pic1."+sfx, id)
		filePic.Close()
		if _, err := os.Stat(storePath + "/" + id + "_test_pic1." + sfx); os.IsNotExist(err) {
			t.Errorf("Image file not saved!")
		}

		if _, err := os.Stat(thumbPath + "/" + id + "_test_pic1" + thumbSuffix + "." + sfx); os.IsNotExist(err) {
			t.Errorf("Thumbnails file not saved!")
		}
		os.Remove(storePath + "/" + id + "_test_pic1." + sfx)
		os.Remove(thumbPath + "/" + id + "_test_pic1" + thumbSuffix + "." + sfx)
	}

}

func TestBase64Good(t *testing.T) {
	e := echo.New()
	data, err := ioutil.ReadFile("test/testdata/test.json")

	if err != nil {
		t.Errorf("Error loadtest data from JSON file: %s\n", err)
	}

	req := httptest.NewRequest(echo.POST, "/upload", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("Error send file in BASE64!\n Response: %d; %s\n", rec.Code, rec.Body)
	} else {
		results, err := getResults(rec, t)
		if err != nil {
			t.Errorf("Error parse json response %s\n", err)
			t.Fail()
		}
		if len(results) == 1 && !results[0].Error {
			for _, value := range results {
				clearTestData(value, t)
			}
		} else {
			t.Errorf("Error in result for good request!")
			t.Fail()
		}
	}
}

func TestBase64Bad(t *testing.T) {
	e := echo.New()
	data, err := ioutil.ReadFile("test/testdata/test_bad.json")

	if err != nil {
		t.Errorf("Error loadtest data from JSON file: %s\n", err)
	}

	req := httptest.NewRequest(echo.POST, "/upload", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Error send bad file in BASE64!\n Response: %d; %s\n", rec.Code, rec.Body)
	} else {
		results, err := getResults(rec, t)
		if err != nil {
			t.Errorf("Error parse json response %s\n", err)
			t.Fail()
		}
		if len(results) == 1 && !results[0].Error {
			t.Errorf("Error recognize bad format of base64!")
		}

	}
}

func TestMultiform(t *testing.T) {
	const path = "test/testdata/test_pic1.jpg"

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("file", path)
	if err != nil {
		t.Errorf("Error writing to buffer: %s", err)
		t.Fail()
	}

	fh, err := os.Open(path)
	if err != nil {
		t.Errorf("Error opening file: %s", err)
		t.Fail()
	}
	defer fh.Close()
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		t.Errorf("Error copy file: %s", err)
		t.Fail()
	}

	bodyWriter.Close()

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/upload", bodyBuf)
	req.Header.Set(echo.HeaderContentType, bodyWriter.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("Error senf file in multipart form")
		t.Fail()
	} else {
		results, err := getResults(rec, t)
		if err != nil {
			t.Errorf("Error parse json response %s\n", err)
			t.Fail()
		}
		if len(results) == 1 && !results[0].Error {
			for _, value := range results {
				clearTestData(value, t)
			}
		} else {
			t.Errorf("Error in result for good request!")
			t.Fail()
		}
	}
}

func clearTestData(value *Result, t *testing.T) {
	t.Helper()
	err := os.Remove(storePath + "/" + value.Filename)
	ext := filepath.Ext(value.Filename)
	name := strings.TrimSuffix(value.Filename, ext)
	errTmb := os.Remove(thumbPath + "/" + name + thumbSuffix + ext)
	if err != nil && errTmb != nil {
		t.Errorf("Error remove test images: %s\n %s", err, errTmb)
	}
}
