package config

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
	. "regexp"
	"fmt"
)
const DOMAINS = "DOMAINS"
const CERTBOT = "CERTBOT"

const REGEX_DOMAIN = "^[a-z0-9]([a-z0-9-]+\\.){1,}[a-z0-9]+$"

type Domain struct {
	Domain 		string
	Proxy 		string
	Encryption	bool
	SslReady 	bool
}
type Domains struct{
	Certbot string
	Domains []Domain
}
type Cert struct{
	Domain 		string
	SslReady 	bool
}
type Certs struct{
	Domains []Cert
}

func ParseDomains(certsConfigPath string)( ret *Domains, err error){

	str := os.Getenv(DOMAINS)
	domains := strings.Split(str, ",")
	if len(domains) == 0{
		return nil, errors.New("Domains is empty")
	}
	var domainRegex *Regexp
	if domainRegex, err = Compile(REGEX_DOMAIN); err != nil{
		return nil, err
	}
	ret = &Domains{}
	ret.Domains = make([]Domain, 0)
	for _, elem := range domains{
		temps := strings.Split(elem, "=")
		if len(temps) != 2{
			fmt.Printf("Invalid format for %s\n", elem)
			continue
		}
		temps[0] = strings.TrimSpace(temps[0])
		temps[1] = strings.TrimSpace(temps[1])
		if !domainRegex.MatchString(temps[0]){
			fmt.Printf("Invalid domain format for %s\n", temps[0])
			continue
		}
		ret.Domains = append(ret.Domains, Domain{temps[0], temps[1], false, false})
		fmt.Printf("[INFO] Add domain: %s, proxy: %s\n", temps[0], temps[1])
	}
	if len(ret.Domains) == 0{
		return nil, errors.New("Domains is empty")
	}
	ret.Certbot = os.Getenv(CERTBOT)
	if len(ret.Certbot) != 0{
		fmt.Printf("[INFO] Certbot is enabled and passed to container: %s\n", ret.Certbot)
	}else{
		fmt.Printf("[INFO] Certbot is disabled\n")
	}



	//
	var certs *Certs
	if certs, err = ParseCerts(certsConfigPath); err != nil{
		fmt.Printf("[INFO] Can not parse certificate config files, so no SSL encryption, %s\n", err.Error())
		return ret, nil
	}
	for i := 0; i < len(ret.Domains); i++{
		for _, dd := range certs.Domains{
			if strings.Compare(ret.Domains[i].Domain, dd.Domain) == 0 {
				fmt.Printf("[INFO] Set domain: %s encryption to TRUE\n", ret.Domains[i].Domain)
				ret.Domains[i].Encryption = true
				ret.Domains[i].SslReady = dd.SslReady
			}
		}
	}


	return ret, nil
}

func ParseCerts(path string)( ret *Certs, err error){
	file, err := os.Open(path) // For read access.
	if err != nil {
		return nil, errors.Wrapf(err, "Open certificates config file %s failed", path)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Read certificates config file %s failed", path)
	}

	ret = &Certs{}
	if err = json.Unmarshal(data, ret); err != nil {
		return nil, errors.Wrapf(err, "Parse certificates config file %s failed", path)
	}
	return
}
