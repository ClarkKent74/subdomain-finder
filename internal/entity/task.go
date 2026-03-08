package entity

import "time"

type Status string

const (
	StatusPending Status = "pending" // задача создана, ожидает свободного воркера
	StatusRunning Status = "running" // воркер взял задачу и выполняет сканирование
	StatusDone    Status = "done"    // сканирование завершено успешно
	StatusFailed  Status = "failed"  // сканирование завершено с ошибкой
)

// IsActive возвращает true если задача ещё не завершена.
func (s Status) IsActive() bool {
	return s == StatusPending || s == StatusRunning
}

// Algorithm — алгоритм обнаружения поддоменов.
type Algorithm string

const (
	AlgorithmPassive Algorithm = "passive"

	AlgorithmBruteforce Algorithm = "bruteforce"

	AlgorithmZoneTransfer Algorithm = "zonetransfer"
)

type Task struct {
	Domain    string
	Status    Status
	Algorithm Algorithm

	Results []string
	Error   string

	CreatedAt time.Time
	UpdatedAt time.Time
}
