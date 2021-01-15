package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

const (
	F_ENV  = "env"
	F_INI  = "ini"
	F_PHP  = "php"
	F_YAML = "yaml"
	F_YML  = "yml"
	F_XML  = "xml"
	F_TXT  = "txt"
)

const (
	FilePerm = 0644
)

func NSSyntax(namespace string) string {
	syntax := path.Ext(namespace)
	if syntax != "" {
		syntax = strings.Trim(syntax, ".")
	}
	switch strings.ToLower(syntax) {
	case F_ENV, F_INI, F_PHP, F_YAML, F_YML, F_XML, F_TXT:
		return strings.ToLower(syntax)
	default:
		return F_ENV
	}
}

// SingleNSInOneFile 将单独一个NS配置数据写入一个文件
func SingleNSInOneFile(fileName, suffix string, data map[string]string) error {
	var content string
	switch strings.ToLower(suffix) {
	case F_ENV, F_INI:
		content, _ = Marshal(data)
	case F_PHP:
		content = "<?php\n\nreturn " + GoTypeToPHPCode(data) + ";\n"
	case F_YAML, F_YML, F_XML, F_TXT:
		content = data["content"]
	}
	return writeFile(fileName, content, FilePerm)
}

// MultiNSInOneFile 将多个NS配置数据写入到一个文件中
func MultiNSInOneFile(fileName, suffix string, nss []string, multiData map[string]map[string]string) error {
	var content string
	switch strings.ToLower(suffix) {
	case F_ENV:
		content = multiDataToDotENV(multiData, nss)
	case F_INI:
		content = multiDataToINI(multiData, nss)
	case F_PHP:
		content = multiDataToPHP(multiData, nss)
	case F_YAML, F_YML, F_XML, F_TXT:
		content = multiDataToTXT(multiData, nss)
	}
	return writeFile(fileName, content, FilePerm)
}

// writeFile 将内容写入文件
func writeFile(filename, content string, perm os.FileMode) error {
	var m sync.Mutex
	m.Lock()
	defer m.Unlock()

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	_ = file.Sync()
	return err
}

// multiDataToDotENV 配置数据转换为dotEnv内容
func multiDataToDotENV(multiData map[string]map[string]string, nss []string) string {
	content := make([]string, 0)
	// 遍历配置数据，拼接配置文件内容
	for _, namespace := range nss {
		if data, ok := multiData[namespace]; ok {
			sortKeys := make([]string, 0, len(data))
			for k := range data {
				sortKeys = append(sortKeys, k)
			}
			sort.Strings(sortKeys) // 进行key自然升序

			// 写一行注释，提高.env文件可读性，用于快速区分namespace配置区块
			content = append(content, "###"+trimNSSuffix(namespace)+"###")

			for _, key := range sortKeys {
				content = append(content, fmt.Sprintf(`%s=%s`, key, data[key]))
			}
			content = append(content, "\n") //增加一个换行进行区分
		}
	}

	return strings.Join(content, "\n")
}

// multiDataToINI 配置数据转换为ini内容
func multiDataToINI(multiData map[string]map[string]string, nss []string) string {
	content := make([]string, 0)
	// 遍历配置数据，拼接配置文件内容
	for _, namespace := range nss {
		if data, ok := multiData[namespace]; ok {
			sortKeys := make([]string, 0, len(data))
			for k := range data {
				sortKeys = append(sortKeys, k)
			}
			sort.Strings(sortKeys) // 进行key自然升序

			// 根据namespace分配置区块
			content = append(content, "["+trimNSSuffix(namespace)+"]")

			for _, key := range sortKeys {
				content = append(content, fmt.Sprintf(`%s=%s`, key, data[key]))
			}
			content = append(content, "\n") //增加一个换行进行区分
		}
	}

	return strings.Join(content, "\n")
}

// multiDataToPHP 配置数据转换为ini内容
func multiDataToPHP(multiData map[string]map[string]string, nss []string) string {
	content := make(map[string]map[string]string)
	for _, namespace := range nss {
		if data, ok := multiData[namespace]; ok {
			content[trimNSSuffix(namespace)] = data
		}
	}

	return "<?php\n\nreturn " + GoTypeToPHPCode(content) + ";\n"
}

// multiDataToTXT 配置数据转换为txt内容
func multiDataToTXT(multiData map[string]map[string]string, nss []string) string {
	content := make([]string, 0)
	for _, namespace := range nss {
		if data, ok := multiData[namespace]; ok {
			if row, ok := data["content"]; ok {
				content = append(content, row)
			}
		}
	}

	return strings.Join(content, "\n")
}

// trimNSSuffix 切除namespace后缀
func trimNSSuffix(namespace string) string {
	ns := strings.Split(namespace, ".")
	if len(ns) > 0 {
		return ns[0]
	} else {
		return namespace
	}
}

// HashFileMd5 获取文件md5值
func HashFileMd5(filePath string) (string, error) {
	var md5Sum string
	file, err := os.Open(filePath)
	if err != nil {
		return md5Sum, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return md5Sum, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	md5Sum = hex.EncodeToString(hashInBytes)
	return md5Sum, nil

}

// CopyFile 复制文件
func CopyFile(sourceFile, toNewFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	return writeFile(toNewFile, string(input), FilePerm)
}
