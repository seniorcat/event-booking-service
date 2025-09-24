package container

import (
	"fmt"
	"sync"

	"github.com/sarulabs/di/v2"
)

var (
	mu       sync.Mutex
	ctn      di.Container
	built    bool
	builders []buildFn
)

type (
	Def       = di.Def
	Key       = di.ContainerKey
	Builder   = di.Builder
	Container = di.Container

	buildFn func(builder *Builder, params map[string]interface{}) error
)

const (
	App        = di.App
	Request    = di.Request
	SubRequest = di.SubRequest
)

// Register регистрирует функцию-билдер, которая добавит зависимости в контейнер.
// Можно вызывать из разных пакетов при инициализации.
func Register(fn buildFn) {
	mu.Lock()
	defer mu.Unlock()
	builders = append(builders, fn)
}

// Instance возвращает singleton-контейнер.
// Если контейнер ещё не создан — он будет собран из зарегистрированных билдов.
// Повторные вызовы возвращают уже готовый контейнер.
func Instance(scopes []string, params map[string]interface{}) (di.Container, error) {
	mu.Lock()
	defer mu.Unlock()

	// если контейнер уже инициализирован — возвращаем его
	if built {
		return ctn, nil
	}

	// создаём билдер контейнера
	builder, err := di.NewBuilder(scopes...)
	if err != nil {
		var zero di.Container
		return zero, fmt.Errorf("не удалось создать билдер контейнера: %w", err)
	}

	// применяем все зарегистрированные функции-билдеры
	for _, fn := range builders {
		if err := fn(builder, params); err != nil {
			var zero di.Container
			return zero, fmt.Errorf("ошибка при применении билдера: %w", err)
		}
	}

	// сохраняем собранный контейнер как singleton
	ctn = builder.Build()
	built = true
	return ctn, nil
}
