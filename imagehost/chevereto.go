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

	log "github.com/sirupsen/logrus"
)

// CheveretoStruct is the response after upload a image file
type CheveretoStruct struct {
	StatusCode int `json:"status_code"`
	Success    struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"success"`
	Image struct {
		Name              string      `json:"name"`
		Extension         string      `json:"extension"`
		Size              string      `json:"size"`
		Width             string      `json:"width"`
		Height            string      `json:"height"`
		Date              string      `json:"date"`
		DateGmt           string      `json:"date_gmt"`
		Title             string      `json:"title"`
		Description       interface{} `json:"description"`
		Nsfw              string      `json:"nsfw"`
		StorageMode       string      `json:"storage_mode"`
		Md5               string      `json:"md5"`
		SourceMd5         interface{} `json:"source_md5"`
		OriginalFilename  string      `json:"original_filename"`
		OriginalExifdata  string      `json:"original_exifdata"`
		Views             string      `json:"views"`
		CategoryID        interface{} `json:"category_id"`
		Chain             string      `json:"chain"`
		ThumbSize         string      `json:"thumb_size"`
		MediumSize        string      `json:"medium_size"`
		ExpirationDateGmt interface{} `json:"expiration_date_gmt"`
		Likes             string      `json:"likes"`
		IsAnimated        string      `json:"is_animated"`
		IsApproved        string      `json:"is_approved"`
		File              struct {
			Resource struct {
				Type string `json:"type"`
			} `json:"resource"`
		} `json:"file"`
		IDEncoded string `json:"id_encoded"`
		Filename  string `json:"filename"`
		Mime      string `json:"mime"`
		URL       string `json:"url"`
		URLViewer string `json:"url_viewer"`
		URLShort  string `json:"url_short"`
		Image     struct {
			Filename  string `json:"filename"`
			Name      string `json:"name"`
			Mime      string `json:"mime"`
			Extension string `json:"extension"`
			URL       string `json:"url"`
			Size      string `json:"size"`
		} `json:"image"`
		Thumb struct {
			Filename  string `json:"filename"`
			Name      string `json:"name"`
			Mime      string `json:"mime"`
			Extension string `json:"extension"`
			URL       string `json:"url"`
			Size      string `json:"size"`
		} `json:"thumb"`
		SizeFormatted      string `json:"size_formatted"`
		DisplayURL         string `json:"display_url"`
		DisplayWidth       string `json:"display_width"`
		DisplayHeight      string `json:"display_height"`
		ViewsLabel         string `json:"views_label"`
		LikesLabel         string `json:"likes_label"`
		HowLongAgo         string `json:"how_long_ago"`
		DateFixedPeer      string `json:"date_fixed_peer"`
		TitleTruncated     string `json:"title_truncated"`
		TitleTruncatedHTML string `json:"title_truncated_html"`
		IsUseLoader        bool   `json:"is_use_loader"`
	} `json:"image"`
	StatusTxt string `json:"status_txt"`
	Error     struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
		Context string `json:"context"`
	} `json:"error"`
}

// Chevereto is the main function for upload image to an image host based of chevereto.com
func Chevereto(url, apiKey, f string) (fp *UploadedImageLink, err error) {

	fp = &UploadedImageLink{
		Direct: "",
		Thumb:  "",
	}

	url_request := fmt.Sprintf("%s/api/1/upload/?key=%s&format=json", url, apiKey)
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(f)
	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("UploaderUploader[Chevereto] - Error open file")
		return fp, errFile1
	}
	part1,
		errFile1 := writer.CreateFormFile("source", filepath.Base(f))
	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("UploaderUploader[Chevereto] - Error CreateFormFile")
		return fp, errFile1
	}
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Copy File")
		return fp, errFile1
	}
	defer file.Close()
	err = writer.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err":     errFile1.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Closing File")
		return fp, errFile1
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url_request, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Make new Request")
		return fp, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Do new Request")
		return fp, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "POST",
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Reading Response Request")
		return fp, err
	}

	log.Debugln(string(body))

	response := &CheveretoStruct{}

	err = json.Unmarshal(body, &response)

	if err != nil {
		log.WithFields(log.Fields{
			"err":  err.Error(),
			"url":  url,
			"file": f,
		}).Errorf("Uploader[Chevereto] - Error Unmarshal Response")
		return fp, err
	}

	if response.StatusTxt != "OK" {

		log.WithFields(log.Fields{
			"Status":  response.StatusTxt,
			"Error":   response.Error.Message,
			"code":    response.Error.Code,
			"Context": response.Error.Context,
			"url":     url,
			"file":    f,
		}).Errorf("Uploader[Chevereto] - Error Upload Image")

		return fp, fmt.Errorf("Uploader[Chevereto] - Error Upload Image; %s", response.Error.Message)
	}

	fp.Direct = response.Image.URL
	fp.Thumb = response.Image.Thumb.URL

	return fp, err
}
