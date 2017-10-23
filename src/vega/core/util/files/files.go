// Copyright (c) 2016, Virtual Gateway Labs. All rights reserved.

// Package files provides utility functions to fetch and manipulate files
package files

import (
//	"crypto/md5"
//	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"os/exec"

	"github.com/dutchcoders/goftp"
)

const (
	protocol_ftp  = "ftp://"
	protocol_http = "http://"

	PROTOCOL_FTP  = "ftp"
	PROTOCOL_HTTP = "http"
)

type Usage struct {
	BytesUsed int64  `json:"bytes_used"`
	BytesFree uint64 `json:"bytes_free"`
	FileCount int    `json:"file_count"`
}

type Checksum struct {
	MD5  string `json:"md5"`
	SHA1 string `json:"sha1"`
}

type FileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	ModTime  string `json:"modified"`
}

type FileRename struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type FileDownload struct {
	size_total  int64
	path        string
	downloading bool
	err         error
	Closed      bool
	client      *goftp.Client
	resp        *http.Response
}

func VerifyRenames(oldDirPath, newDirPath string, renames []FileRename) []error {
	errs := []error{}

	// check all tasks first and convert them to full paths
	for _, rename := range renames {
		if rename.OldName == "" {
			errs = append(errs, errors.New("source name cannot be empty"))
			return errs
		}

		if rename.NewName == "" {
			rename.NewName = rename.OldName
		}

		oldPath := oldDirPath + "/" + rename.OldName
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			errs = append(errs, errors.New(fmt.Sprintf("[%v] does not exist", rename.OldName)))
			continue
		}

		newPath := newDirPath + "/" + rename.NewName
		if _, err := os.Stat(newPath); err == nil && oldPath != newPath {
			errs = append(errs, errors.New(fmt.Sprintf("[%v] already exist", rename.NewName)))
			continue
		}
	}

	return errs
}

func getFileInfo(full_path string) (fileinfo os.FileInfo, err error) {
	file, err := os.Open(full_path)
	if err != nil {
		return
	}
	defer file.Close()

	fileinfo, err = file.Stat()
	if err != nil {
		return
	}

	if fileinfo.IsDir() {
		err = errors.New("Target is a directory")
		return
	}

	return
}

func (fd *FileDownload) Abort() (err error) {
	if fd.client != nil {
		fd.err = fd.client.Close()
		if fd.err != nil {
			return fd.err
		}
	} else if fd.resp != nil {
		fd.err = fd.resp.Body.Close()
		if fd.err != nil {
			return fd.err
		}
	}

	fd.Closed = true

	fd.err = os.Remove(fd.path)
	if fd.err != nil {
		return fd.err
	}

	return
}

func (fd *FileDownload) GetDLProgress() (size_on_disk int64, progress uint8, err error) {

	if fd.err != nil {
		return 0, 0, fd.err
	}

	if !fd.downloading {
		return 0, 0, nil
	}

	fi, err := getFileInfo(fd.path)
	if err != nil {
		return
	}

	size_on_disk = fi.Size()
	progress = uint8(float32(size_on_disk) / float32(fd.size_total) * 100)

	return
}

func Download(location, protocol, username, password, dirpath, filename string) (fd *FileDownload, err error) {
	_, errf := GetFileInfo(dirpath, filename)
	if errf == nil {
		return nil, fmt.Errorf("File %s already exist", filename)
	}

	dest_path := dirpath + "/" + filename

	fd = new(FileDownload)
	if protocol == "ftp" {
		location = strings.TrimPrefix(location, protocol_ftp)
		url_slice := strings.SplitN(location, "/", 2)

		go download_ftp(fd, url_slice[0], url_slice[1], username, password, dest_path)
	} else if protocol == PROTOCOL_HTTP || protocol == "" {
		go download_http(fd, location, username, password, dest_path)
	} else {
		err = errors.New("Unknown protocol")
	}
	runtime.Gosched()

	return fd, err
}

func download_ftp(fd *FileDownload, host string, remote_path string,
	username string, password string, dest string) {

	defer func() {
		recover()
	}()

	config := goftp.Config{
		User:               username,
		Password:           password,
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             os.Stderr,
	}

	fd.downloading = false
	fd.path = dest

	ftp, err := goftp.DialConfig(config, host)
	if err != nil {
		println("conn", err.Error())
		fd.err = err
		return
	}
	defer ftp.Close()

	fi, err := ftp.Stat(remote_path)
	if err != nil {
		println("stat", err.Error())
		fd.err = err
		return
	}

	fd.size_total = fi.Size()
	fd.client = ftp

	file, err := os.OpenFile(dest, syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL, 0666)
	if err != nil {
		println("open", err.Error())
		fd.err = err
		return
	}
	defer file.Close()

	fd.downloading = true
	err = ftp.Retrieve(remote_path, file)
	if err != nil {
		println("recv", err.Error())
		fd.err = err
		defer os.Remove(dest)
	}
}

