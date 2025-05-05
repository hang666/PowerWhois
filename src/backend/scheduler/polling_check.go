package scheduler

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"typonamer/config"
	"typonamer/constant"
	"typonamer/database"
	"typonamer/log"
	"typonamer/lookup/lookuper"
	"typonamer/lookup/lookuperror"
	"typonamer/lookup/lookupinfo"
	"typonamer/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key for all domain check results
	redisDomainKey = "domain_checks"
	// Maximum concurrent domain checks
	maxConcurrent = 100
	// Number of retries for webhook notification
	webhookRetryCount = 3
	// Interval between webhook retries
	webhookRetryInterval = 1 * time.Second
)

type DomainCheckResult struct {
	LastCheck            time.Time             `json:"last_check"`
	Result               lookupinfo.DomainInfo `json:"result"`
	FirstSuccessNotified bool                  `json:"first_success_notified"`
}

type PollingCheck struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var pollingCheck *PollingCheck

// StartPollingCheck starts the polling check task
func StartPollingCheck() {
	if pollingCheck != nil {
		log.Warn("Polling check already started")
		return
	}

	cfg := config.GetConfig()
	if !cfg.PollingCheck.Enabled {
		log.Info("Polling check is disabled")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	pollingCheck = &PollingCheck{
		cancel: cancel,
	}

	pollingCheck.wg.Add(1)
	go pollingCheck.run(ctx)

	log.Info("Polling check started")
}

// StopPollingCheck stops the polling check task
func StopPollingCheck() {
	if pollingCheck == nil {
		return
	}

	pollingCheck.cancel()
	pollingCheck.wg.Wait()
	pollingCheck = nil

	log.Info("Polling check stopped")
}

func (p *PollingCheck) run(ctx context.Context) {
	defer p.wg.Done()

	cfg := config.GetConfig()
	ticker := time.NewTicker(time.Duration(cfg.PollingCheck.Interval) * time.Second)
	defer ticker.Stop()

	// Get Redis client
	rdb, err := database.GetRedis()
	if err != nil {
		log.Error("Failed to connect to Redis:", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkDomains(ctx, rdb)
		}
	}
}

func (p *PollingCheck) checkDomains(ctx context.Context, rdb *redis.Client) {
	cfg := config.GetConfig()

	// Read domains from file
	domains, err := readDomainsFromFile(cfg.PollingCheck.FilePath)
	if err != nil {
		log.Error("Failed to read domains from file:", err)
		return
	}

	// Create channel for limiting concurrency
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// Load existing results from Redis
	var results map[string]DomainCheckResult
	resultsData, err := rdb.Get(ctx, redisDomainKey).Bytes()
	if err == redis.Nil {
		results = make(map[string]DomainCheckResult)
	} else if err != nil {
		log.Error("Failed to get domain results from Redis:", err)
		results = make(map[string]DomainCheckResult)
	} else {
		if err := json.Unmarshal(resultsData, &results); err != nil {
			log.Error("Failed to unmarshal domain results:", err)
			results = make(map[string]DomainCheckResult)
		}
	}

	// Use mutex to protect map access
	var resultsMutex sync.Mutex

	for _, domain := range domains {
		select {
		case <-ctx.Done():
			return
		default:
			wg.Add(1)
			go func(domain string) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire
				defer func() { <-sem }() // Release

				result, err := lookuper.Lookup(domain, cfg.PollingCheck.CheckType)
				queryResult, _ := p.CheckLookupResultHandler(domain, result, err)
				if queryResult.RegisterStatus == constant.DomainRegisterStatusError {
					log.Errorf("Failed to lookup domain %s: %v", domain, err)
					return
				}

				isFree := queryResult.RegisterStatus == constant.DomainRegisterStatusFree
				result = lookupinfo.DomainInfo{
					DomainName:       domain,
					LookupType:       queryResult.LookupType,
					ViaProxy:         queryResult.ViaProxy,
					DomainStatus:     queryResult.RawDomainStatus,
					CreationDate:     queryResult.CreatedDate,
					ExpiryDate:       queryResult.ExpiryDate,
					NameServer:       queryResult.NameServer,
					RawResponse:      queryResult.RawResponse,
					CustomizedResult: queryResult.RegisterStatus,
				}

				resultsMutex.Lock()
				oldResult, exists := results[domain]

				// Update results
				results[domain] = DomainCheckResult{
					LastCheck:            time.Now(),
					Result:               result,
					FirstSuccessNotified: oldResult.FirstSuccessNotified || !isFree,
				}

				// Check if this is the first success
				isFirstSuccess := isFree && (!exists || !oldResult.FirstSuccessNotified)
				resultsMutex.Unlock()

				// Send webhook notification for first success
				if isFirstSuccess {
					msg := fmt.Sprintf("%s 域名开放注册啦！", domain)
					p.sendWebhookNotification(
						cfg.PollingCheck.NotifyWebhook,
						cfg.PollingCheck.NotifyMethod,
						cfg.PollingCheck.NotifyBody,
						cfg.PollingCheck.NotifyHeaders,
						msg,
					)

					// Mark as notified
					resultsMutex.Lock()
					if r, ok := results[domain]; ok {
						r.FirstSuccessNotified = true
						results[domain] = r
					}
					resultsMutex.Unlock()
				}
			}(domain)
		}
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Save all results to Redis
	resultsData, err = json.Marshal(results)
	if err != nil {
		log.Error("Failed to marshal domain results:", err)
		return
	}

	err = rdb.Set(ctx, redisDomainKey, resultsData, 0).Err() // No expiration
	if err != nil {
		log.Error("Failed to save domain results to Redis:", err)
	}
}

func readDomainsFromFile(filePath string) ([]string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return domains, nil
}

func (p *PollingCheck) CheckLookupResultHandler(domain string, lookupResult lookupinfo.DomainInfo, lookupErr error) (lookupinfo.QueryResult, error) {
	switch lookupResult.LookupType {
	case constant.LookupTypeWhois, constant.LookupTypeRDAP:
		if lookupErr == nil {
			// If the whois query is successful, parse the result and save it to redis
			queryResult := lookupinfo.QueryResult{
				Domain:          domain,
				LookupType:      lookupResult.LookupType,
				ViaProxy:        lookupResult.ViaProxy,
				RegisterStatus:  constant.DomainRegisterStatusTaken,
				CreatedDate:     lookupResult.CreationDate,
				ExpiryDate:      lookupResult.ExpiryDate,
				NameServer:      slice.Map(lookupResult.NameServer, utils.LowerString),
				DnsLite:         utils.GetDnsLite(lookupResult.NameServer),
				RawDomainStatus: lookupResult.DomainStatus,
				DomainStatus:    utils.GetDomainHumanStatus(lookupResult.DomainStatus),
			}
			return queryResult, nil
		} else if errors.Is(lookupErr, lookuperror.ErrorWhoisNotFound) {
			// If the whois query is not found, set the register status to free
			queryResult := lookupinfo.QueryResult{
				Domain:         domain,
				LookupType:     lookupResult.LookupType,
				ViaProxy:       lookupResult.ViaProxy,
				RegisterStatus: constant.DomainRegisterStatusFree,
			}
			return queryResult, nil
		} else {
			queryResult := lookupinfo.QueryResult{
				Domain:         domain,
				LookupType:     lookupResult.LookupType,
				ViaProxy:       lookupResult.ViaProxy,
				RegisterStatus: constant.DomainRegisterStatusError,
				QueryError:     utils.GetDomainHumanError(lookupErr),
			}
			return queryResult, lookupErr
		}
	case constant.LookupTypeDNS:
		if lookupErr == nil {
			if len(lookupResult.NameServer) > 0 {
				takenResult := lookupinfo.QueryResult{
					Domain:         domain,
					LookupType:     lookupResult.LookupType,
					RegisterStatus: constant.DomainRegisterStatusTaken,
					NameServer:     slice.Map(lookupResult.NameServer, utils.LowerString),
					DnsLite:        utils.GetDnsLite(lookupResult.NameServer),
				}
				return takenResult, nil
			} else {
				freeResult := lookupinfo.QueryResult{
					Domain:         domain,
					LookupType:     lookupResult.LookupType,
					RegisterStatus: constant.DomainRegisterStatusFree,
				}
				return freeResult, nil
			}
		} else if errors.Is(lookupErr, lookuperror.ErrorNsNotFound) {
			freeResult := lookupinfo.QueryResult{
				Domain:         domain,
				LookupType:     lookupResult.LookupType,
				RegisterStatus: constant.DomainRegisterStatusFree,
			}
			return freeResult, nil
		} else {
			errorResult := lookupinfo.QueryResult{
				Domain:         domain,
				LookupType:     lookupResult.LookupType,
				RegisterStatus: constant.DomainRegisterStatusError,
				QueryError:     utils.GetDomainHumanError(lookupErr),
			}
			return errorResult, lookupErr
		}
	default:
		// Customize api whois result
		if lookupErr != nil {
			errorResult := lookupinfo.QueryResult{
				Domain:         domain,
				LookupType:     lookupResult.LookupType,
				RegisterStatus: constant.DomainRegisterStatusError,
				QueryError:     utils.GetDomainHumanError(lookupErr),
			}
			return errorResult, lookupErr
		} else {
			switch lookupResult.CustomizedResult {
			case constant.DomainRegisterStatusTaken:
				takenResult := lookupinfo.QueryResult{
					Domain:         domain,
					LookupType:     lookupResult.LookupType,
					RegisterStatus: constant.DomainRegisterStatusTaken,
				}
				return takenResult, nil
			case constant.DomainRegisterStatusFree:
				freeResult := lookupinfo.QueryResult{
					Domain:         domain,
					LookupType:     lookupResult.LookupType,
					RegisterStatus: constant.DomainRegisterStatusFree,
				}
				return freeResult, nil
			}
		}
	}
	return lookupinfo.QueryResult{}, nil
}

func (p *PollingCheck) sendWebhookNotification(webhookURL string, method string, body string, headers map[string]string, msg string) {
	if webhookURL == "" || method == "" {
		return
	}
	webhookURL = strings.ReplaceAll(webhookURL, "{msg}", msg)
	body = strings.ReplaceAll(body, "{msg}", msg)
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < webhookRetryCount; i++ {
		req, err := http.NewRequest(method, webhookURL, bytes.NewBuffer([]byte(body)))
		if err != nil {
			log.Errorf("Failed to create webhook request: %v", err)
			continue
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Errorf("Failed to send webhook notification: %v", err)
			time.Sleep(webhookRetryInterval)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Info("Webhook notification sent successfully")
			return
		} else {
			log.Errorf("Webhook notification failed with status code: %d", resp.StatusCode)
			time.Sleep(webhookRetryInterval)
		}
	}

	log.Errorf("Failed to send webhook notification after %d retries", webhookRetryCount)
}
