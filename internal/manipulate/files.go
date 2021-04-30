package manipulate

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DeleteFile delete file from disk
func DeleteFile(path string) error {
	// delete file
	var err = os.Remove(path)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "DeleteFile",
		}).Errorf("Error in DeleteFile")
		return err
	}
	return nil
}

func hashFileMD5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "hashFileMD5",
			"file":    filePath,
		}).Error("os.Open")
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "hashFileMD5",
			"file":    filePath,
		}).Error("io.Copy")
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

// ChangeMD5 change md5 hash of file
func ChangeMD5(path string) error {

	returnMD5String, err := hashFileMD5(path)

	if err != nil {
		return err
	}

	log.Debugln("MD5 BEFORE: ", returnMD5String)

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "ChangeMD5",
			"file":    path,
		}).Error("os.OpenFile")
		return err
	}

	defer file.Close()

	if _, err := file.WriteString("0"); err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"handler": "ChangeMD5",
			"file":    path,
		}).Error("file.WriteString")
		return err
	}

	returnMD5String2, err := hashFileMD5(path)

	if err != nil {
		return err
	}

	log.Debugln("MD5 AFTER: ", returnMD5String2)

	return nil

}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// RemoveSpaceFromFile from a folder, rename all files without spaces and
// strange chars
func RemoveSpaceFromFile(dir string) error {
	err := filepath.Walk(dir,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				log.WithFields(log.Fields{
					"Err":  err,
					"File": file,
				}).Errorln("Manipulate[File] - Reading Directory")
				return fmt.Errorf("Manipulate[File] - Reading Directory: %s", err.Error())
			}
			if !info.IsDir() {

				oldName := info.Name()
				oldNameWithoutExt := strings.TrimSuffix(oldName, filepath.Ext(oldName))
				oldNameExt := filepath.Ext(info.Name())
				log.Debugf("Old Name: %s", oldName)
				re, err := regexp.Compile(`[^\w]`)
				if err != nil {
					log.WithFields(log.Fields{
						"Err":  err,
						"File": file,
					}).Errorln("Manipulate[File] - Regex File")
				}

				newName := strings.Join(strings.Fields(re.ReplaceAllString(oldNameWithoutExt, " ")), "_")

				log.Debugf("New name: %s", newName)

				err = os.Rename(filepath.Join(filepath.Dir(file), oldName), filepath.Join(filepath.Dir(file), fmt.Sprintf("%s%s", newName, oldNameExt)))
				if err != nil {
					log.WithFields(log.Fields{
						"Err":  err,
						"File": file,
					}).Errorln("Manipulate[File] - Rename File")
					return fmt.Errorf("Manipulate[File] - Rename File: %s", err.Error())
				}

				return nil
			}
			return nil
		})

	if err != nil {
		return err
	}
	return nil
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
