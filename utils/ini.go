package utils

import (
	"log"
	"strings"
	"sync"

	"gopkg.in/ini.v1"
)

var (
	iniFile *ini.File
	rwini   sync.RWMutex
)

// 初始化ini文件对象
// 第一参数: 可以是配置文件路径 或 []byte配置数据
func InitIniFile(source interface{}) (bool, error) {
	rwini.Lock()
	defer rwini.Unlock()
	file, e := ini.Load(source)
	if e != nil {
		log.Printf("Fail to load, Error:%v\n", e.Error())
		return false, e
	}
	iniFile = file
	return true, nil
}

func IniReadString(strSection, strKey string) string {
	rwini.RLock()
	defer rwini.RUnlock()

	strVal := iniFile.Section(strSection).Key(strKey).String()
	return strVal
}

func IniReadStringSlice(strSection, strKey string) []string {
	rwini.RLock()
	defer rwini.RUnlock()

	strVal := iniFile.Section(strSection).Key(strKey).String()
	sSlice := strings.Split(strVal, ",")
	return sSlice
}
func IniReadBool(strSection, strKey string) bool {
	rwini.RLock()
	defer rwini.RUnlock()

	isVal := iniFile.Section(strSection).Key(strKey).MustBool()
	return isVal
}

func IniReadInt64(strSection, strKey string) int64 {
	rwini.RLock()
	defer rwini.RUnlock()

	iVal := iniFile.Section(strSection).Key(strKey).MustInt64()
	return iVal
}

func IniReadInt(strSection, strKey string) int {
	rwini.RLock()
	defer rwini.RUnlock()

	iVal := iniFile.Section(strSection).Key(strKey).MustInt()
	return iVal
}
