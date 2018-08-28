package gomatrix

// Room represents a single Matrix room.
type Room struct {
	ID    string
	State map[EventType]map[string]*Event
}

// UpdateState updates the room's current state with the given Event. This will clobber events based
// on the type/state_key combination.
func (room Room) UpdateState(event *Event) {
	_, exists := room.State[event.Type]
	if !exists {
		room.State[event.Type] = make(map[string]*Event)
	}
	room.State[event.Type][*event.StateKey] = event
}

// GetStateEvent returns the state event for the given type/state_key combo, or nil.
func (room Room) GetStateEvent(eventType EventType, stateKey string) *Event {
	stateEventMap, _ := room.State[eventType]
	event, _ := stateEventMap[stateKey]
	return event
}

// GetMembershipState returns the membership state of the given user ID in this room. If there is
// no entry for this member, 'leave' is returned for consistency with left users.
func (room Room) GetMembershipState(userID string) string {
	state := "leave"
	event := room.GetStateEvent(StateMember, userID)
	if event != nil {
		state = event.Content.Membership
	}
	return state
}

// NewRoom creates a new Room with the given ID
func NewRoom(roomID string) *Room {
	// Init the State map and return a pointer to the Room
	return &Room{
		ID:    roomID,
		State: make(map[EventType]map[string]*Event),
	}
}