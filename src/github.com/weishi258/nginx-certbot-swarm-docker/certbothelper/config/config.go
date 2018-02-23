package config

import (
	gen "github.com/weishi258/nginx-certbot-swarm-docker/generator/config"
	"os"
	"strings"
	"fmt"
	"github.com/pkg/errors"
	. "regexp"
	"encoding/json"
)




type Domains struct{
	Domains []string
}

func GetCertConfig(certsConfigPath string)( ret *gen.Certs, bRefresh bool, err error){
	var domains *Domains
	if domains, err = parseDomains(); err != nil {
		return nil, false, err
	}
	ret = &gen.Certs{}
	ret.Domains = make([]gen.Cert, len(domains.Domains))
	for i := 0; i < len(ret.Domains); i++{
		ret.Domains[i].Domain = domains.Domains[0]
	}

	var certs *gen.Certs
	if certs, err = gen.ParseCerts(certsConfigPath); err != nil{
		fmt.Printf("[INFO] %s\n", err.Error())
		certs = &gen.Certs{}
		certs.Domains = make([]gen.Cert, 0)
	}
	// lets do compare
	sameCount := 0
	for i := 0; i < len(ret.Domains); i++{
		ret.Domains[i].Domain = domains.Domains[i]
		for _, elem := range certs.Domains{
			if strings.Compare(ret.Domains[i].Domain, elem.Domain) == 0{
				ret.Domains[i].SslReady = elem.SslReady
				sameCount++
				break
			}
		}
	}

	return ret, !(sameCount == len(ret.Domains) && sameCount == len(certs.Domains)), nil
}

func WriteCerts(path string, certs *gen.Certs) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil{
		return errors.Wrapf(err, "Open to write failed")
	}
	defer file.Close()

	var outputStr []byte
	if outputStr, err = json.Marshal(certs); err != nil{
		return errors.Wrapf(err, "Marshal json failed")
	}
	if _, err = file.Write(outputStr); err != nil{
		return errors.Wrapf(err, "Write to file failed")
	}
	return nil
}

func parseDomains()( ret *Domains, err error){
	str := os.Getenv(gen.DOMAINS)
	domains := strings.Split(str, ",")
	if len(domains) == 0{
		return nil, errors.New("Domains is empty")
	}
	var domainRegex *Regexp
	if domainRegex, err = Compile(gen.REGEX_DOMAIN); err != nil{
		return nil, err
	}
	ret = &Domains{}
	ret.Domains = make([]string, 0)
	for _, elem := range domains{
		elem = strings.TrimSpace(elem)
		if !domainRegex.MatchString(elem){
			fmt.Printf("Invalid domain format for %s\n", elem)
			continue
		}
		ret.Domains = append(ret.Domains, elem)
		fmt.Printf("[INFO] Add domain: %s\n", elem)
	}
	if len(ret.Domains) == 0{
		return nil, errors.New("Domains is empty")
	}


	return ret, nil
}