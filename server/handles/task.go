package handles

import (
	"math"

	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/xhofe/tache"
)

type TaskInfo struct {
	ID       string      `json:"id"`
	UserID   uint        `json:"user_id"`
	Name     string      `json:"name"`
	State    tache.State `json:"state"`
	Status   string      `json:"status"`
	Progress float64     `json:"progress"`
	Error    string      `json:"error"`
}

func getTaskInfo[T fs.TaskWithInfo](task T) TaskInfo {
	errMsg := ""
	if task.GetErr() != nil {
		errMsg = task.GetErr().Error()
	}
	progress := task.GetProgress()
	// if progress is NaN, set it to 100
	if math.IsNaN(progress) {
		progress = 100
	}
	return TaskInfo{
		ID:       task.GetID(),
		UserID:   task.GetUserID(),
		Name:     task.GetName(),
		State:    task.GetState(),
		Status:   task.GetStatus(),
		Progress: progress,
		Error:    errMsg,
	}
}

func getTaskInfos[T fs.TaskWithInfo](tasks []T) []TaskInfo {
	return utils.MustSliceConvert(tasks, getTaskInfo[T])
}

func filter_user_tasks[T fs.TaskInterface](c *gin.Context, tasks []T) []T {
	user := c.MustGet("user").(*model.User)
	if user.IsGeneral() {
		ftasks := make([]T, 0, len(tasks))
		for _, task := range tasks {
			if task.GetUserID() == user.ID {
				ftasks = append(ftasks, task)
			}
		}
		return ftasks
	} else if user.IsAdmin() {
		return tasks
	} else {
		return make([]T, 0)
	}
}

func _GetByState[T fs.TaskInterface](c *gin.Context, manager *tache.Manager[T], state ...tache.State) []T {
	tasks := manager.GetByState(state...)
	return filter_user_tasks(c, tasks)
}

func _RemoveByState[T fs.TaskInterface](c *gin.Context, manager *tache.Manager[T], state ...tache.State) {
	tasks := _GetByState(c, manager, state...)
	for _, task := range tasks {
		manager.Remove(task.GetID())
	}
}

func _RetryAllFailed[T fs.TaskInterface](c *gin.Context, manager *tache.Manager[T]) {
	tasks := _GetByState(c, manager, tache.StateFailed)
	for _, task := range tasks {
		manager.Retry(task.GetID())
	}
}

func taskRoute[T fs.TaskWithInfo](g *gin.RouterGroup, manager *tache.Manager[T]) {
	g.GET("/undone", func(c *gin.Context) {
		common.SuccessResp(c, getTaskInfos(_GetByState(c, manager, tache.StatePending, tache.StateRunning,
			tache.StateCanceling, tache.StateErrored, tache.StateFailing, tache.StateWaitingRetry, tache.StateBeforeRetry)))
	})
	g.GET("/done", func(c *gin.Context) {
		common.SuccessResp(c, getTaskInfos(_GetByState(c, manager, tache.StateCanceled, tache.StateFailed, tache.StateSucceeded)))
	})
	g.POST("/info", func(c *gin.Context) {
		tid := c.Query("tid")
		task, ok := manager.GetByID(tid)
		if !ok {
			common.ErrorStrResp(c, "task not found", 404)
			return
		}
		common.SuccessResp(c, getTaskInfo(task))
	})
	g.POST("/cancel", func(c *gin.Context) {
		tid := c.Query("tid")
		manager.Cancel(tid)
		common.SuccessResp(c)
	})
	g.POST("/delete", func(c *gin.Context) {
		tid := c.Query("tid")
		manager.Remove(tid)
		common.SuccessResp(c)
	})
	g.POST("/retry", func(c *gin.Context) {
		tid := c.Query("tid")
		manager.Retry(tid)
		common.SuccessResp(c)
	})

	g.POST("/clear_done", func(c *gin.Context) {
		_RemoveByState(c, manager, tache.StateCanceled, tache.StateFailed, tache.StateSucceeded)
		//manager.RemoveByState(tache.StateCanceled, tache.StateFailed, tache.StateSucceeded)
		common.SuccessResp(c)
	})
	g.POST("/clear_succeeded", func(c *gin.Context) {
		_RemoveByState(c, manager, tache.StateSucceeded)
		//manager.RemoveByState(tache.StateSucceeded)
		common.SuccessResp(c)
	})
	g.POST("/retry_failed", func(c *gin.Context) {
		_RetryAllFailed(c, manager)
		//manager.RetryAllFailed()
		common.SuccessResp(c)
	})
}

func SetupTaskRoute(g *gin.RouterGroup) {
	taskRoute(g.Group("/upload"), fs.UploadTaskManager)
	taskRoute(g.Group("/copy"), fs.CopyTaskManager)
	taskRoute(g.Group("/offline_download"), tool.DownloadTaskManager)
	taskRoute(g.Group("/offline_download_transfer"), tool.TransferTaskManager)
}
