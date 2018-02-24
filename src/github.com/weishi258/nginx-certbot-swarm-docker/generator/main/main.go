package main

import (
	"flag"
	"os"
	"fmt"
	. "github.com/weishi258/nginx-certbot-swarm-docker/generator/config"
	. "text/template"
)


var Version string
var BuildTime string

var outputPath string
var templatePath string
var certificateConfigPath string

func main(){

	var err error

	var printVer bool

	flag.BoolVar(&printVer, "version", false, "print generator version")
	flag.StringVar(&outputPath, "out", "/etc/nginx/config.d/default.conf", "nginx proxy config output path")
	flag.StringVar(&templatePath, "template", "/app/templates/default.tpl", "template path")
	flag.StringVar(&certificateConfigPath, "certs", "/etc/letsencrypt/certs.json", "certificates config path")
	flag.Parse()

	defer func(){
		if err != nil{
			os.Exit(1)
		}else{
			os.Exit(0)
		}
	}()
	fmt.Printf("[INFO] Nginx Generator Version: %s, BuildTime: %s\n", Version, BuildTime)
	if printVer{
		os.Exit(0)
	}

	var domains *Domains
	if domains, err = ParseDomains(certificateConfigPath); err != nil{
		fmt.Printf("[ERROR] parse domains config file failed, %s\n", err.Error())
		os.Exit(1)
	}


	var template *Template
	if template, err = ParseFiles(templatePath); err != nil{
		fmt.Printf("[ERROR] parse template file failed, %s\n", err.Error())
		os.Exit(1)
	}


	var outFile *os.File
	if outFile, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		fmt.Printf("[ERROR] Open output file %s failed, %s\n", outputPath, err.Error())
		os.Exit(1)
	}
	defer outFile.Close()


	if err = template.Execute(outFile, domains); err != nil{
		fmt.Printf("[ERROR] Execute template failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("[INFO] Generate nginx proxy config successful\n")

}
