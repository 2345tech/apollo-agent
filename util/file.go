package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

/*
给单独一个NS写独立的env配置文件
*/
func WriteSeparateEnv(name string, data map[string]string) {
	fileName := name + ".env"
	err := WriteEnvFile(data, fileName)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	}
}

/*
给单独一个NS写独立的ini配置文件
格式分解使用env
*/
func WriteSeparateIni(name string, data map[string]string) {
	fileName := name + ".ini"
	err := WriteEnvFile(data, fileName)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	}
}

/*
给单独一个NS写独立的php配置文件
*/
func WriteSeparatePHP(name string, data map[string]string) {
	fileName := name + ".php"
	arrayData := "<?php\n\nreturn "
	arrayData += StructTOPhpArray(data)
	arrayData += ";\n"
	err := ioutil.WriteFile(fileName, []byte(arrayData), 0644)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	}
}

/*
给单独一个NS写独立的text配置文件
*/
func WriteSeparateText(name, ext string, data map[string]string) {
	fileName := name + "." + ext
	if _, ok := data["content"]; !ok {
		log.Fatal("[ERROR] content not set")
	}
	err := ioutil.WriteFile(fileName, []byte(data["content"]), 0644)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	}

}

/*
将一个大map写入到env文件中
*/
func WriteDotEnv(allConfig map[string]map[string]string, namespaces []string, appId, path, fileName string) (string, string) {
	contentLines := make([]string, 0)
	// 遍历配置数据，拼接配置文件内容
	for _, namespace := range namespaces {
		if configData, ok := allConfig[appId+namespace]; ok {
			sortKeys := make([]string, 0, len(configData))
			for k := range configData {
				sortKeys = append(sortKeys, k)
			}
			sort.Strings(sortKeys) // 进行key自然升序

			// 写一行注释，提高.env文件可读性，用于快速区分namespace配置区块
			contentLines = append(contentLines, "###"+namespace+"###")

			for _, key := range sortKeys {
				contentLines = append(contentLines, fmt.Sprintf(`%s=%s`, key, configData[key]))
			}
			contentLines = append(contentLines, "\n") //增加一个换行进行区分
		}
	}

	configFileName := path + "/" + appId + ".env"
	if fileName != "" {
		configFileName = path + "/" + fileName
	}
	err := WriteContentIntoEnvFile(strings.Join(contentLines, "\n"), configFileName+".tmp")
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	} else {
		log.Printf("[INFO] [appId] %v config merge success. file is %v \n", appId, configFileName)
		return configFileName, configFileName + ".tmp"
	}
	return "", ""
}

/*
将一个大map写入到php文件中
*/
func WritePHP(allConfig map[string]map[string]string, namespaces []string, appId, path, fileName string) (string, string) {
	configData := make(map[string]map[string]string)
	for _, namespace := range namespaces {
		if data, ok := allConfig[appId+namespace]; ok {
			configData[strings.TrimPrefix(namespace, appId)] = data
		}
	}
	configFileName := path + "/" + appId + ".php"
	if fileName != "" {
		configFileName = path + "/" + fileName
	}
	arrayData := "<?php\n\nreturn "
	arrayData += StructTOPhpArray(configData)
	arrayData += ";\n"
	err := ioutil.WriteFile(configFileName+".tmp", []byte(arrayData), 0644)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	} else {
		log.Printf("[INFO] [appId] %v config merge success. file is %v \n", appId, configFileName)
		return configFileName, configFileName + ".tmp"
	}
	return "", ""
}

/*
将一个大map写入到text文件中
*/
func WriteText(allConfig map[string]map[string]string, namespaces []string, appId, path, fileName string) (string, string) {
	configData := make(map[string]map[string]string)
	for _, namespace := range namespaces {
		if data, ok := allConfig[appId+namespace]; ok {
			configData[strings.TrimPrefix(namespace, appId)] = data
		}
	}
	configFileName := path + "/" + appId + ".text"
	if fileName != "" {
		configFileName = path + "/" + fileName
	}
	textString := ""
	for _, data := range configData {
		if content, ok := data["content"]; ok {
			textString += content + "\n"
		}
	}
	err := ioutil.WriteFile(configFileName+".tmp", []byte(textString), 0644)
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	} else {
		log.Printf("[INFO] [appId] %v config merge success. file is %v \n", appId, configFileName)
		return configFileName, configFileName + ".tmp"
	}
	return "", ""
}

/*
将一个大map写入到ini文件中
*/
func WriteIni(allConfig map[string]map[string]string, namespaces []string, appId, path, fileName string) (string, string) {
	contentLines := make([]string, 0)
	// 遍历配置数据，拼接配置文件内容
	for _, namespace := range namespaces {
		if configData, ok := allConfig[appId+namespace]; ok {
			sortKeys := make([]string, 0, len(configData))
			for k := range configData {
				sortKeys = append(sortKeys, k)
			}
			sort.Strings(sortKeys) // 进行key自然升序

			// 根据namespace分配置区块
			contentLines = append(contentLines, "["+namespace+"]")

			for _, key := range sortKeys {
				contentLines = append(contentLines, fmt.Sprintf(`%s=%s`, key, configData[key]))
			}
			contentLines = append(contentLines, "\n") //增加一个换行进行区分
		}
	}

	configFileName := path + "/" + appId + ".ini"
	if fileName != "" {
		configFileName = path + "/" + fileName
	}
	err := WriteContentIntoEnvFile(strings.Join(contentLines, "\n"), configFileName+".tmp")
	if err != nil {
		log.Fatal("[ERROR] " + err.Error())
	} else {
		log.Printf("[INFO] [appId] %v config merge success. file is %v \n", appId, configFileName)
		return configFileName, configFileName + ".tmp"
	}
	return "", ""
}

/*
获取文件的md5 hash值
*/
func HashFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}

/*
复制文件
*/
func Copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}
	return nil
}
