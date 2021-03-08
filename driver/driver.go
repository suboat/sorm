package driver

import (
	"net/url"
)

// 通用数据库链接参数
type ArgConn struct {
	Driver string     `json:"driver"` //
	Params url.Values `json:"params"` //
	//
	Host     string `json:"host"`     //
	Port     string `json:"port"`     //
	Database string `json:"database"` //
	User     string `json:"user"`     //
	Password string `json:"password"` //
}
