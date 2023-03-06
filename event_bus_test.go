package event_bus

import (
	"log"
	"testing"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Log("New EventBus not created!")
		t.Fail()
	}
}

type failHandler struct {
}

type successHandler struct {
	t *testing.T
}

func (h successHandler) OnEvent(bus EventBus, s string) {
	if s != "test" {
		h.t.Fail()
	}

}

type wrongHandler struct {
}

func (h wrongHandler) OnEvent(bus EventBus) {
}

type wrongEventHandler struct {
	t *testing.T
}

func (h wrongEventHandler) OnEvent(s EventBus, v int) {
	h.t.Fail()
}

type wrongEventArgsHandler struct {
	t *testing.T
}

func (h wrongEventArgsHandler) OnEvent(s string, v int) {
	h.t.Fail()
}

func TestSubscribe(t *testing.T) {
	bus := NewEventBus()
	if true == bus.Subscribe(func() {}) {
		log.Print("fail: callback no args")
		t.Fail()
	}
	if false == bus.Subscribe(func(b EventBus, s string) {}) {
		log.Print("fail: callback correct")
		t.Fail()
	}

	f := &failHandler{}
	if true == bus.Subscribe(f) {
		log.Print("fail: fail handler")
		t.Fail()
	}

	w := &wrongHandler{}
	if true == bus.Subscribe(w) {
		log.Print("fail: wrong handler")
		t.Fail()
	}

	a := &wrongEventArgsHandler{}
	if true == bus.Subscribe(a) {
		log.Print("fail: wrong arg handler")
		t.Fail()
	}

	s := &successHandler{}
	if false == bus.Subscribe(s) {
		log.Print("fail: handler correct")
		t.Fail()
	}
}

func TestFire(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe(func(b EventBus, s string) {
		if s != "test" {
			t.Fail()
		}
	})

	s := &successHandler{}
	if false == bus.Subscribe(s) {
		t.Fail()
	}

	w := &wrongEventHandler{}
	if false == bus.Subscribe(w) {
		t.Fail()
	}

	bus.Fire("test")
}
