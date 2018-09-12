package main

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetFileFromPath_goodImg(t *testing.T) {
	e := echo.New()
	fv := make(url.Values)
	fv.Set("path", "https://cdn.pixabay.com/photo/2015/10/06/19/28/landscape-975091_960_720.jpg")

	req := httptest.NewRequest(echo.POST, "/upload", strings.NewReader(fv.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusOK {
		t.Errorf("Error get file from path!\n Response: %d; %s\n", rec.Code, rec.Body)
		t.Fail()
	}

	results, err := getResults(rec, t)
	if err != nil {
		t.Errorf("Error parse json response %s\n", err)
		t.Fail()
	}
	if len(results) == 1 {
		if results[0].Error {
			t.Errorf("Error upload good image!")
			t.Fail()
		} else {
			for _, value := range results {
				clearTestData(value, t)
			}
		}
	}

}

func TestGetFileFromPath_badImg(t *testing.T) {
	e := echo.New()
	fv := make(url.Values)
	fv.Set("path", "https://github.com/mushinnomushin/ZeuS_2.0.8.9/blob/master/README.txt")

	req := httptest.NewRequest(echo.POST, "/upload", strings.NewReader(fv.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	chooseHandler(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Error get file from path!\n Response: %d; %s\n", rec.Code, rec.Body)
		t.Fail()
	} else {
		results, err := getResults(rec, t)
		if err != nil {
			t.Errorf("Error parse json response %s\n", err)
			t.Fail()
		}
		if len(results) == 1 && !results[0].Error {
			t.Errorf("Error recognize bad format of image!")
		}
	}

}