func download_http(fd *FileDownload, url string, username string, password string,
	dest string) {

	defer func() {
		recover()
	}()

	fd.downloading = false
	fd.path = dest

	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}

	resp, err := http.Get(url)

	defer func() {
		fd.err = err
	}()

	if err != nil {
		return
	}
	defer resp.Body.Close()

	fd.resp = resp

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	fd.size_total = int64(size)
	if err != nil {
		return
	}

	file, err := os.OpenFile(dest, syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	fd.downloading = true
	n, err := io.Copy(file, resp.Body)
	fd.size_total = n
	if err != nil {
		defer os.Remove(dest)
	}
}

func Exist(dirPath, filename string) bool {
	path := dirPath + "/" + filename
	_, err := os.Stat(path)
	return err == nil
}

func NotExist(dirPath, filename string) bool {
	path := dirPath + "/" + filename
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func Renames(oldDirPath, newDirPath string, renames []FileRename) (errs []error) {
	// check all tasks first and convert them to full paths
	for i, rename := range renames {
		if rename.NewName == "" {
			rename.NewName = rename.OldName
		}

		rename.OldName = oldDirPath + "/" + rename.OldName
		rename.NewName = newDirPath + "/" + rename.NewName
		renames[i] = rename
	}

	if len(errs) > 0 {
		return errs
	}

	for _, rename := range renames {
		if rename.OldName == rename.NewName {
			// skip same paths
			continue
		}

		if err := os.Rename(rename.OldName, rename.NewName); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func DirSize(dirPath string) (size int64, err error) {
	err = filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func CountFile(dirPath string) (int, error) {
	files, err := ioutil.ReadDir(dirPath)

	return len(files), err
}

func GetUsage(dirPath string) (Usage, error) {
	usage := Usage{}

	if size, err := DirSize(dirPath); err != nil {
		return usage, err
	} else {
		usage.BytesUsed = size
	}

	if count, err := CountFile(dirPath); err != nil {
		return usage, err
	} else {
		usage.FileCount = count
	}

	if free, err := FreeSpace(dirPath); err != nil {
		return usage, err
	} else {
		usage.BytesFree = free
	}

	return usage, nil
}

func List(rootDirPath string) (fileInfos []FileInfo, err error) {
	fileInfos = []FileInfo{}

	files, err := ioutil.ReadDir(rootDirPath)
	if err != nil {
		return
	}

	for _, file := range files {
		var fileInfo FileInfo

		fileInfo.Filename = file.Name()
		fileInfo.Size = file.Size()

		m_time := file.ModTime()
		fileInfo.ModTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
			m_time.Year(), m_time.Month(), m_time.Day(),
			m_time.Hour(), m_time.Minute(), m_time.Second())

		fileInfos = append(fileInfos, fileInfo)
	}

	return
}

func Delete(dirPath, filename string) (err error) {
	full_path := dirPath + "/" + filename
	return os.Remove(full_path)
}

func GetMD5(dirPath, filename string) (checksum string, err error) {
	//data, err := ioutil.ReadFile(dirPath + "/" + filename)
	Cmd := fmt.Sprintf("md5sum %s", dirPath + "/" + filename)
	standout, err := exec.Command("/bin/sh","-c",Cmd).CombinedOutput()
	if err != nil {
		return "", err
	}
	ret := strings.Split(string(standout), " ")
    return fmt.Sprintf("%s", ret[0]), nil
}

func GetSHA1(dirPath, filename string) (checksum string, err error) {
	//data, err := ioutil.ReadFile(dirPath + "/" + filename)
	Cmd := fmt.Sprintf("sha1sum %s", dirPath + "/" + filename)
	standout, err := exec.Command("/bin/sh","-c",Cmd).CombinedOutput()
	if err != nil {
		return "", err
	}
	ret := strings.Split(string(standout), " ")
	return fmt.Sprintf("%s", ret[0]), nil
}

func FreeSpace(rootPath string) (size uint64, err error) {
	var stat syscall.Statfs_t

	err = syscall.Statfs(rootPath, &stat)

	size = stat.Bavail * uint64(stat.Bsize)

	return size, err
}

func Upload(dirPath, filename string, rd io.Reader) (err error) {
	file, err := os.OpenFile(dirPath+"/"+filename, syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	io.Copy(file, rd)

	return
}

func GetFileInfo(dirPath, filename string) (fileInfo FileInfo, err error) {

	full_path := dirPath + "/" + filename
	fi, err := getFileInfo(full_path)
	if err != nil {
		return
	}

	fileInfo.Filename = fi.Name()
	fileInfo.Size = fi.Size()

	m_time := fi.ModTime()
	fileInfo.ModTime = fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ",
		m_time.Year(), m_time.Month(), m_time.Day(),
		m_time.Hour(), m_time.Minute(), m_time.Second())

	return
}
