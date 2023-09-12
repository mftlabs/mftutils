package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

func (ac *AppConfig) GetSftpConfig(profile string) (sftpConfig SftpConfig, err error) {
	for _, v := range ac.Profiles {
		if v.Profile == profile {
			sftpConfig = v.SftpConfig
			return
		}
	}
	err = fmt.Errorf("profile %s not found", profile)
	return
}

func (ac *AppConfig) GetSftpConfigByIndex(index int) (sftpConfig SftpConfig, err error) {
	if index < 0 || index >= len(ac.Profiles) {
		err = fmt.Errorf("index %d out of range", index)
		return
	}
	sftpConfig = ac.Profiles[index].SftpConfig
	return
}

func (ac *AppConfig) UploadFiles(profile string, fpath string) (err error) {
	sftpConfig, err := ac.GetSftpConfig(profile)
	if err != nil {
		return
	}
	return sftpConfig.UploadFiles(fpath)
}

func (sc *SftpConfig) UploadFile(fpath string) (err error) {
	fmt.Println("Uploading file", fpath, "to", sc.Host)
	client, err := sc.Connect()
	if err != nil {
		return
	}
	defer client.Close()
	// open source file
	srcFile, err := os.Open(fpath)
	if err != nil {
		return
	}
	defer srcFile.Close()
	// create destination file
	fileinfo, err := srcFile.Stat()
	if err != nil {
		return
	}
	fmt.Println("Uploading file", fileinfo.Name(), "to", sc.RemotePath)
	dstFile, err := client.Create(sc.RemotePath + "/" + fileinfo.Name())
	if err != nil {
		return
	}
	defer dstFile.Close()
	// copy source file to destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return
	}
	srcFile.Close()
	dstFile.Close()
	fmt.Println("Uploaded file", fileinfo.Name(), "to", sc.RemotePath, "successfully")
	if sc.CleanupOnUpload {
		fmt.Println("Deleting file", fpath)
		err = os.Remove(fpath)
		if err != nil {
			return
		}
	}
	return
}
func (sc *SftpConfig) UploadDir(fpath string) (err error) {
	fmt.Println("Uploading directory", fpath, "to", sc.Host)
	client, err := sc.Connect()
	if err != nil {
		return
	}
	defer client.Close()
	// open source file
	files := make([]string, 0)
	err = filepath.Walk(fpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return
	}
	for _, v := range files {
		var srcFile *os.File
		var stat os.FileInfo
		stat, err = os.Stat(v)
		if err != nil {
			return
		}
		if stat.IsDir() {
			continue
		}
		srcFile, err = os.Open(v)
		if err != nil {
			return
		}
		defer srcFile.Close()
		// create destination file
		var fileinfo os.FileInfo
		fileinfo, err = srcFile.Stat()
		if err != nil {
			return
		}
		//fmt.Println("Uploading directory", fileinfo.Name(), "to", sc.RemotePath)
		/*err = client.Mkdir(sc.RemotePath + "/" + fileinfo.Name())
		if err != nil {
			return
		}*/
		var dstFile *sftp.File
		dstFile, err = client.Create(sc.RemotePath + "/" + fileinfo.Name())
		if err != nil {
			return
		}
		defer dstFile.Close()
		// copy source file to destination file
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return
		}
		srcFile.Close()
		dstFile.Close()
		fmt.Println("Uploaded file", fileinfo.Name(), "to", sc.RemotePath, "successfully")
		if sc.CleanupOnUpload {
			fmt.Println("Deleting file", v)
			err = os.Remove(v)
			if err != nil {
				return
			}
		}

	}
	return
}
func (sc *SftpConfig) UploadFiles(fpath string) (err error) {
	stat, err := os.Stat(fpath)
	if err != nil {
		return
	}
	if stat.IsDir() {
		//fmt.Println("Uploading directory", fpath, "to", sc.Host)
		err = sc.UploadDir(fpath)
	} else {
		err = sc.UploadFile(fpath)
	}
	return
}

func (sc *SftpConfig) Dial() (*ssh.Client, error) {
	fmt.Println("Dialing to", sc.Host)
	config := &ssh.ClientConfig{
		User: sc.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(sc.Passwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", sc.Host, sc.Port), config)
}
func (sc *SftpConfig) Connect() (client *sftp.Client, err error) {
	fmt.Println("Connecting to", sc.Host)
	conn, err := sc.Dial()
	client, err = sftp.NewClient(conn)
	if err != nil {
		return
	}
	return
}
