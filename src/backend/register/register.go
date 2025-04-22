package register

import (
	"fmt"
	"sync"
	"time"

	"typonamer/config"
	"typonamer/constant"
	"typonamer/log"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/zh-five/golimit"
)

var limiterList = map[string]*golimit.GoLimit{}

const (
	defaultRegisterTimeout = 10 * time.Second
)

func init() {
	SetupLimiter()
}

func SetupLimiter() {
	cfg := config.GetConfig()

	for _, api := range cfg.RegisterApis {
		limiterList[api.ApiName] = golimit.NewGoLimit(uint(api.ConcurrencyLimit))
	}
}

func Register(domain string, registerType string) (RegisterInfo, error) {
	registerInfo := RegisterInfo{
		RegisterType: registerType,
		DomainName:   domain,
	}
	cfg := config.GetConfig()

	var apiInfo config.RegisterApi
	for _, api := range cfg.RegisterApis {
		if api.ApiName == registerType {
			apiInfo = api
			break
		}
	}

	if apiInfo.ApiName == "" {
		log.Debugf("Invalid register type: %s, no register api found", registerType)
		registerInfo.RegisterStatus = constant.RegisterStatusError
		registerInfo.RawResponse = fmt.Sprintf("Invalid register type: %s, no register api found", registerType)
		return registerInfo, fmt.Errorf("%w: %s", ErrorInvalidRegisterType, registerType)
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
			response, err = sendRegisterRequest(apiUrl)
		})

		wg.Wait()
	} else {
		response, err = sendRegisterRequest(apiUrl)
	}

	if err != nil {
		log.Errorf("Request register api %s response error: %v", apiInfo.ApiName, err)
		registerInfo.RegisterStatus = constant.RegisterStatusError
		if response != "" {
			registerInfo.RawResponse = response
		} else {
			registerInfo.RawResponse = err.Error()
		}
		return registerInfo, err
	}

	registerInfo.RawResponse = response

	log.Debugf("Request register api %s with domain %s, response: %s", apiInfo.ApiName, domain, response)

	if strutil.ContainsAny(response, apiInfo.SuccessText) && strutil.ContainsAny(response, apiInfo.FailText) {
		log.Errorf("Request register api %s with domain %s, response both contains success and fail text", apiInfo.ApiName, domain)
		registerInfo.RegisterStatus = constant.RegisterStatusError
		return registerInfo, fmt.Errorf("%w: %s", ErrorCustomizeApiRegisterResult, domain)
	} else if strutil.ContainsAny(response, apiInfo.SuccessText) {
		log.Debugf("Request register api %s with domain %s, response contains all success text", apiInfo.ApiName, domain)
		registerInfo.RegisterStatus = constant.RegisterStatusSuccess
	} else if strutil.ContainsAny(response, apiInfo.FailText) {
		log.Debugf("Request register api %s with domain %s, response contains all fail text", apiInfo.ApiName, domain)
		registerInfo.RegisterStatus = constant.RegisterStatusFailed
	} else {
		log.Errorf("Request register api %s with domain %s, response not contains success or fail text", apiInfo.ApiName, domain)
		registerInfo.RegisterStatus = constant.RegisterStatusError
		// registerInfo.RawResponse = fmt.Sprintf("Request register api %s with domain %s, response not contains success or fail text", apiInfo.ApiName, domain)
		return registerInfo, fmt.Errorf("%w: %s", ErrorCustomizeApiRegisterResult, domain)
	}

	return registerInfo, nil
}

func sendRegisterRequest(apiUrl string) (string, error) {
	client := resty.New()
	client.SetTimeout(defaultRegisterTimeout)

	response, err := client.R().Get(apiUrl)
	if err != nil {
		return "", err
	}

	return response.String(), nil
}
