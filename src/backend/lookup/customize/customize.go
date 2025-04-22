package customize

import (
	"fmt"
	"sync"
	"time"

	"typonamer/config"
	"typonamer/constant"
	"typonamer/log"
	"typonamer/lookup/lookuperror"
	"typonamer/lookup/lookupinfo"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/zh-five/golimit"
)

var limiterList = map[string]*golimit.GoLimit{}

func init() {
	SetupLimiter()
}

func SetupLimiter() {
	cfg := config.GetConfig()

	for _, api := range cfg.WhoisApis {
		limiterList[api.ApiName] = golimit.NewGoLimit(uint(api.ConcurrencyLimit))
	}
}

func CustomizeLookup(domain string, queryType string) (lookupinfo.DomainInfo, error) {
	var domainInfo = lookupinfo.DomainInfo{
		DomainName: domain,
		LookupType: queryType,
	}

	cfg := config.GetConfig()

	var apiInfo config.WhoisApi
	for _, api := range cfg.WhoisApis {
		if api.ApiName == queryType {
			apiInfo = api
			break
		}
	}

	if apiInfo.ApiName == "" {
		log.Debugf("Invalid query type: %s, no whois api found", queryType)
		return domainInfo, fmt.Errorf("%w: %s", lookuperror.ErrorInvalidQueryType, queryType)
	}

	apiUrl := strutil.TemplateReplace(apiInfo.ApiUrl, map[string]string{
		"domain": domain,
	})

	var response string
	var err error

	if maputil.HasKey(limiterList, apiInfo.ApiName) {
		limiter := limiterList[apiInfo.ApiName]

		var wg sync.WaitGroup
		wg.Add(1)

		limiter.Do(func() {
			defer wg.Done()
			response, err = getResponse(apiUrl)
		})

		wg.Wait()
	} else {
		response, err = getResponse(apiUrl)
	}

	if err != nil {
		log.Errorf("Request whois api %s response error: %v", apiInfo.ApiName, err)
		if response != "" {
			domainInfo.RawResponse = response
		} else {
			domainInfo.RawResponse = err.Error()
		}
		return domainInfo, err
	}

	domainInfo.RawResponse = response

	log.Debugf("Request whois api %s with domain %s, response: %s", apiInfo.ApiName, domain, response)

	if strutil.ContainsAny(response, apiInfo.FreeText) && strutil.ContainsAny(response, apiInfo.TakenText) {
		log.Errorf("Request whois api %s with domain %s, response both contains free and taken text", apiInfo.ApiName, domain)
		return domainInfo, fmt.Errorf("%w: %s", lookuperror.ErrorCustomizeApiWhoisResult, domain)
	} else if strutil.ContainsAny(response, apiInfo.FreeText) {
		log.Debugf("Request whois api %s with domain %s, response contains all free text", apiInfo.ApiName, domain)
		domainInfo.CustomizedResult = constant.DomainRegisterStatusFree
	} else if strutil.ContainsAny(response, apiInfo.TakenText) {
		log.Debugf("Request whois api %s with domain %s, response contains all taken text", apiInfo.ApiName, domain)
		domainInfo.CustomizedResult = constant.DomainRegisterStatusTaken
	} else {
		log.Errorf("Request whois api %s with domain %s, response not contains free or taken text", apiInfo.ApiName, domain)
		return domainInfo, fmt.Errorf("%w: %s", lookuperror.ErrorCustomizeApiWhoisResult, domain)
	}

	return domainInfo, nil
}

func getResponse(apiUrl string) (string, error) {
	cfg := config.GetConfig()

	client := resty.New()
	client.SetTimeout(time.Duration(cfg.WhoisTimeout) * time.Second)

	if cfg.RetryOnTimeout {
		client.SetRetryCount(cfg.RetryMax).
			SetRetryWaitTime(time.Duration(cfg.RetryInterval) * time.Second).
			SetRetryMaxWaitTime(time.Duration(cfg.RetryInterval) * time.Duration(cfg.RetryMax))
	}

	response, err := client.R().Get(apiUrl)
	if err != nil {
		return "", err
	}

	return response.String(), nil
}
