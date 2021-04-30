package upload

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func isValidURL(toTest string) bool {
	re := regexp.MustCompile(`^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`)
	log.Debugf("Pattern: %v\n", re.String())
	if re.MatchString(toTest) {
		return true
	}
	return false

}

//uploadExeCMD exec bash script to upload file
//upload.sh -a 'username:password' file_host_name /path/to/file

func uploadExeCMD(f, path, filehost, username, password, minSpeed string) (l string, err error) {

	/**************************************************
	*
	* @f file
	* @filehost sitename
	* @path where the script is located path/to/app/
	* @username username
	* @password password
	*
	**************************************************/

	uploadPath := fmt.Sprintf("%supload.sh", path)
	log.Debugf("Uploader Path: %s", uploadPath)
	//Check if ffmpeg is installed
	path, err = exec.LookPath(uploadPath)
	if err != nil {
		log.Fatalf("Bash script for upload is required")
	}
	log.Debugf("Bash script for upload is available at %s\n", path)

	var login string
	if password == "" {
		login = username
	} else {
		login = fmt.Sprintf("%s:%s", username, password)
	}

	//cmd := exec.Command(parts[0], parts[1])
	cmd := exec.Command(uploadPath, "--min-rate", minSpeed, "-a", login, filehost, f)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()

	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"cmd":     cmd,
			"output":  stderr.String(),
			"handler": "UploaderExeCMD",
		}).Errorf("EXEC[Uploader] -Error in EXEC")
		return l, err
	}

	l = out.String()
	log.Debugf("Result: " + out.String())

	if ok := isValidURL(l); !ok {
		log.WithFields(log.Fields{
			"err":     out.String(),
			"cmd":     cmd,
			"output":  stderr.String(),
			"handler": "UploaderExeCMD",
		}).Errorf("EXEC[Uploader] - Not a valid URL in response")
		return l, err
	}
	return l, nil
}

// ToFileHost The main function to call
/**************************************************
*
* @f file to upload, full path
* @filehost name of filehost to upload
* @path where the script is located path/to/app/
* @username username
* @password password
* @minSpeed min speed of upload in kb/s EX: 1500
*
**************************************************/
func ToFileHost(f, path, filehost, username, password, minSpeed string) (l string, err error) {

	l, err = uploadExeCMD(f, path, filehost, username, password, minSpeed)
	if err != nil {
		return l, err
	}

	return l, nil
}
