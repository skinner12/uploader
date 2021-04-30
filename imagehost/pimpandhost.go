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
	"strconv"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// ResponsePost iis the response of 2ns requesto to get albumID
type ResponsePost struct {
	HTML    string `json:"html"`
	URL     string `json:"url"`
	AlbumID int    `json:"albumId"`
}

// FinalResponse is the final reponse with url of image uploaded
type FinalResponse struct {
	Files []struct {
		Name            string `json:"name"`
		Size            int    `json:"size"`
		Type            string `json:"type"`
		IsTotalUploaded bool   `json:"is_total_uploaded"`
		OriginalName    string `json:"original_name"`
		DynamicPath     string `json:"dynamicPath"`
		FileID          string `json:"fileId"`
		Image           struct {
			Num0    string `json:"0"`
			Num180  string `json:"180"`
			Num250  string `json:"250"`
			Num500  string `json:"500"`
			Num750  string `json:"750"`
			ID      int    `json:"id"`
			PageURL string `json:"pageUrl"`
			EditURL string `json:"editUrl"`
			Title   string `json:"title"`
			FileURL string `json:"fileUrl"`
		} `json:"image"`
	} `json:"files"`
}

// PimpAndHost ...
func PimpAndHost(f string) (fp *UploadedImageLink, err error) {

	var csrf string

	fp = &UploadedImageLink{
		Direct: "",
		Thumb:  "",
	}

	doc, err := goquery.NewDocument("https://pimpandhost.com/")

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Query",
			"url":     "https://pimpandhost.com/",
		}).Errorf("Uploader[PimpAndHost] - Error reading homepage")
		return fp, err
	}

	// Get <meta name="csrf-token" content="FWKYYNU9aoA83f9ceOiG6H0XsQ8NcRyag6QUFwfsqVVUU-8GkxBf9BGklA8qm7bYE1DSVmg2adba3HdiP9ydZg==">
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); name == "csrf-token" {
			csrf, _ = s.Attr("content")
			log.Debugf("Uploader[PimpAndHost] - Csrf Token: %s\n", csrf)
		}
	})

	// Posting to obtain some values
	urlUpload := "http://pimpandhost.com/album/create-by-uploading"
	method := "POST"

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

	req, err := http.NewRequest(method, urlUpload, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "make request",
			"url":     "http://pimpandhost.com/album/create-by-uploading",
		}).Errorf("Uploader[PimpAndHost] - Error reading page")
		return fp, err
	}

	req.Header.Add("X-CSRF-Token", csrf)
	req.Header.Set("User-Agent", os.Getenv("User-Agent"))

	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Client Request",
			"url":     "http://pimpandhost.com/album/create-by-uploading",
		}).Errorf("Uploader[PimpAndHost] - Error reading homepage")
		return fp, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"handler": "Response Body",
			"url":     "http://pimpandhost.com/album/create-by-uploading",
		}).Errorf("Uploader[PimpAndHost] - Response Body")
		return fp, err
	}

	log.Debugln("ImageHostUploader[PimpAndHost] - First Request", string(body))

	var a ResponsePost

	err = json.Unmarshal(body, &a)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"handler": "Make json",
			"url":     "http://pimpandhost.com/album/create-by-uploading",
		}).Errorf("Uploader[PimpAndHost] - Reading json")
		return fp, err
	}

	// NOW Upload file
	urlUpload = "https://pimpandhost.com/image/upload-file"
	method = "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	file, errFile4 := os.Open(f)
	defer file.Close()

	if errFile4 != nil {
		log.WithFields(log.Fields{
			"err":     errFile4,
			"handler": "Error opening file",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error opening file")
		return fp, errFile4
	}

	_ = writer.WriteField("albumId", strconv.FormatInt(int64(a.AlbumID), 10))
	_ = writer.WriteField("fileId", "234242342423") // TODO: make rando number
	_ = writer.WriteField("filename", file.Name())

	part4,
		errFile4 := writer.CreateFormFile("files[]", filepath.Base(f))
	_, errFile4 = io.Copy(part4, file)

	if errFile4 != nil {
		log.WithFields(log.Fields{
			"err":     errFile4,
			"handler": "Error coping file",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error coping file")
		return fp, err
	}

	err = writer.Close()

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error closing file",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error closing file")
		return fp, err
	}

	//client = &http.Client{}
	req, err = http.NewRequest(method, urlUpload, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Request",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error Request")
		return fp, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", os.Getenv("User-Agent"))

	res, err = client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Doing Request",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error Reques Dot")
		return fp, err
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Error Read Body",
			"url":     urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Error Read Body")
		return fp, err
	}

	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status code": res.StatusCode,
			"handler":     "Response error",
			"url":         urlUpload,
		}).Errorf("Uploader[PixHost] - Upload")
		return fp, fmt.Errorf("ImageHostUploader[PimpAndHost] - Error Response Code: %d for %s", res.StatusCode, urlUpload)
	}

	log.Debugln("ImageHostUploader[PimpAndHost] - Second Request", string(body))

	var image *FinalResponse

	err = json.Unmarshal(body, &image)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Reading body to json",
			"url":     urlUpload,
		}).Errorf("ImageHostUploader[PimpAndHost] -  Reading json")
		return fp, err
	}
	if len(image.Files) == 0 {
		log.WithFields(log.Fields{
			"handler": "Response empty",
			"url":     urlUpload,
		}).Errorf("ImageHostUploader[PimpAndHost] - Files are empty, not uploaded")
		return fp, fmt.Errorf("ImageHostUploader[PimpAndHost] - Files are empty, not uploaded")
	}

	log.Debugln("Image response unmarshal: ", image)

	log.WithFields(log.Fields{
		"direct link":      image.Files[0].Image.Num0,    // Image full size
		"page image":       image.Files[0].Image.PageURL, // URL to edit image
		"edit_url":         image.Files[0].Image.EditURL,
		"title":            image.Files[0].Image.Title,
		"name":             image.Files[0].Name,
		"size":             image.Files[0].Size,
		"original_name":    image.Files[0].OriginalName,
		"dynamicPath":      image.Files[0].DynamicPath,
		"fileId":           image.Files[0].FileID,
		"extrasmall_thumb": image.Files[0].Image.Num180,
		"small_thumb":      image.Files[0].Image.Num250,
		"medium_thumb":     image.Files[0].Image.Num500,
		"large_thumb":      image.Files[0].Image.Num750,
		"url":              urlUpload,
	}).Debugln("Uploader[PimpAndHost] - Links ")

	// Check if file is total uploaded
	if !image.Files[0].IsTotalUploaded {
		log.WithFields(log.Fields{
			"uploaded": image.Files[0].IsTotalUploaded,
			"handler":  "Reading body to json",
			"url":      urlUpload,
		}).Errorf("Uploader[PimpAndHost] - Not Total Uploaded")
		return fp, err
	}

	fp = &UploadedImageLink{
		Direct: image.Files[0].Image.Num0,
		Thumb:  image.Files[0].Image.Num250,
	}

	return fp, nil

}
