package filehost

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
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"

	log "github.com/sirupsen/logrus"
)

//BaseURL API base url of k2s
const BaseURL = "https://keep2share.cc/api/v2/"

//LoginValue Login to k2s
type LoginValue struct {
	Username string
	Password string
}

//LoginResponse Response after Login to k2s
type LoginResponse struct {
	Status    string `json:"status"`
	Code      int    `json:"code"`
	AuthToken string `json:"auth_token"`
	Message   string `json:"message"`
}

// GetUploadFormData This method allows to recieve file uploading form.
type GetUploadFormData struct {
	Status     string `json:"status"`
	Code       int    `json:"code"`
	FormAction string `json:"form_action"`
	FileField  string `json:"file_field"`
	Message    string `json:"message"`
	FormData   struct {
		Ajax      bool   `json:"ajax"`
		Params    string `json:"params"`
		Signature string `json:"signature"`
	} `json:"form_data"`
}

// FileUploaded This is a response when file is uploaded
type FileUploaded struct {
	Status     string `json:"status"`
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	UserFileID string `json:"user_file_id"`
	Link       string `json:"link"`
	Message    string `json:"message"`
}

// Keep2ShareValue Pass value to functions
type Keep2ShareValue struct {
	Ristretto    *ristretto.Cache
	Token        string
	Path         string // path of file to upload
	LinkUploaded string
	Login        LoginValue
}

//login login to k2s and save token into cache
func (k *Keep2ShareValue) login() error {

	username := k.Login.Username
	password := k.Login.Password

	values := map[string]string{"username": username, "password": password}
	jsonValue, err := json.Marshal(values)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "login",
		}).Error("Error in json.Marshal")
		return fmt.Errorf("Uploader[K2S] - login to K2C: %s", err.Error())
	}

	req, err := http.NewRequest("POST", BaseURL+"login", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "login",
		}).Error("Uploader[K2S] - Error in client.Do(req)")
		return fmt.Errorf("Uploader[K2S] - login to K2C: %s", err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var responseObject LoginResponse

	err = json.Unmarshal(body, &responseObject)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Login",
		}).Error("Uploader[K2S] - json.Unmarshal")
		return fmt.Errorf("Uploader[K2S] - Login: %s", err.Error())
	}

	if responseObject.Code != 200 {
		log.WithFields(log.Fields{
			"err":           responseObject.Message,
			"response Code": responseObject.Code,
			"status":        responseObject.Status,
			"handler":       "Login",
		}).Error("Uploader[K2S] -  Response")
		return fmt.Errorf("Uploader[K2S] - Login: %s", responseObject.Message)
	}

	log.Debugf("Token new post %s", responseObject.AuthToken)

	k.setToken(responseObject.AuthToken) //set token to cache
	return nil

}

//request make request
// return body
func request(action string, payload map[string]string) ([]byte, error) {

	jsonValue, err := json.Marshal(payload)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "request",
			"action":  action,
		}).Error("Uploader[K2S] -  Error in json.Marshal")
		return nil, fmt.Errorf("Uploader[K2S] - request %s: %s", action, err.Error())
	}

	req, err := http.NewRequest("POST", BaseURL+action, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "request",
			"action":  action,
		}).Error("Uploader[K2S] -  Error in client.Do(req)")
		return nil, fmt.Errorf("Uploader[K2S] - request %s: %s", action, err.Error())
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return body, nil
}

// testK2S This method allows to check auth-token lifetime.
// return false if token is expired
// return true if token is alive
func (k *Keep2ShareValue) testK2S() bool {

	values := map[string]string{"auth_token": k.Token}

	body, err := request("test", values)

	if err != nil {
		return false
	}

	var response struct {
		Status  string
		Code    int
		Message string
	}

	if err := json.Unmarshal(body, &response); err != nil {
		if err != nil {
			log.WithFields(log.Fields{
				"err":     err.Error(),
				"handler": "Test",
			}).Error("json.Unmarshal")
			return false
		}
	}

	if response.Code != 200 {
		log.WithFields(log.Fields{
			"err":           response.Message,
			"response Code": response.Code,
			"status":        response.Status,
			"handler":       "Test",
		}).Error("Uploader[K2S] - Response")
		return false
	}

	return true

}

// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func streamingUploadFile(params map[string]string, paramName, path string, w *io.PipeWriter, file *os.File) {
	defer file.Close()
	defer w.Close()
	writer := multipart.NewWriter(w)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		log.Errorf("Uploader[K2S] - Create Form error: %s", err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Errorf("Uploader[K2S] - Copy error: %s", err)
		return
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		log.Errorf("Uploader[K2S] - Close error: %s", err)
		return
	}
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()
	go streamingUploadFile(params, paramName, path, w, file)
	return http.NewRequest("POST", uri, r)
}

func uploadMultipartFile(client *http.Client, params map[string]string, uri, key, path string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("Uploader[K2S] - Request error: %s", err)
	}

	mwriter := multipart.NewWriter(writer)

	fmt.Println("1")

	/*for key, val := range params {
		mwriter.WriteField(key, val)
	}*/

	fmt.Println("2")
	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		w, err := mwriter.CreateFormFile(key, path)
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}
		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("Uploader[K2S] - error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mwriter.Close(); err != nil {
			errchan <- err
			return
		}
	}()

	//mwriter.WriteField("ajax", "strconv.FormatBool(responseObject.FormData.Ajax")
	//mwriter.WriteField("params", "responseObject.FormData.Params")
	//mwriter.WriteField("signature", "responseObject.FormData.Signature")
	req.Header.Add("Content-Type", mwriter.FormDataContentType())
	resp, err := client.Do(req)
	merr := <-errchan

	if err != nil || merr != nil {
		return resp, fmt.Errorf("Uploader[K2S] - http error: %v, multipart error: %v", err, merr)
	}

	return resp, nil
}

// Upload make upload to k2s server
// return link uploaded
func (k *Keep2ShareValue) upload() (link string, err error) {

	/*
	* First request uploading form
	*
	 */

	//get Form For Upload
	values := map[string]string{"auth_token": os.Getenv("k2stoken")}

	body, err := request("getUploadFormData", values)

	if err != nil {
		return link, err
	}

	log.Debugln(string(body))

	var responseObject GetUploadFormData

	err = json.Unmarshal(body, &responseObject)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "upload",
		}).Error("Upload[K2S] - json.Unmarshal")
		return link, fmt.Errorf("Uploader[K2S] - Login: %s", err.Error())
	}

	if responseObject.Code != 200 {
		log.WithFields(log.Fields{
			//"err":           err.Error(),
			"response Code": responseObject.Code,
			"status":        responseObject.Status,
			"response":      responseObject,
			"handler":       "upload",
			"message":       responseObject.Message,
		}).Error("Upload[K2S] - Response")
		return link, fmt.Errorf("Uploader[K2S] - Upload: %s", responseObject.Status)
	}

	/*
	* Make request to "form_action" received from first request
	*
	 */

	/*extraParams := map[string]string{
		"ajax":      strconv.FormatBool(responseObject.FormData.Ajax),
		"params":    responseObject.FormData.Params,
		"signature": responseObject.FormData.Signature,
	}*/

	// Open the file.
	file, err := os.Open(k.Path)
	if err != nil {
		log.Errorln("Uploader[K2S] - Open file error : %s", err)
		return
	}

	// Schedule the file to be closed once
	// the function returns.
	defer file.Close()

	// https://play.golang.org/p/Tmt7v3fIQF
	// Create a synchronous in-memory pipe. This pipe will
	// allow us to use a Writer as a Reader.
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	//var mpWriter *multipart.Writer
	mpWriter := multipart.NewWriter(pipeWriter)
	//body2 := &bytes.Buffer{}

	// Create a goroutine to perform the write since
	// the Reader will block until the Writer is closed.
	go func() {

		defer pipeWriter.Close()

		// Create a Reader that writes to the hash what it reads
		// from the file.
		fileReader := io.TeeReader(file, pipeWriter)

		// Create the multipart writer to put everything together.

		// Add the keys we generated.
		mpWriter.WriteField("ajax", strconv.FormatBool(responseObject.FormData.Ajax))
		mpWriter.WriteField("params", responseObject.FormData.Params)
		mpWriter.WriteField("signature", responseObject.FormData.Signature)

		fileWriter, err := mpWriter.CreateFormFile(responseObject.FileField, filepath.Base(k.Path))

		// Write the contents of the file to the multipart form.
		_, err = io.Copy(fileWriter, fileReader)
		if err != nil {
			log.Errorln("Uploader[K2S] - Write error : %s", err)
			return
		}

		// Close the Writer which will cause the Reader to unblock.

		if err := mpWriter.Close(); err != nil {
			log.Errorln("Uploader[K2S] - Close error : %s", err)
			return
		}
	}()

	// Wait until the Writer is closed, then write the
	// Pipe to Stdout.
	//_, err = io.Copy(os.Stdout, pipeReader)

	_, err = io.Copy(os.Stdout, pipeReader)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "upload",
		}).Error("Uploader[K2S] - client.Do")
		return link, fmt.Errorf("Uploader[K2S] - reader: %s", err.Error())
	}

	//os.Exit(1)
	//client := &http.Client{}

	// Send post request
	res, err := http.Post(responseObject.FormAction, mpWriter.FormDataContentType(), pipeReader)
	/*req, err := http.NewRequest("POST", responseObject.FormAction, pipeReader)
	//req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Set("Content-Type", mpWriter.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "upload",
		}).Error("Upload[K2S] - client.Do")
		return link, fmt.Errorf("upload: %s", err.Error())
	}
	*/
	fmt.Println(res)

	response, _ := ioutil.ReadAll(res.Body)

	// Check the response
	if res.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(response),
			"handler":  "upload",
		}).Error("Uploader[K2S] - response")
		return link, fmt.Errorf("Uploader[K2S] - Upload: %d", res.StatusCode)
	}

	var uploadedFile FileUploaded

	err = json.Unmarshal(response, &uploadedFile)

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "Upload",
		}).Error("Uploader[K2S] - json.Unmarshal")
		return link, fmt.Errorf("Uploader[K2S] - Upload: %s", err.Error())
	}

	if uploadedFile.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status":   res.StatusCode,
			"response": string(response),
			"handler":  "postImage",
		}).Error("Uploader[K2S] - response")
		return link, fmt.Errorf("Uploader[K2S] - postImage: %d", res.StatusCode)
	}

	link = fmt.Sprintf("%s/%s", uploadedFile.Link, extractNameFile(k.Path))

	// Replace http from response with https
	//link = strings.Replace(link, "http", "https", -1)

	k.LinkUploaded = link

	return link, nil
}

