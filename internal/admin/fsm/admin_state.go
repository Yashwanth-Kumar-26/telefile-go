package admin

import "strconv"

// AdminState represents the current step in an admin multi-step flow.
// Empty string means no active admin flow.
type AdminState string

const (
	StateNone                AdminState = ""
	StateBroadcast           AdminState = "broadcast"
	StateLookupUser          AdminState = "lookup_user"
	StateSetStorage          AdminState = "set_storage"
	StateAddBonus            AdminState = "add_bonus"
	StateBanUser             AdminState = "ban_user"
	StateUnbanUser           AdminState = "unban_user"
	StateCreateGift          AdminState = "create_gift"
	StateCreateGiftExpiry    AdminState = "create_gift_expiry"
	StateCreateGiftClaims    AdminState = "create_gift_claims"
	StateCreatePlanName      AdminState = "create_plan_name"
	StateCreatePlanPrice     AdminState = "create_plan_price"
	StateCreatePlanStorage   AdminState = "create_plan_storage"
	StateCreatePlanDays      AdminState = "create_plan_days"
	StateCreatePlanDesc      AdminState = "create_plan_desc"
)

// AdminConversation holds the current FSM state and data for an admin.
type AdminConversation struct {
	State AdminState
	Data  map[string]interface{}
}

// AdminFSM manages admin FSM state per user.
type AdminFSM struct {
	// convs maps senderID -> conversation state
	convs map[int64]*AdminConversation
}

func NewAdminFSM() *AdminFSM {
	return &AdminFSM{convs: make(map[int64]*AdminConversation)}
}

// Get returns the conversation for a user, creating one if absent.
func (fsm *AdminFSM) Get(senderID int64) *AdminConversation {
	if c, ok := fsm.convs[senderID]; ok {
		return c
	}
	c := &AdminConversation{State: StateNone, Data: make(map[string]interface{})}
	fsm.convs[senderID] = c
	return c
}

// SetState transitions a user to a new state.
func (fsm *AdminFSM) SetState(senderID int64, state AdminState) {
	c := fsm.Get(senderID)
	c.State = state
	c.Data = make(map[string]interface{})
}

// SetData stores a value in the conversation.
func (fsm *AdminFSM) SetData(senderID int64, key string, value interface{}) {
	fsm.Get(senderID).Data[key] = value
}

// GetData retrieves a stored value.
func (fsm *AdminFSM) GetData(senderID int64, key string) interface{} {
	return fsm.Get(senderID).Data[key]
}

// GetIntData retrieves an int value with a default.
func (fsm *AdminFSM) GetIntData(senderID int64, key string) int64 {
	v := fsm.GetData(senderID, key)
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case string:
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// End clears the conversation for a user.
func (fsm *AdminFSM) End(senderID int64) {
	fsm.Get(senderID).State = StateNone
}

// HasState returns true if the user is in the given state.
func (fsm *AdminFSM) HasState(senderID int64, state AdminState) bool {
	return fsm.Get(senderID).State == state
}