package historylog

import "gorm.io/gorm"

// MessageOption is a QOption that injects a historylog message into the DB context.
// It implements repository.QOption via PrepareQuery.
type MessageOption struct {
	Msg string
}

// PrepareQuery adds the message to the gorm DB's context so historylog middleware
// can record it.
func (m *MessageOption) PrepareQuery(q *gorm.DB) *gorm.DB {
	if m.Msg == "" {
		return q
	}
	return q.WithContext(WithMessage(q.Statement.Context, m.Msg))
}

// Message returns a MessageOption carrying the given historylog message string.
func Message(msg string) *MessageOption {
	return &MessageOption{Msg: msg}
}
