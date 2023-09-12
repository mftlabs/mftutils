package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type SftpProfile struct {
	Profile    string     `json:"profile"`
	SftpConfig SftpConfig `json:"sftp_config"`
}
type SftpConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"username"`
	Passwd          string `json:"password"`
	RemotePath      string `json:"remote_path"`
	CleanupOnUpload bool   `json:"cleanup_on_upload"`
}

type AppConfig struct {
	Profiles []SftpProfile `json:"profiles"`
}

func LoadConfig(conf string, profile string) (appConf AppConfig, err error) {
	content, err := os.ReadFile(conf)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(content), &appConf)
	return
}

func main() {
	var conf string
	var profile string
	var fpath string
	var version bool
	flag.StringVar(&conf, "c", "./config.json", "config file path")
	flag.StringVar(&profile, "p", "default", "profile name")
	flag.StringVar(&fpath, "f", "", "file path")
	flag.BoolVar(&version, "v", false, "version")
	flag.Parse()

	if version {
		_, err := os.Stat("./version.txt")
		var f *os.File
		var newfile bool = false
		if err != nil {
			f, err = os.Create("./version.txt")
			newfile = true
		} else {
			f, err = os.OpenFile("./version.txt", os.O_RDWR, 0644)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1004)
		}
		defer f.Close()
		if newfile {
			curdate := time.Now().Format("20060102")
			curtime := time.Now().Format("150405")
			ver := "v" + curdate + curtime
			_, err = f.WriteString(ver)
			if err != nil {
				fmt.Println(err)
				os.Exit(1005)
			}
			fmt.Printf("%v", ver)
			f.Close()
		} else {
			buf := make([]byte, 100)
			_, err = f.Read(buf)
			if err != nil {
				fmt.Println(err)
				os.Exit(1006)
			}
			fmt.Printf("%v", string(buf))
			f.Close()
		}
		os.Exit(0)
	}

	if fpath == "" {
		flag.Usage()
		os.Exit(1000)
	}
	appConf, err := LoadConfig(conf, profile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1001)
	}
	err = appConf.UploadFiles(profile, fpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1002)
	}
	sc, err := appConf.GetSftpConfig(profile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1003)
	}
	fmt.Printf("Successfully uploaded file(s) %s to %s\n", fpath, sc.Host)
	os.Exit(0)
}
