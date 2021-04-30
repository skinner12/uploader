package imagehost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/skinner12/uploader/internal/manipulate"
)

// PostImagesUpload is the json response after upload image
type PostImagesUpload struct {
	Status string `json:"status"`
	URL    string `json:"url"`
	Error  string `json:"error"`
}

// PostImages upload image to postimages.org
func PostImages(f string) (fp *UploadedImageLink, err error) {

	fp = &UploadedImageLink{
		Direct: "",
		Thumb:  "",
	}

	// Load HomePage to get session value
	url := "https://postimages.org"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org",
		}).Errorf("Uploader[PostImages] - Error reading homepage make request")
		return fp, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org",
		}).Errorf("Uploader[PostImages] - Error reading homepage from Client")
		return fp, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org",
		}).Errorf("Uploader[PostImages] - Error reading homepage")
		return fp, err
	}

	var reToken = regexp.MustCompile(`(?m)\("token"\,"(\w+)"\)`)
	matchToken := reToken.FindStringSubmatch(string(body))

	if matchToken == nil {
		log.Errorln("Uploader[PostImages] - Error get session")
		return fp, err
	}

	log.Debugf("Uploader[PostImages] - Token: %s", matchToken[1])

	var charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	uploadSession := manipulate.StringWithCharset(32, charset) // Make 32 chars random

	log.Debugf("Uploader[PostImages] - Upload Session 32 char: %s", uploadSession)

	// Start Uploading
	url = "https://postimages.org/json/rr"
	method = "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(f)
	defer file.Close()

	part1,
		errFile1 := writer.CreateFormFile("file", filepath.Base(f))
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "GET",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Error Copy File")
		return
	}
	_ = writer.WriteField("token", matchToken[1])
	_ = writer.WriteField("upload_session", uploadSession)
	_ = writer.WriteField("numfiles", "1")
	_ = writer.WriteField("expire", "0")
	err = writer.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Error writer close")
		return
	}

	reqUpload, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Error make request")
		return
	}

	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	reqUpload.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")

	resUpload, err := client.Do(reqUpload)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Error doing request")
		return
	}
	defer resUpload.Body.Close()

	bodyUpload, err := ioutil.ReadAll(resUpload.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "GET",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Error reading request")
		return
	}

	log.Debugf("Uploader[PostImages] - Uplaod result: %s", string(bodyUpload))

	// Read Json Response
	var image PostImagesUpload

	err = json.Unmarshal(bodyUpload, &image)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Reading body to json",
			"url":     "https://postimages.org/json/rr",
		}).Errorf("Uploader[PostImages] - Reading json")
		return fp, err
	}

	log.WithFields(log.Fields{
		"Status": image.Status,
		"URL":    image.URL,
		"url":    "https://postimages.org/json/rr",
	}).Debugf("Uploader[PostImages] - Response json")

	// Check if STATUS is OK

	if image.Status != "OK" {
		return fp, fmt.Errorf("Uploader[PostImages] - Get wrong status response: %s with error %s", image.Status, image.Error)
	}

	// Load URL's url to get direct image
	// AND
	// Extract direct link

	doc, err := goquery.NewDocument(image.URL)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Query",
			"url":     image.URL,
		}).Errorf("Uploader[PostImages] - Error reading Upload Result Page")
		return fp, err
	}

	// Get  <input id="code_direct" type="text" value="https://i.postimg.cc/CK27FmCY/download-botton.png" autocomplete="off" readonly="">
	// Direct Link
	link, found := doc.Find("#code_direct").Attr("value")

	if !found {

		log.WithFields(log.Fields{
			"handler": "Find Link",
			"url":     image.URL,
		}).Errorf("Uploader[PostImages] - Image's link not found")

		return fp, fmt.Errorf("Uploader[PostImages] - Image's link not found")

	}

	// Get <input id="code_web_thumb" type="text" value="<a href='https://postimg.cc/Ffj3pbnn' target='_blank'>
	// <img src='https://i.postimg.cc/Ffj3pbnn/download-botton.png' border='0' alt='download-botton'/></a>" autocomplete="off" readonly="">
	// Thumb Link

	thumb, found := doc.Find("#code_web_thumb").Attr("value")

	if !found {

		log.WithFields(log.Fields{
			"handler": "Find Link",
			"url":     image.URL,
		}).Errorf("Uploader[PostImages] - Image's link not found")

		return fp, fmt.Errorf("Uploader[PostImages] - Image's link not found")

	}

	log.Debugf("Uploader[PostImages] - Thumb RAW code: %s", thumb)

	var reImg = regexp.MustCompile(`(?m)src\s*=\s*'(.+?)'`)
	matchImg := reImg.FindStringSubmatch(thumb)

	if matchImg == nil {
		log.Errorln("Uploader[PostImages] - Error get thumb Image")
		return fp, err
	}

	log.Debugf("Uploader[PostImages] - Thumb: %s", matchImg[1])

	fp = &UploadedImageLink{
		Direct: link,
		Thumb:  matchImg[1],
	}

	// Final return
	return fp, nil

}
