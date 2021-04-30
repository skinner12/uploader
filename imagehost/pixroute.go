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
	"strings"

	"github.com/PuerkitoBio/goquery"

	log "github.com/sirupsen/logrus"
)

// PixRouteUpload is the json response after upload image
type PixRouteUpload struct {
	FileStatus string `json:"file_status"`
	FileCode   string `json:"file_code"`
	Error      string `json:"error"`
}

// PixRoute upload image to pixroute.com
func PixRoute(f string) (fp *UploadedImageLink, err error) {
	fp = &UploadedImageLink{
		Direct: "",
		Thumb:  "",
	}

	// Load HomePage to get URL value
	url := "https://pixroute.com/"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://pixroute.com/",
		}).Errorf("Uploader[PixRoute] - Error reading homepage make request")
		return fp, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "GET",
			"url":     "https://pixroute.com/",
		}).Errorf("Uploader[PixRoute] - Error reading homepage from Client")
		return fp, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Query",
			"url":     "https://pixroute.com/",
		}).Errorf("Uploader[PostImages] - Error reading From Response")
		return fp, err
	}

	// Get  <form id="uploadfile" action="https://img150.pixroute.com/cgi-bin/upload.cgi?upload_type=file&utype=anon">
	// URL for UPLOAD
	url, found := doc.Find("#uploadfile").Attr("action")

	if !found {
		log.WithFields(log.Fields{
			"handler": "Find Link",
			"url":     "https://pixroute.com/",
		}).Errorf("Uploader[PostImages] - Image's link not found")

		return fp, fmt.Errorf("Uploader[PostImages] - Image's link not found")
	}

	log.Debugln("URL to Upload to:", url)

	// Start Upload
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
			"handler": "POST",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error creating form")
		return fp, err
	}

	_ = writer.WriteField("utype1", "anon")
	_ = writer.WriteField("file_public1", "1")
	_ = writer.WriteField("thumb_size1", "190x190")
	_ = writer.WriteField("per_row1", "1")
	_ = writer.WriteField("to_folder1", "0")
	_ = writer.WriteField("upload1", "Start Upload")
	_ = writer.WriteField("keepalive1", "1")
	err = writer.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error write fields")
		return fp, err
	}

	reqUpload, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error make request")
		return fp, err
	}
	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	reqUpload.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")

	resUpload, err := client.Do(reqUpload)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error do request")
		return fp, err

	}
	defer resUpload.Body.Close()

	bodyUpload, err := ioutil.ReadAll(resUpload.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error read response")
		return fp, err
	}
	log.Debugln("Uploader[PixRoute] -", string(bodyUpload))

	// Read Json Response
	var image []PixRouteUpload

	err = json.Unmarshal(bodyUpload, &image)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"handler":  "Reading body to json",
			"response": string(bodyUpload),
			"url":      url,
		}).Errorf("Uploader[PixRoute] - Reading json")
		return fp, err
	}

	if len(image) == 0 {
		log.WithFields(log.Fields{
			"url":      url,
			"response": string(bodyUpload),
		}).Errorf("Uploader[PixRoute] - Response json empty")
		return fp, fmt.Errorf("Uploader[PixRoute] - Json missing")
	}

	log.WithFields(log.Fields{
		"Status": image[0].FileStatus,
		"Code":   image[0].FileCode,
		"url":    url,
	}).Debugf("Uploader[PixRoute] - Response json")

	// Check if STATUS is OK

	if image[0].FileStatus != "OK" {
		return fp, fmt.Errorf("Uploader[PixRoute] - Get wrong status response: %s with error %s", image[0].FileStatus, image[0].Error)
	}

	// Now read the result page for extract links

	url = fmt.Sprintf("https://pixroute.com/?op=upload_result&st=OK&fn=%s&per_row=1", image[0].FileCode)
	method = "GET"

	reqFinal, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
			"url": url,
		}).Errorf("Uploader[PixRoute] - Make request")
		return fp, err
	}

	reqFinal.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")

	resFinal, err := client.Do(reqFinal)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"response": string(bodyUpload),
			"url":      url,
		}).Errorf("Uploader[PixRoute] - Client make do")
		return fp, err
	}
	defer resFinal.Body.Close()

	// Grab Links
	docFinal, err := goquery.NewDocumentFromResponse(resFinal)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Query",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error reading From Response")
		return fp, err
	}

	// Get  <form id="uploadfile" action="https://img150.pixroute.com/cgi-bin/upload.cgi?upload_type=file&utype=anon">
	// URL for UPLOAD
	docFinal.Find(".sharetabs .box").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s)
		band, _ := s.Find("a").Attr("href")
		title, _ := s.Find("img").Attr("src")
		text := s.Find("textarea").Text()
		fmt.Printf("Review %d: %s - %s\n", i, band, title)

		var reImg = regexp.MustCompile(`(?m)src\s*=\s*"(.+?)"`)
		matchImg := reImg.FindStringSubmatch(text)

		if len(matchImg) > 0 {
			log.Debugf("Uploader[PixRoute] - Thumb: %s", matchImg[1])
			fp = &UploadedImageLink{
				Direct: strings.Replace(matchImg[1], "_t", "", 1),
				Thumb:  matchImg[1],
			}
		}

	})

	if fp.Direct == "" {
		log.WithFields(log.Fields{
			"handler": "Query",
			"url":     url,
		}).Errorf("Uploader[PixRoute] - Error get links image")
		return fp, fmt.Errorf("Uploader[PixRoute] - Links empty")
	}

	// Final return
	return fp, nil
}
