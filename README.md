Simple implementation of EventBus

```go
import event_bus "github.com/darvik80/event-bus"

type handler struct {
	
}

func (h handler) OnEvent(bus event_bus.EventBus, s string) {
    log.Infof("Handle event: %s", s)
}

...
bus := event_bus.New()
...
bus := event_bus.New(WithCacheSize(10), WitchPoolSize(0))
...
bus.Subscribe(func(_ event_bus.EventBus, s string) {
    log.Infof("Handle event: %s", s)
})

h := &handler{}
bus.Subscribe(h)

...
bus.Fire("Hello World!")

bus.Schedule(time.Second, false, "Hello World once!")

bus.Schedule(time.Second, true, "Hello World periodic!")

```