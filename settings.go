package main

type SettingsType struct {
	WebPort            int    `json:"web_port"`
	LocalPort          int    `json:"local_port"`
	DatabaseServer     string `json:"database_server"`
	DatabasePort       string `json:"database_port"`
	DatabaseName       string `json:"database_name"`
	DatabaseLogin      string `json:"database_login"`
	DatabasePassword   string `json:"database_password"`
	WebFiles           string `json:"web_files"`
	SSLCertificateFile string `json:"ssl_certificate_file"`
	SSLPrivateKeyFile  string `json:"ssl_private_key_file"`
	SerialPort         string `json:"serial_port"`
}
