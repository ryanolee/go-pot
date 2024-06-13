package stall

type Staller interface {
	// Binds a given staller instance to the staller pool so that it can be shut down in the event
	// there are too many active connections and it can deregister itself when it is done
	BindToPool(chan Staller)

	// Shuts down the staller instance and cleans up any resources
	Close()

	// Gets the group identifier for the staller
	GetGroupIdentifier() string

	// Gets the identifier for the staller
	GetIdentifier() uint64
}