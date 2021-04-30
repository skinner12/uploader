package filehost

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
)

// GetServerStruct is the response to get uplaod server
type GetServerStruct struct {
	Msg        string `json:"msg"`
	ServerTime string `json:"server_time"`
	Status     int    `json:"status"`
	Result     string `json:"result"`
}

// LinkFileUploaded is the struct where is saved
// the different URL for html and embeded code
type LinkFileUploaded struct {
	HTMLCode    string
	EmbededCode string
}

// GetServer return the server where upload file
func getServer(apiKey string) (s string, err error) {
	client := &http.Client{}

	server := fmt.Sprintf("https://api.gounlimited.to/api/upload/server?key=%s", apiKey)

	req, _ := http.NewRequest("GET", server, nil)

	resp, err := client.Do(req)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "getServer",
			"request": resp,
		}).Errorf("Uploader[goUnlimited] - Error in client.Do")
		return s, err
	}

	defer resp.Body.Close()

	log.Debugf("StatusCode: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"err":     resp.StatusCode,
			"handler": "getServer",
		}).Error("Uploader[goUnlimited] - Error in resp.StatusCode")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var responseObject GetServerStruct

	err = json.Unmarshal(body, &responseObject)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "getServer",
			"body":    string(body),
		}).Error("Uploader[goUnlimited] - json.Unmarshal")
		return s, fmt.Errorf("getServer: %s", err.Error())
	}

	if responseObject.Status != 200 {
		log.WithFields(log.Fields{
			"err":     "Response is not OK",
			"handler": "getServer",
			"message": responseObject.Msg,
		}).Error("Uploader[goUnlimited] - json response")
		return s, fmt.Errorf("getServer: Response is not OK")
	}

	s = responseObject.Result
	log.Debugf("Uploader[goUnlimited] - URL to Upload: %s\n", responseObject.Result)

	return s, nil
}

// uploadFile upload a file to a server of gounlimited.com
// url string, is the url to upload to
// f string, path of file
func uploadFile(url, p, apiKey string) (l LinkFileUploaded, err error) {

	file, err := os.Open(p)
	if err != nil {
		log.Errorln("Uploader[goUnlimited] - Open File", err)
		return
	}

	// Schedule the file to be closed once
	// the function returns.
	defer file.Close()

	pipeReader, pipeWriter := io.Pipe()

	//var mpWriter *multipart.Writer
	mpWriter := multipart.NewWriter(pipeWriter)
	//body2 := &bytes.Buffer{}

	// Create a goroutine to perform the write since
	// the Reader will block until the Writer is closed.
	go func() {

		// Create a Reader that writes to the hash what it reads
		// from the file.
		//fileReader := io.TeeReader(file, pipeWriter)

		// Create the multipart writer to put everything together.

		fileWriter, err := mpWriter.CreateFormFile("file", filepath.Base(file.Name()))

		// Write the contents of the file to the multipart form.
		_, err = io.Copy(fileWriter, file)
		if err != nil {
			fmt.Println("Write File", err)
			return
		}

		// Add the keys we generated.
		mpWriter.WriteField("api_key", apiKey)

		// Close the Writer which will cause the Reader to unblock.
		defer mpWriter.Close()
		defer pipeWriter.Close()
	}()

	// Wait until the Writer is closed, then write the
	// Pipe to Stdout.
	_, err = io.Copy(os.Stdout, pipeReader)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "upload",
		}).Error("Uploader[goUnlimited] - client.Do")
		return l, fmt.Errorf("reader: %s", err.Error())
	}

	os.Exit(1)

	req, err := http.NewRequest("POST", url, pipeReader)
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())

	client := &http.Client{}
	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "uploadFile",
		}).Error("Uploader[goUnlimited] - client.Do")
		return l, fmt.Errorf("Uploader[goUnlimitd] - Client DO: %s", err.Error())
	}

	response, _ := ioutil.ReadAll(res.Body)

	log.Debugln("Uploader[goUnlimited] - Response Upload File: %s\n", string(response))

	// Check the response
	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(response),
			"handler":  "uploadFile",
		}).Error("Upload[goUnlimited] - response")
		return l, fmt.Errorf("Upload[goUnlimited] - upload File: %d", res.StatusCode)
	}

	idFile := ExtractID(string(response))

	if idFile == "" {
		log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(response),
			"handler":  "uploadFile",
		}).Error("Upload[goUnlimited] - response is empty")
		return l, fmt.Errorf("Upload[goUnlimited] - Uploaded file response: %s", string(response))
	}

	l = LinkFileUploaded{
		HTMLCode:    fmt.Sprintf("https://gounlimited.to/%s/%s", idFile, extractNameFile(p)),
		EmbededCode: fmt.Sprintf("https://gounlimited.to/embed-%s.html", idFile),
	}

	fmt.Println(l)

	return l, nil

}

func extractNameFile(f string) string {
	return filepath.Base(f)
}

//ExtractID extract ID from HTML response of uploaded file
// it takes the value of textarea name="fn"
func ExtractID(element string) string {
	re := regexp.MustCompile(`(<textarea\b[^>]*\bname\s*=\s*(?:\"|')\s*fn\s*(?:\"|')[^<]([A-Za-z0-9]+)<\/textarea>)`)

	submatchall := re.FindAllStringSubmatch(string(element), -1)
	if submatchall == nil {
		return ""
	}

	return submatchall[0][2]
}

// GOUnlimited upload a file to gounlimited.com
func GOUnlimited(path, apikey string) (s LinkFileUploaded, err error) {

	l, err := getServer(apikey)

	if err != nil {
		return s, err
	}

	f, err := uploadFile(l, path, apikey)

	if err != nil {
		return s, err
	}

	log.Infoln("Upload[goUnlimited] - File URL:", f)

	return f, nil

}
