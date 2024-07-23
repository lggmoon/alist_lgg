package fs

import (
	"github.com/xhofe/tache"
)

type TaskInterface interface {
	tache.Task
	GetUserID() uint
	SetUserID(user_id uint)
}

type TaskWithInfo interface {
	TaskInterface
	tache.Info
}

type TaskData struct {
	tache.Base
	UserID uint `json:"user_id"`
}

func (t *TaskData) GetUserID() uint {
	return t.UserID
}

func (t *TaskData) SetUserID(user_id uint) {
	t.UserID = user_id
}
