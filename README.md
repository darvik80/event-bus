Simple implementation of EventBus

```go
import event_bus "github.com/darvik80/event-bus"
...
bus := event_bus.NewEventBus()
bus.Subscribe(func(_ event_bus.EventBus, s string) {
log.Infof("Handle event: %s", s)
})
...
bus.Fire("Hello World!")

```