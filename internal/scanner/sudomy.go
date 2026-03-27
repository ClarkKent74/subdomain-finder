package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type SudomyScanner struct {
	path              string
	timeout           time.Duration
	virusTotalKey     string
	shodanKey         string
	censysKey         string
	securityTrailsKey string
}

func NewSudomyScanner(
	path string,
	timeout time.Duration,
	virusTotalKey string,
	shodanKey string,
	censysKey string,
	securityTrailsKey string,
) *SudomyScanner {
	return &SudomyScanner{
		path:              path,
		timeout:           timeout,
		virusTotalKey:     virusTotalKey,
		shodanKey:         shodanKey,
		censysKey:         censysKey,
		securityTrailsKey: securityTrailsKey,
	}
}

func (s *SudomyScanner) Scan(ctx context.Context, domain string) ([]string, error) {
	if err := s.writeAPIFile(); err != nil {
		return nil, fmt.Errorf("ошибка записи API ключей: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("sudomy-%s-*", domain))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания временной директории: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scanCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	args := s.buildArgs(domain, tmpDir)

	cmd := exec.CommandContext(scanCtx, s.path, args...)
	cmd.Env = s.buildEnv()

	if err := cmd.Run(); err != nil {
		if scanCtx.Err() != nil {
			return nil, fmt.Errorf("сканирование Sudomy прервано: %w", scanCtx.Err())
		}
	}

	resultPath := filepath.Join(tmpDir, "Sudomy-Output", domain, "subdomain.txt")
	results, err := s.parseResult(resultPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения результатов Sudomy: %w", err)
	}

	return deduplicate(results), nil
}

func (s *SudomyScanner) buildArgs(domain, outputDir string) []string {
	return []string{
		"-d", domain,
		"-a",
		"--no-probe",
		"-o", outputDir,
	}
}

func (s *SudomyScanner) writeAPIFile() error {
	content := fmt.Sprintf(
		"SHODAN_API=\"%s\"\nCENSYS_API=\"%s\"\nCENSYS_SECRET=\"\"\nVIRUSTOTAL=\"%s\"\nBINARYEDGE=\"\"\nSECURITY_TRAILS=\"%s\"\n",
		s.shodanKey,
		s.censysKey,
		s.virusTotalKey,
		s.securityTrailsKey,
	)
	return os.WriteFile("/usr/lib/sudomy/sudomy.api", []byte(content), 0644)
}

func (s *SudomyScanner) buildEnv() []string {
	return os.Environ()
}

func (s *SudomyScanner) parseResult(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("не удалось открыть файл результатов: %w", err)
	}
	defer f.Close()

	var subdomains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.Contains(line, "?") {
			continue
		}
		subdomains = append(subdomains, line)
	}

	return subdomains, scanner.Err()
}

func deduplicate(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
