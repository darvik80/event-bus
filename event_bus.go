package event_bus

import (
	"reflect"
	"sync"
)

type Handler interface {
}

type Event interface {
}

type EventBus interface {
	Subscribe(h Handler) bool
	Fire(e Event)
}

func NewEventBus() EventBus {
	return &eventBus{
		mutex: sync.Mutex{},
	}
}

type eventHandlerMeta struct {
	handler  Handler
	callback reflect.Value
}

type eventBus struct {
	handlers []eventHandlerMeta
	mutex    sync.Mutex
}

func validate(arg, origin reflect.Type) bool {
	return arg.Name() == origin.Name() && arg.PkgPath() == origin.PkgPath()
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
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, h := range bus.handlers {
		bus.tryCall(e, h)
	}
}
