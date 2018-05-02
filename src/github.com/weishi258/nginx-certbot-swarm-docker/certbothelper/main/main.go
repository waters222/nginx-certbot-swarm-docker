package main

import (
	"os"
	"os/signal"
	"syscall"
	"flag"
	"fmt"
	"time"
	gen "github.com/weishi258/nginx-certbot-swarm-docker/generator/config"
	. "github.com/weishi258/nginx-certbot-swarm-docker/certbothelper/config"
	"os/exec"
	"strconv"
	"sync"
)


var Version string
var BuildTime string

var sleepInterval int
var sigChan chan os.Signal

var certificateConfigPath string
var certificateDirectory string
var staging bool
var email string




const (
	EMAIL = "EMAIL"
	STAGING = "STAGING"
	PAUSING_TIME = 10
	DEFAULT_SLEEP_INTERVAL = 3600
)
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
	flag.IntVar(&sleepInterval, "interval", DEFAULT_SLEEP_INTERVAL, "file modify watcher sleepInterval in seconds")
	flag.StringVar(&certificateConfigPath, "certs", "/etc/letsencrypt/certs.json", "certificates config path")
	flag.StringVar(&certificateDirectory, "d", "/etc/letsencrypt/live", "certificates base directory")
	flag.Parse()
	if len(certificateDirectory) == 0{
		certificateDirectory = "/etc/letsencrypt/live"
	}
	defer func() {
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()
	fmt.Printf("[INFO] CertbotHelper Version: %s, BuildTime: %s\n", Version, BuildTime)
	if printVer {
		os.Exit(0)
	}


	email = os.Getenv(EMAIL)
	if len(email) == 0{
		fmt.Printf("[ERROR] Email address empty\n")
		os.Exit(1)
	}
	if staging, err = strconv.ParseBool(os.Getenv(STAGING)); err != nil{
		fmt.Printf("[INFO] Staging env variable parse error %s\n", os.Getenv(STAGING))
	}
	if staging{
		fmt.Printf("[INFO] Certbot is in staging mode\n")
	}else{
		fmt.Printf("[INFO] Certbot is in production mode\n")
	}

	// make sure it always positive
	if sleepInterval <= 0{
		sleepInterval = 1
	}
	certModTime := make(map[string]time.Time)
	certModTimeMutex := &sync.Mutex{}

	duration := time.Duration(sleepInterval) * time.Second

	timer := time.NewTimer(duration)

	process(certModTime, certModTimeMutex)

	go func(){
		for range timer.C{
			process(certModTime, certModTimeMutex)
			timer.Reset(duration)
		}
	}()



	go func() {
		sig := <-sigChan
		fmt.Printf("[INFO] CertbotHelper caught signal %v for exit", sig)
		timer.Stop()
		done <- true
	}()
	<-done
}


func process(certModTime map[string]time.Time, certModeTimeMutex *sync.Mutex){
	var err error
	var bRefresh bool
	var certs *gen.Certs
	if certs, bRefresh, err = GetCertConfig(certificateConfigPath); err != nil{
		fmt.Printf("[ERROR] Failed to process domains, %s\n", err.Error())
		return
	}

	sslNeedDomainsIdx := make([]int, 0)
	sslReadyDomainsIdx := make([]int, 0)

	for i := 0; i < len(certs.Domains); i++{
		if certs.Domains[i].SslReady{
			sslReadyDomainsIdx = append(sslReadyDomainsIdx, i)
		}else{
			sslNeedDomainsIdx = append(sslNeedDomainsIdx, i)
		}
	}
	if bRefresh{
		bRefresh = false
		if err = WriteCerts(certificateConfigPath, certs); err != nil{
			fmt.Printf("[ERROR] Write to certificate file failed %s, %s\n", certificateConfigPath, err.Error())
		}else{
			fmt.Printf("[INFO] Write to certificate file %s successful\n", certificateConfigPath)
		}
		time.Sleep(PAUSING_TIME * time.Second)
		fmt.Printf("[INFO] Wait for %d seconds for nginx to pick up changes\n",PAUSING_TIME)

	}
	if len(sslNeedDomainsIdx) > 0{

		// populate domains string list
		certbotArgs := make([]string, 5)
		certbotArgs[0] = "certonly"
		certbotArgs[1] = "-m"
		certbotArgs[2] = email
		certbotArgs[3] = "--agree-tos"
		certbotArgs[4] = "--non-interactive"

		if staging{
			certbotArgs = append(certbotArgs, "--staging")
		}

		certbotArgs = append(certbotArgs, "--webroot")
		certbotArgs = append(certbotArgs, "-w")
		certbotArgs = append(certbotArgs, "/usr/share/nginx/html")

		for _, idx := range sslNeedDomainsIdx{
			certbotArgs = append(certbotArgs, "-d")
			certbotArgs = append(certbotArgs, certs.Domains[idx].Domain)
		}

		//cmd := exec.Command("certbot","-webroot", "-w", "/usr/share/nginx/html", "-d", sslNeedDomains...)
		cmd := exec.Command("certbot", certbotArgs...)
		if consoleOut, err := cmd.Output(); err != nil{
			fmt.Printf("[ERROR] Exec certbot failed, %s\n", err.Error())
			fmt.Printf("%s",consoleOut)
		}else{
			bRefresh = true
			fmt.Printf("%s",consoleOut)
			for _, idx := range sslNeedDomainsIdx{
				certs.Domains[idx].SslReady = true
			}
		}
	}
	if len(sslReadyDomainsIdx) > 0{
		// lets check if certificate has been update recently
		bNotifyNginxProxy := false
		for _, cert := range certs.Domains {
			if stat, err := os.Stat(fmt.Sprintf("%s/%s", certificateDirectory, cert.Domain)); err == nil{
				lastModTime := stat.ModTime()
				certModeTimeMutex.Lock()
				if domainModTime, ok := certModTime[cert.Domain]; ok{
					if lastModTime.After(domainModTime) {
						certModTime[cert.Domain] = lastModTime
						fmt.Printf("[INFO] Certificate for domain %s has changed", cert.Domain)
						bNotifyNginxProxy = true
					}
				}else{
					fmt.Printf("[INFO] Certificate for domain %s has added", cert.Domain)
					certModTime[cert.Domain] = lastModTime
					bNotifyNginxProxy = true
				}
				certModeTimeMutex.Unlock()
			}
		}

		certbotArgs := make([]string, 0)
		if staging{
			certbotArgs = append(certbotArgs, "--staging")
		}
		certbotArgs = append(certbotArgs, "renew")
		if bRefresh || bNotifyNginxProxy {
			certbotArgs = append(certbotArgs, "--post-hook")
			certbotArgs = append(certbotArgs, fmt.Sprintf("touch %s", certificateConfigPath))
		}
		cmd := exec.Command("certbot", certbotArgs...)
		if consoleOut, err := cmd.Output(); err != nil{
			fmt.Printf("[ERROR] Exec certbot renew failed, %s\n", err.Error())
			fmt.Printf("%s",consoleOut)
			return
		}else{
			fmt.Printf("%s",consoleOut)
		}
	}
	if bRefresh{
		if err = WriteCerts(certificateConfigPath, certs); err != nil{
			fmt.Printf("[ERROR] Write to certificate file failed %s, %s\n", certificateConfigPath, err.Error())
		}else{
			fmt.Printf("[INFO] Write to certificate file %s successful\n", certificateConfigPath)
		}
	}
	fmt.Printf("[INFO] Sleep for %d seconds\n", sleepInterval)


}