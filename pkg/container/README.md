# Wrapper container

## Описание
Данный модуль предоставляет обёртку над контейнером внедрения зависимостей (DI).  
Внутри используется пакет [di/v2](https://github.com/sarulabs/di).

## Пример использования
```go
package main

import (
	"log"

	"gitlab.com/headliner/dev/go/skeleton/v1/modules/container"
)

type Some struct{}

func New() *Some {
	return &Some{}
}

const DIWrapper = "example"

// Регистрируем зависимость при инициализации
func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIWrapper,
			Build: func(ctn container.Container) (interface{}, error) {
				return New(), nil
			},
		})
	})
}

func main() {
	// Создаём контейнер (singleton)
	diContainer, err := container.Instance([]string{container.App}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Достаём зависимость
	var s *Some
	if err := diContainer.Fill(DIWrapper, &s); err != nil {
		log.Fatal(err)
	}

	log.Println(s)
}
```

## Пример конфигурации

У данного модуля отсутствует конфигурация через файл.

## Зависимости от модулей

Зависимости у данного модуля отсутствуют.