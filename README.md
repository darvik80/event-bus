Simple implementation of EventBus

```go
import event_bus "github.com/darvik80/event-bus"

type handler struct {
	
}

func (h handler) OnEvent(bus event_bus.EventBus, s string) {
    log.Infof("Handle event: %s", s)
}

...
bus := event_bus.NewEventBus()
bus.Subscribe(func(_ event_bus.EventBus, s string) {
    log.Infof("Handle event: %s", s)
})

h := &handler{}
bus.Subscribe(h)

...
bus.Fire("Hello World!")

```