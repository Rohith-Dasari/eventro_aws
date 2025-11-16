package models

type EventUpdate struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Duration    *string        `json:"duration,omitempty"`
	Category    *EventCategory `json:"category,omitempty"`
	IsBlocked   *bool          `json:"isBlocked,omitempty"`
}
