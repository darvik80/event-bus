package event_bus

import (
	"reflect"
	"runtime"
	"sync"
	"time"
)

type Handler interface {
}

type Event interface {
}

type Options struct {
	cacheSize int
	poolSize  int
}

type Option func(*Options)

func WithCacheSize(cacheSize int) Option {
	return func(cfg *Options) {
		cfg.cacheSize = cacheSize
	}
}

func WithPoolSize(poolSize int) Option {
	if poolSize <= 0 {
		poolSize = runtime.NumCPU()
	}
	return func(cfg *Options) {
		cfg.poolSize = poolSize
	}
}

type EventBus interface {
	Subscribe(h Handler) bool
	Fire(e Event)
	Send(e Event)
	Schedule(d time.Duration, rep bool, e Event)
	Shutdown()
}

func New(opts ...Option) EventBus {
	var options = &Options{
		cacheSize: 1,
		poolSize:  1,
	}

	for _, opt := range opts {
		opt(options)
	}

	bus := &eventBus{
		mutex: sync.Mutex{},
		bus:   make(chan Event, options.cacheSize),
	}

	for idx := 0; idx < options.poolSize; idx++ {
		go func() {
			for e := range bus.bus {
				bus.callHandlers(e)
			}
		}()
	}

	return bus
}

func (bus *eventBus) Shutdown() {
	close(bus.bus)
}

type eventHandlerMeta struct {
	handler  Handler
	callback reflect.Value
}

type eventBus struct {
	handlers []eventHandlerMeta
	mutex    sync.Mutex
	bus      chan Event
}

func validate(arg, origin reflect.Type) bool {
	return arg.Name() == origin.Name() && arg.PkgPath() == origin.PkgPath()
}

func (bus *eventBus) callHandlers(e Event) {
	bus.mutex.Lock()
	copyHandlers := make([]eventHandlerMeta, len(bus.handlers))
	copy(copyHandlers, bus.handlers)
	bus.mutex.Unlock()
	for _, h := range copyHandlers {
		bus.tryCall(e, h)
	}
}

func (bus *eventBus) tryCall(e Event, meta eventHandlerMeta) bool {
	if !validate(reflect.ValueOf(e).Type(), meta.callback.Type().In(1)) {
		return false
	}

	meta.callback.Call([]reflect.Value{reflect.ValueOf(bus), reflect.ValueOf(e)})

	return true
}

func (bus *eventBus) Subscribe(h Handler) bool {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	meta := eventHandlerMeta{
		handler: h,
	}

	val := reflect.ValueOf(h)

	// validate func or method
	if val.Kind() == reflect.Func {
		meta.callback = val
	} else {
		method := val.MethodByName("OnEvent")
		if false == method.IsValid() {
			return false
		}

		meta.callback = method
	}

	// validate must be 2 arguments
	if meta.callback.Type().NumIn() != 2 {
		return false
	}

	// validate first argument - must be EventBus
	if meta.callback.Type().In(0).Name() != "EventBus" {
		return false
	}

	bus.handlers = append(bus.handlers, meta)

	return true
}

func (bus *eventBus) Fire(e Event) {
	bus.bus <- e
}

func (bus *eventBus) Send(e Event) {
	bus.callHandlers(e)
}

func (bus *eventBus) Schedule(d time.Duration, rep bool, e Event) {
	go func() {
		if rep {
			ticker := time.NewTicker(d)
			for range ticker.C {
				bus.Fire(e)
			}
		} else {
			time.AfterFunc(d, func() {
				bus.Fire(e)
			})
		}
	}()
}