// setToken set token to cache
// https://github.com/dgraph-io/ristretto
func (k *Keep2ShareValue) setToken(t string) {

	k.Token = t
	log.Debugln("Token to save in cache", k.Token)

	k.Ristretto.Set("token", t, 1)
	os.Setenv("k2stoken", t)

}

// getToken get token into cache and return
// true if present
// false if not present
func (k *Keep2ShareValue) getToken() bool {

	if os.Getenv("k2stoken") != "" {
		return true
	}

	value, found := k.Ristretto.Get("token")
	if !found {
		log.Debugln("Missing Token")
		return false
	}

	log.Debugln(value)
	return true

}

// setUp start struct
func setUp(login LoginValue) (*Keep2ShareValue, error) {
	a, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return &Keep2ShareValue{}, err
	}

	return &Keep2ShareValue{
		Ristretto: a,
		Login:     login,
	}, nil
}

// K2S main function to upload to K2S
func K2S(path, username, password string) (l string, err error) {

	log.Infof("Uploader[K2S] - Starting uploading %s with K2S", path)

	var k *Keep2ShareValue

	login := LoginValue{
		Username: username,
		Password: password,
	}

	k, err = setUp(login)

	k.Path = path

	if err != nil {
		return l, err
	}

	// Check if token is present into cache
	ok := k.getToken()

	// If not present, make login and save token into cache
	if !ok {
		log.Debugln("Uploader[K2S] - Token not present, request new one")
		err = k.login()

		if err != nil {
			return l, err
		}
	}

	log.Debugln("Uploader[K2S] - Token from method", k.Token)

	// wait for value to pass through buffers
	time.Sleep(10 * time.Millisecond)

	test := k.testK2S()

	if !test {
		log.Debugln("Uploader[K2S] - Token expired, request new one")
		err = k.login()

		if err != nil {
			return l, err
		}
	}

	ok = k.getToken()

	if !ok {
		log.Debugln("Uploader[K2S] - Token not present again")
	}

	if ok {
		log.Debugln("Uploader[K2S] - Token present")
	}

	l, err = k.upload()

	if err != nil {
		return l, err
	}

	log.Debugf("Uploader[K2S] - File Uploaded: %s", l)

	return l, nil
}
