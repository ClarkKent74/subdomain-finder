package entity

import "errors"

var (
	ErrTaskNotFound = errors.New("задача не найдена")

	ErrTaskAlreadyExists = errors.New("активная задача для этого домена уже существует")

	ErrQueueFull = errors.New("очередь задач переполнена, повторите позже")

	ErrInvalidDomain = errors.New("некорректный формат домена")

	ErrInvalidAlgorithm = errors.New("некорректный алгоритм, доступны: passive, bruteforce, zonetransfer")

	ErrZoneTransferDenied = errors.New("все NS-серверы отклонили zone transfer")

	ErrStoreFull = errors.New("хранилище задач заполнено, повторите позже")
)
