package database

var (
	AllowedEvents map[string]bool = map[string]bool{
		"like": true, "dislike": true, "playlist": true,
	}
	AllowedTypes map[string]bool = map[string]bool{
		"book": true, "tv": true, "movie": true,
	}
)

type UserEventPullRequest struct {
	Token     string `json:"access_token"`
	EventName string `json:"event"`
	ItemType  string `json:"type"`
}

type UserEventPushRequest struct {
	UserEventPullRequest
	ItemId string `json:"id"`
}

type Event struct {
	ItemId   uint64 `json:"id"`
	Name     string `json:"name"`
	ItemType string `json:"type"`
}

type UserEventPullResponse struct {
	Items []Event `json:"items"`
}

// ValidateFields checks if the fields might be checked, it doesn't guarantee
// that all the credentials will be valid.
func (u *UserEventPushRequest) ValidateFields() bool {
	// check if the token is passed
	if u.Token == "" || u.ItemId == "" {
		return false
	}
	if _, ok := AllowedEvents[u.EventName]; !ok {
		return false
	}
	if _, ok := AllowedTypes[u.ItemType]; !ok {
		return false
	}
	return true
}
