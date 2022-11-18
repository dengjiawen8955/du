package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

type DownLoadInterface interface {
	// 下载, 返回本地绝对路径
	DownLoad(url string) (string, error)
}

type DownLoadImpl struct {
	dir string // 保存的目录
}

func NewDownLoad(dir string) DownLoadInterface {
	return &DownLoadImpl{dir: dir}
}

func (d *DownLoadImpl) DownLoad(url string) (string, error) {
	// 获取文件路径
	filePath, err := d.GetFilePath(url)
	if err != nil {
		return "", err
	}

	// 下载
	err = downloadFile(url, filePath, func(length, downLen int64) {
		fmt.Println(length, downLen)
	})

	return filePath, err
}

func downloadFile(url string, localPath string, fb func(length, downLen int64)) error {
	var (
		fsize   int64
		buf     = make([]byte, 32*1024)
		written int64
	)
	tmpFilePath := localPath + ".download"
	fmt.Println(tmpFilePath)
	//创建一个http client
	client := new(http.Client)
	//client.Timeout = time.Second * 60 //设置超时时间
	//get方法获取资源
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	//读取服务器返回的文件大小
	fsize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		fmt.Println(err)
	}
	//创建文件
	file, err := os.Create(tmpFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	if resp.Body == nil {
		return errors.New("body is null")
	}
	defer resp.Body.Close()
	//下面是 io.copyBuffer() 的简化版本
	for {
		//读取bytes
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			//写入bytes
			nw, ew := file.Write(buf[0:nr])
			//数据长度大于0
			if nw > 0 {
				written += int64(nw)
			}
			//写入出错
			if ew != nil {
				err = ew
				break
			}
			//读取是数据长度不等于写入的数据长度
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		//没有错误了快使用 callback
		fb(fsize, written)
	}
	fmt.Println(err)
	if err == nil {
		file.Close()
		err = os.Rename(tmpFilePath, localPath)
		fmt.Println(err)
	}
	return err
}

func (d *DownLoadImpl) GetFilePath(url string) (string, error) {
	// 检查目录是否存在不存在就新建
	if _, err := os.Stat(d.dir); os.IsNotExist(err) {
		err = os.MkdirAll(d.dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	// 获取文件名
	fileName := path.Base(url)
	filePath := path.Join(d.dir, fileName)

	// 检查文件是否存在, 如果存在添加序号
	for i := 1; FileExist(filePath); i++ {
		fileName = fmt.Sprintf("%d_", i) + fileName
		filePath = path.Join(d.dir, fileName)
	}

	return filePath, nil
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}