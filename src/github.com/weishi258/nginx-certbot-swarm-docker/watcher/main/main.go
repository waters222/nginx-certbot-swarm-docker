package main

import (
	"flag"
	"os"
	"fmt"
	"os/signal"
	"syscall"
	"time"
	"os/exec"
	"strings"
)

var Version string
var BuildTime string

var outputPath string
var certificateConfigPath string
var templatePath string
var generatorPath string
var nginxReload string

var interval int
var sigChan chan os.Signal


var nginxModifed int64 = 0
var certsModifed int64 = 0

var nginxReloadArgs []string

func main() {
	// waiting loop for signal
	sigChan = make(chan os.Signal, 5)
	done := make(chan bool)

	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGINT)


	var err error

	var printVer bool

	flag.BoolVar(&printVer, "version", false, "print watcher version")
	flag.IntVar(&interval, "interval", 1, "file modify watcher interval in seconds")
	flag.StringVar(&outputPath, "out", "/etc/nginx/conf.d/default.conf", "nginx proxy config output path")
	flag.StringVar(&templatePath, "template", "/app/templates/default.tpl", "template path")
	flag.StringVar(&certificateConfigPath, "certs", "/etc/letsencrypt/certs.json", "certificates config path")

	flag.StringVar(&generatorPath, "generator", "/app/generator", "generator app path")
	flag.StringVar(&nginxReload, "nginx", "nginx -s reload", "/usr/sbin/nginx reload command")
	flag.Parse()

	defer func() {
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	fmt.Printf("[INFO] Nginx Watcher Version: %s, BuildTime: %s\n", Version, BuildTime)
	if printVer {
		os.Exit(0)
	}

	fmt.Printf("[INFO] Watching nginx config file %s\n", outputPath)
	fmt.Printf("[INFO] Watching certificate config file %s\n", certificateConfigPath)


	nginxReloadArgs = strings.Split(nginxReload, " ")
	if len(nginxReloadArgs) < 1 {
		fmt.Printf("[ERROR] nginx reload command invalid\n")
		os.Exit(1)
	}

	// make sure it always positive
	if interval <= 0{
		interval = 1
	}

	duration := time.Duration(interval) * time.Second

	timer := time.NewTimer(duration)
	go func(){
		for range timer.C{
			process()
			timer.Reset(duration)
		}
	}()

	go func() {
		sig := <-sigChan
		fmt.Printf("[INFO] Watcher caught signal %v for exit", sig)
		timer.Stop()
		done <- true
	}()
	<-done

}

func process(){
	_certsModifed := checkFileLastUpdated(certificateConfigPath)
	_nginxModifed := checkFileLastUpdated(outputPath)

	// make sure always generate nginx default conf no matter what
	if _certsModifed != certsModifed || _nginxModifed == -1{
		fmt.Printf("[INFO] Certificates config %s modified from %v to %v\n", certificateConfigPath, certsModifed, _certsModifed)
		cmd := exec.Command(generatorPath, "-out", outputPath, "-certs", certificateConfigPath, "-template", templatePath)
		if consoleOut, err := cmd.Output(); err != nil{
			fmt.Printf("[ERROR] Exec generator failed, %s\n", err.Error())
			fmt.Printf("%s",consoleOut)
			return
		}else{
			fmt.Printf("%s",consoleOut)
		}
		certsModifed = _certsModifed
	}

	if _nginxModifed != nginxModifed {
		fmt.Printf("[INFO] Nginx config modified from %v to %v\n", nginxModifed, _nginxModifed)
		var cmd *exec.Cmd
		if len(nginxReloadArgs) > 1{
			cmd = exec.Command(nginxReloadArgs[0], nginxReloadArgs[1:]...)
		}else{
			cmd = exec.Command(nginxReloadArgs[0])
		}

		if consoleOut, err := cmd.Output(); err != nil{
			fmt.Printf("[ERROR] Exec nginx reload failed, %s\n", err.Error())
			fmt.Printf("%s",consoleOut)
			return
		}else{
			fmt.Printf("%s",consoleOut)
		}
		nginxModifed = _nginxModifed
	}



}

func checkFileLastUpdated(path string) int64{
	if fileInfo, err := os.Stat(path); err != nil{
		if !os.IsNotExist(err){
			fmt.Printf("[ERROR] Get file %s info failed, %s\n", outputPath, err.Error())
		}
		return -1
	}else{
		return fileInfo.ModTime().UnixNano()
	}
}
