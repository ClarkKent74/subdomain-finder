package scanner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type ZoneTransferScanner struct {
	timeout time.Duration
}

func NewZoneTransferScanner(timeout time.Duration) *ZoneTransferScanner {
	return &ZoneTransferScanner{timeout: timeout}
}

func (s *ZoneTransferScanner) Scan(ctx context.Context, domain string) ([]string, error) {
	nameservers, err := s.getNS(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения NS записей: %w", err)
	}
	if len(nameservers) == 0 {
		return nil, fmt.Errorf("NS записи для домена %s не найдены", domain)
	}

	var lastErr error
	for _, ns := range nameservers {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		subdomains, err := s.axfr(domain, ns)
		if err != nil {
			lastErr = err
			continue
		}
		return deduplicate(subdomains), nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("все NS серверы отклонили zone transfer, последняя ошибка: %w", lastErr)
	}

	return nil, fmt.Errorf("все NS серверы отклонили zone transfer")
}

func (s *ZoneTransferScanner) getNS(ctx context.Context, domain string) ([]string, error) {
	c := &dns.Client{Timeout: s.timeout}

	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeNS)

	resp, _, err := c.ExchangeContext(ctx, msg, "8.8.8.8:53")
	if err != nil {
		return nil, fmt.Errorf("DNS запрос NS записей завершился ошибкой: %w", err)
	}

	var nameservers []string
	for _, rr := range resp.Answer {
		if ns, ok := rr.(*dns.NS); ok {
			nameservers = append(nameservers, strings.TrimSuffix(ns.Ns, "."))
		}
	}

	return nameservers, nil
}

func (s *ZoneTransferScanner) axfr(domain, nameserver string) ([]string, error) {
	t := &dns.Transfer{
		DialTimeout:  s.timeout,
		ReadTimeout:  s.timeout,
		WriteTimeout: s.timeout,
	}

	msg := &dns.Msg{}
	msg.SetAxfr(dns.Fqdn(domain))

	channel, err := t.In(msg, nameserver+":53")
	if err != nil {
		return nil, fmt.Errorf("ошибка AXFR запроса к %s: %w", nameserver, err)
	}

	var subdomains []string
	suffix := "." + domain

	for envelope := range channel {
		if envelope.Error != nil {
			return nil, fmt.Errorf("ошибка получения зоны от %s: %w", nameserver, envelope.Error)
		}

		for _, rr := range envelope.RR {
			name := strings.TrimSuffix(rr.Header().Name, ".")
			name = strings.ToLower(name)

			if strings.HasSuffix(name, suffix) && name != domain {
				subdomains = append(subdomains, name)
			}
		}
	}

	return subdomains, nil
}
