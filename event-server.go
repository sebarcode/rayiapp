package rayiapp

/*
type EventHubFn func(evs *EventServer, btr byter.Byter) (kaos.EventHub, error)

var (
	mapEventHubFns = map[string]EventHubFn{}
)

func NewEventHub(evs *EventServer, btr byter.Byter) (kaos.EventHub, error) {
	fn, ok := mapEventHubFns[evs.ServerType]
	if !ok {
		return nil, fmt.Errorf("event hub %s is not exist", evs.ServerType)
	}
	return fn(evs, btr)
}

func RegisterEventHubFn(key string, fn EventHubFn) {
	mapEventHubFns[key] = fn
}
*/
