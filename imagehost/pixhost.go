package imagehost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// PixHostUpload json response after upload
type PixHostUpload struct {
	Name    string `json:"name"`
	ShowURL string `json:"show_url"`
	ThURL   string `json:"th_url"`
}

// PixHost ...
// API Doc: https://pixhost.to/api/index.html
func PixHost(f string) (fp *UploadedImageLink, err error) {

	fp = &UploadedImageLink{
		Direct: "",
		Thumb:  "",
	}

	urlUpload := "https://api.pixhost.to/images"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(f)

	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1,
			"handler": "Error coping file",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error coping file")
		return fp, err
	}

	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("img", filepath.Base(f))
	_, errFile1 = io.Copy(part1, file)

	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1,
			"handler": "Error creating form",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error creating form")
		return fp, err
	}

	_ = writer.WriteField("content_type", "1")
	err = writer.Close()

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error closing file",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error closing file")
		return fp, err
	}

	var client *http.Client
	if os.Getenv("proxy") != "" {
		proxyURL, err := url.Parse(os.Getenv("proxy"))
		if err != nil {
			log.WithFields(log.Fields{
				"event": "FastPic Close File",
				"File":  f,
				"err":   err,
			}).Error("Upload Image")
			return fp, err
		}

		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest(method, urlUpload, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Request",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error Request")
		return fp, err
	}

	req.Header.Add("Accept", "application/json")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Make Request",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error Make Request")
		return fp, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Read Body",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Error Read Body")
		return fp, err
	}

	log.Debugln("ImageHostUploader[PixHost] - Upload Response", string(body))

	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status code": res.StatusCode,
			"handler":     "Response error",
			"url":         "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Upload")
		return fp, fmt.Errorf("ImageHostUploader[PixHost] - Error Response Code: %d for https://api.pixhost.to/images", res.StatusCode)
	}

	var image PixHostUpload

	err = json.Unmarshal(body, &image)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Reading body to json",
			"url":     "https://api.pixhost.to/images",
		}).Errorf("Uploader[PixHost] - Reading json")
		return fp, err
	}

	log.WithFields(log.Fields{
		"Name":    image.Name,
		"ShowURl": image.ShowURL,
		"Thumb":   image.ThURL,
		"url":     "https://api.pixhost.to/images",
	}).Debugf("Uploader[PixHost] - Response json")

	// Load SHOWURL's url to get direct image
	// AND
	// Extract direct link

	doc, err := goquery.NewDocument(image.ShowURL)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Query",
			"url":     image.ShowURL,
		}).Errorf("Uploader[PimpAndHost] - Error reading homepage")
		return fp, err
	}

	// Get  <img id="image" data-zoom="out" class="image-img" src="https://img43.pixhost.to/images/286/222.jpg" alt="222.jpg"/>
	link, found := doc.Find("#image").Attr("src")

	if !found {

		log.WithFields(log.Fields{
			"handler": "Find Link",
			"url":     image.ShowURL,
		}).Errorf("Uploader[PimpAndHost] - Image's link not found")

		return fp, fmt.Errorf("Image's link not found")

	}

	fp = &UploadedImageLink{
		Direct: link,
		Thumb:  image.ThURL,
	}

	return fp, err
}
