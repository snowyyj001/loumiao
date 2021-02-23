package llog

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/snowyyj001/loumiao/config"
)

//压缩文件
//files 文件数组，可以是不同dir下的文件或者文件夹
//dest 压缩文件存放地址
func compress(filenames []string, dest string) error {
	d, e := os.Create(dest)
	if e != nil {
		fmt.Errorf(e.Error())
	}
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, name := range filenames {
		file, err := os.Open(name)
		if err == nil {
			doCompress(file, "", w)
			os.Remove(name)
		} else {
			fmt.Errorf(err.Error())
		}
	}
	return nil
}

func doCompress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = doCompress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//将旧的文件压缩到zip中备份
func dealLogs(filename string) {
	var files []string

	_, err := os.Stat(filename)
	if err != nil {
		return
	}

	filepath.Walk(fmt.Sprintf("logs/%s", config.SERVER_NAME), func(pathname string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if runtime.GOOS == "windows" {
			if strings.HasPrefix(pathname, fmt.Sprintf("logs\\%s\\%s.", config.SERVER_NAME, config.SERVER_NAME)) && strings.HasSuffix(pathname, ".llog") {
				files = append(files, pathname)
			}
		} else {
			if strings.HasPrefix(pathname, fmt.Sprintf("logs/%s/%s.", config.SERVER_NAME, config.SERVER_NAME)) && strings.HasSuffix(pathname, ".llog") {
				files = append(files, pathname)
			}
		}

		return nil
	})
	strtime := time.Now().Format("2006-01-02.15.04.05") //":"不能作为文件名
	zipfile := fmt.Sprintf("logs/%s/%s-%s.zip", config.SERVER_NAME, config.SERVER_NAME, strtime)
	compress(files, zipfile)
}
