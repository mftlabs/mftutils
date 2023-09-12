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
		curdate := time.Now().Format("20060102")
		curtime := time.Now().Format("150405")
		fmt.Println(curdate + "-" + curtime)
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
