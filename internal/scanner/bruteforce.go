package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// defaultWordlist — встроенный словарь, используется если пользователь не загрузил свой.
var defaultWordlist = []string{
	"www", "mail", "ftp", "smtp", "pop", "imap", "api", "dev", "staging",
	"test", "admin", "portal", "dashboard", "app", "mobile", "cdn", "static",
	"assets", "media", "img", "images", "video", "docs", "support", "help",
	"blog", "shop", "store", "pay", "payment", "secure", "login", "auth",
	"oauth", "sso", "git", "gitlab", "jenkins", "ci", "monitoring", "grafana",
	"prometheus", "kibana", "elasticsearch", "redis", "mysql", "postgres",
	"backup", "old", "new", "v1", "v2", "internal", "intranet", "corporate",
	"vpn", "remote", "mx", "ns1", "ns2", "autodiscover", "webmail", "cpanel",
	"whm", "plesk", "phpmyadmin", "beta", "alpha", "sandbox", "demo", "qa",
	"uat", "prod", "production", "stage", "preview", "status", "health",
}

type BruteforceScanner struct {
	resolver    *net.Resolver
	workerCount int
	timeout     time.Duration
}

// NewBruteforceScanner создаёт брутфорс сканер.
func NewBruteforceScanner(resolverAddr string, workerCount int, timeout time.Duration) *BruteforceScanner {
	return &BruteforceScanner{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: timeout}
				return d.DialContext(ctx, "udp", resolverAddr)
			},
		},
		workerCount: workerCount,
		timeout:     timeout,
	}
}

// Scan — использует встроенный словарь.
func (s *BruteforceScanner) Scan(ctx context.Context, domain string) ([]string, error) {
	return s.ScanWithWords(ctx, domain, defaultWordlist)
}

// ScanWithWords запускает брутфорс с переданным словарём.
func (s *BruteforceScanner) ScanWithWords(ctx context.Context, domain string, words []string) ([]string, error) {
	if len(words) == 0 {
		return nil, fmt.Errorf("словарь пустой")
	}
	return s.scan(ctx, domain, words)
}

// scan выполняет брутфорс с переданным словарём через worker pool.
func (s *BruteforceScanner) scan(ctx context.Context, domain string, words []string) ([]string, error) {
	wildcardIPs := s.detectWildcard(ctx, domain)

	jobs := make(chan string, len(words))
	results := make(chan string, len(words))

	var wg sync.WaitGroup
	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for prefix := range jobs {
				if ctx.Err() != nil {
					return
				}
				subdomain := fmt.Sprintf("%s.%s", prefix, domain)
				if s.resolve(ctx, subdomain, wildcardIPs) {
					results <- subdomain
				}
			}
		}()
	}

	for _, word := range words {
		jobs <- word
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var found []string
	for subdomain := range results {
		found = append(found, subdomain)
	}

	return found, nil
}

// resolve проверяет что поддомен резолвится и не является wildcard.
func (s *BruteforceScanner) resolve(ctx context.Context, subdomain string, wildcardIPs map[string]struct{}) bool {
	addrs, err := s.resolver.LookupHost(ctx, subdomain)
	if err != nil || len(addrs) == 0 {
		return false
	}
	if len(wildcardIPs) == 0 {
		return true
	}
	for _, addr := range addrs {
		if _, isWildcard := wildcardIPs[addr]; !isWildcard {
			return true
		}
	}
	return false
}

// detectWildcard проверяет наличие wildcard DNS.
func (s *BruteforceScanner) detectWildcard(ctx context.Context, domain string) map[string]struct{} {
	probe := fmt.Sprintf("wildcard-probe-zxqjvk.%s", domain)
	addrs, err := s.resolver.LookupHost(ctx, probe)
	if err != nil || len(addrs) == 0 {
		return nil
	}
	ips := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		ips[addr] = struct{}{}
	}
	return ips
}
