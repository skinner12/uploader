package imagehost

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

// FastpicResponse is the XML response of upload
type FastpicResponse struct {
	XMLName     xml.Name `xml:"UploadSettings"`
	Text        string   `xml:",chardata"`
	Xsi         string   `xml:"xsi,attr"`
	Xsd         string   `xml:"xsd,attr"`
	Imagepath   string   `xml:"imagepath"`
	Imageid     string   `xml:"imageid"`
	Session     string   `xml:"session"`
	Status      string   `xml:"status"`
	Error       string   `xml:"error"`
	Viewurl     string   `xml:"viewurl"`
	Viewfullurl string   `xml:"viewfullurl"`
	Thumbpath   string   `xml:"thumbpath"`
	Sessionurl  string   `xml:"sessionurl"`
}

// UploadedImageLink is the struct with image's links
type UploadedImageLink struct {
	Direct string
	Thumb  string
}

// FastPic Creates a new file upload http request with optional extra params
// https://github.com/Apkawa/uimge/tree/master/uimge/hosts
func FastPic(f string) (fp *UploadedImageLink, err error) {
	urlSite := "https://fastpic.ru/upload?api=1"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(f)

	if errFile1 != nil {
		log.WithFields(log.Fields{
			"event": "FastPic Open File",
			"File":  f,
			"err":   errFile1,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, errFile1)
	}

	defer file.Close()
	part1,
		errFile1 := writer.CreateFormFile("file1", filepath.Base(f))
	_, errFile1 = io.Copy(part1, file)

	if errFile1 != nil {
		log.WithFields(log.Fields{
			"event":          "FastPic Create Form File",
			"File":           f,
			"CreateFormFile": part1,
			"err":            errFile1,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, errFile1)
	}

	_ = writer.WriteField("method", "file")
	_ = writer.WriteField("check_thumb", "on")
	_ = writer.WriteField("uploading", "1")
	err = writer.Close()

	if err != nil {
		log.WithFields(log.Fields{
			"event": "FastPic Close File",
			"File":  f,
			"err":   err,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, err)
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

	req, err := http.NewRequest(method, urlSite, payload)

	if err != nil {
		log.WithFields(log.Fields{
			"event": "FastPic New Request",
			"File":  f,
			"err":   err,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, err)
	}

	req.Header.Add("User-Agent", "FPUploader")
	req.Header.Add("Content-Type", "multipart/form-data; boundary=--------------------------496798324961553600247059")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)

	if err != nil {
		log.WithFields(log.Fields{
			"event": "FastPic Make Client",
			"File":  f,
			"err":   err,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	//fmt.Println(string(body))

	log.Debugf("Response body: %s", string(body))

	response := &FastpicResponse{}

	err = xml.Unmarshal(body, &response)

	if err != nil {
		log.WithFields(log.Fields{
			"event": "FastPic Unmurshal",
			"File":  f,
			"Body":  string(body),
			"err":   err,
		}).Error("Upload Image")
		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, err)
	}

	if response.Error != "" {

		return fp, xerrors.Errorf("Uploader[ImageUpload] - FastPic.ru  %s: %s\n", f, response.Error)
	}

	if response.Error == "" {
		log.Debugln(response.Imageid)
		log.Debugln(response.Thumbpath)
		log.Debugln(response.Imagepath)
		log.Debugln(response.Viewfullurl)

	}

	log.WithFields(log.Fields{
		"event":       "FastPic",
		"File":        f,
		"ImageID":     response.Imageid,
		"Thumbpath":   response.Thumbpath,
		"Imagepath":   response.Imagepath,
		"Viewfullurl": response.Viewfullurl,
	}).Debugln("Upload Image")

	fp = &UploadedImageLink{
		//Direct: fmt.Sprintf("%s?noht=1&tk=6510", strings.Replace(response.Imagepath, "http", "https", 1)),
		Direct: fmt.Sprintf("%s", strings.Replace(response.Imagepath, "http", "https", 1)),
		Thumb:  response.Thumbpath,
	}

	return fp, nil

}
