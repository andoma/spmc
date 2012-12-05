package main

import "os"
import "encoding/json"
import "fmt"

type DbConfig struct {
	Addr string `json:"address"`;
	User string `json:"username"`;
	Pass string `json:"password"`;
	Name string `json:"database"`;
};


type MailConfig struct {
	Sender string `json:"sender"`;
	Server string `json:"server"`;
};

type Config struct {
	Db DbConfig  `json:"db"`;
	Mail MailConfig  `json:"mail"`;
}

var config *Config;


func loadConf() {

	if config != nil {
		return;
	}

	file, err := os.Open("config.json");
	if err != nil {
		fmt.Printf("Unable to open config file\n");
		os.Exit(1);
	}

	defer file.Close();

	config = new(Config);
	err = json.NewDecoder(file).Decode(config);
	if err != nil {
		fmt.Printf("Unable to parse config file\n");
		os.Exit(1);
	}
}