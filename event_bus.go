package event_bus

import (
	"reflect"
	"runtime"
	"sync"
	"time"
)

type Source interface {
}

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
	FireAnonymous(e Event)
	Fire(s Source, e Event)
	SendAnonymous(e Event)
	Send(s Source, e Event)
	Schedule(s Source, d time.Duration, rep bool, e Event)
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
		bus:   make(chan busItem, options.cacheSize),
	}

	for idx := 0; idx < options.poolSize; idx++ {
		go func() {
			for e := range bus.bus {
				bus.callHandlers(e.s, e.e)
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

type busItem struct {
	s Source
	e Event
}

type eventBus struct {
	handlers []eventHandlerMeta
	mutex    sync.Mutex
	bus      chan busItem
}

func validate(arg, origin reflect.Type) bool {
	return arg.Name() == origin.Name() && arg.PkgPath() == origin.PkgPath()
}

func (bus *eventBus) callHandlers(s Source, e Event) {
	bus.mutex.Lock()
	copyHandlers := make([]eventHandlerMeta, len(bus.handlers))
	copy(copyHandlers, bus.handlers)
	bus.mutex.Unlock()
	for _, h := range copyHandlers {
		bus.tryCall(s, e, h)
	}
}

func (bus *eventBus) tryCall(s Source, e Event, meta eventHandlerMeta) bool {
	if s == nil {
		s = bus
	}
	idx := meta.callback.Type().NumIn()
	if idx < 2 {
		return false
	}
	if !validate(reflect.ValueOf(e).Type(), meta.callback.Type().In(idx-1)) {
		return false
	}

	if meta.callback.Type().NumIn() == 2 {
		meta.callback.Call([]reflect.Value{reflect.ValueOf(bus), reflect.ValueOf(e)})
	} else {
		meta.callback.Call([]reflect.Value{reflect.ValueOf(bus), reflect.ValueOf(s), reflect.ValueOf(e)})
	}

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

	// validate must be 2,3 arguments
	numIn := meta.callback.Type().NumIn()
	if numIn < 2 || numIn > 3 {
		return false
	}

	// validate first argument - must be EventBus
	if meta.callback.Type().In(0).Name() != "EventBus" {
		return false
	}

	bus.handlers = append(bus.handlers, meta)

	return true
}

func (bus *eventBus) Fire(s Source, e Event) {
	bus.bus <- busItem{s, e}
}

func (bus *eventBus) Send(s Source, e Event) {
	bus.callHandlers(s, e)
}

func (bus *eventBus) FireAnonymous(e Event) {
	bus.Fire(nil, e)
}

func (bus *eventBus) SendAnonymous(e Event) {
	bus.Send(nil, e)
}

func (bus *eventBus) Schedule(s Source, d time.Duration, rep bool, e Event) {
	go func() {
		if rep {
			ticker := time.NewTicker(d)
			for range ticker.C {
				bus.Fire(s, e)
			}
		} else {
			time.AfterFunc(d, func() {
				bus.Fire(s, e)
			})
		}
	}()
}
