package controllers

import (
	"errors"
	"fmt"
	"mime"
	nethttp "net/http"
	"strconv"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	frameworkerrors "github.com/goravel/framework/errors"

	"github.com/agiy/go_backend_sample/app/facades"
	"github.com/agiy/go_backend_sample/app/http/requests"
	"github.com/agiy/go_backend_sample/app/http/resources"
	"github.com/agiy/go_backend_sample/app/models"
)

var errInvalidTaskID = errors.New("IDは正の整数を指定してください。")

// TaskController はTaskのCRUDをHTTPへ公開する。
type TaskController struct{}

// NewTaskController はTask controllerを作成する。
func NewTaskController() *TaskController {
	return new(TaskController)
}

// Index はID昇順でTaskを取得する。取得件数は最大100件に制限する。
func (r *TaskController) Index(ctx http.Context) http.Response {
	page, err := taskPageParameter(ctx, "page", 1, 1_000_000)
	if err != nil {
		return ctx.Response().Json(nethttp.StatusBadRequest, http.Json{"message": err.Error()})
	}
	perPage, err := taskPageParameter(ctx, "per_page", 20, 100)
	if err != nil {
		return ctx.Response().Json(nethttp.StatusBadRequest, http.Json{"message": err.Error()})
	}

	tasks := make([]models.Task, 0)
	var total int64
	if err := facades.Orm().WithContext(ctx.Context()).Query().Model(&models.Task{}).
		OrderBy("id").Paginate(page, perPage, &tasks, &total); err != nil {
		return taskFailure(ctx, err)
	}

	return ctx.Response().Json(nethttp.StatusOK, resources.NewTaskCollection(tasks, page, perPage, total))
}

// Show は指定したIDのTaskを取得する。
func (r *TaskController) Show(ctx http.Context) http.Response {
	id, err := taskID(ctx)
	if err != nil {
		return taskFailure(ctx, err)
	}
	var task models.Task
	if err := facades.Orm().WithContext(ctx.Context()).Query().FindOrFail(&task, id); err != nil {
		return taskFailure(ctx, err)
	}

	return ctx.Response().Json(nethttp.StatusOK, resources.NewTaskResource(task))
}

// Store は入力検証後にTaskを作成する。
func (r *TaskController) Store(ctx http.Context) http.Response {
	var input requests.TaskRequest
	if response := validateTask(ctx, &input); response != nil {
		return response
	}
	task := models.Task{Title: input.Title, Description: input.Description, Completed: input.Completed}
	if err := facades.Orm().WithContext(ctx.Context()).Query().Create(&task); err != nil {
		return taskFailure(ctx, err)
	}

	return ctx.Response().Header("Location", fmt.Sprintf("/api/tasks/%d", task.ID)).
		Json(nethttp.StatusCreated, resources.NewTaskResource(task))
}

// Update はPUTの入力でTaskを置き換える。省略した説明・完了状態は既定値に戻す。
func (r *TaskController) Update(ctx http.Context) http.Response {
	id, err := taskID(ctx)
	if err != nil {
		return taskFailure(ctx, err)
	}
	var input requests.TaskRequest
	if response := validateTask(ctx, &input); response != nil {
		return response
	}

	var task models.Task
	if err := facades.Orm().WithContext(ctx.Context()).Transaction(func(tx orm.Query) error {
		// 存在確認から更新まで行をロックし、並行削除後にSaveが再作成することを防ぐ。
		if err := tx.LockForUpdate().FindOrFail(&task, id); err != nil {
			return fmt.Errorf("更新対象のTask取得: %w", err)
		}
		task.Title = input.Title
		task.Description = input.Description
		task.Completed = input.Completed
		// Saveはzero valueも保存するため、falseや空文字への変更が欠落しない。
		if err := tx.Save(&task); err != nil {
			return fmt.Errorf("作業項目の更新: %w", err)
		}

		return nil
	}); err != nil {
		return taskFailure(ctx, err)
	}

	return ctx.Response().Json(nethttp.StatusOK, resources.NewTaskResource(task))
}

// Destroy は指定したTaskを物理削除する。
func (r *TaskController) Destroy(ctx http.Context) http.Response {
	id, err := taskID(ctx)
	if err != nil {
		return taskFailure(ctx, err)
	}
	result, err := facades.Orm().WithContext(ctx.Context()).Query().Where("id", id).Delete(&models.Task{})
	if err != nil {
		return taskFailure(ctx, err)
	}
	if result.RowsAffected == 0 {
		return taskFailure(ctx, frameworkerrors.OrmRecordNotFound)
	}

	return ctx.Response().NoContent()
}

func validateTask(ctx http.Context, input *requests.TaskRequest) http.Response {
	mediaType, _, err := mime.ParseMediaType(ctx.Request().Header("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return ctx.Response().Json(nethttp.StatusUnsupportedMediaType, http.Json{"message": "Content-Typeはapplication/jsonを指定してください。"})
	}
	validationErrors, err := ctx.Request().ValidateRequest(input)
	if err != nil {
		return taskFailure(ctx, err)
	}
	if validationErrors != nil {
		return ctx.Response().Json(nethttp.StatusUnprocessableEntity, http.Json{
			"message": "入力内容を確認してください。", "errors": validationErrors.All(),
		})
	}

	return nil
}

func taskID(ctx http.Context) (int64, error) {
	id, err := strconv.ParseInt(ctx.Request().Route("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidTaskID
	}

	return id, nil
}

func taskPageParameter(ctx http.Context, name string, fallback, maximum int) (int, error) {
	value, err := strconv.Atoi(ctx.Request().Query(name, strconv.Itoa(fallback)))
	if err != nil || value < 1 || value > maximum {
		return 0, fmt.Errorf("%sは1から%dの整数を指定してください", name, maximum)
	}

	return value, nil
}

func taskFailure(ctx http.Context, err error) http.Response {
	if errors.Is(err, errInvalidTaskID) {
		return ctx.Response().Json(nethttp.StatusBadRequest, http.Json{"message": errInvalidTaskID.Error()})
	}
	if errors.Is(err, frameworkerrors.OrmRecordNotFound) {
		return ctx.Response().Json(nethttp.StatusNotFound, http.Json{"message": "Taskが見つかりません。"})
	}
	if inputError, ok := errors.AsType[*requests.InputError](err); ok {
		if inputError.Field == "" {
			return ctx.Response().Json(nethttp.StatusBadRequest, http.Json{"message": inputError.Message})
		}
		return ctx.Response().Json(nethttp.StatusUnprocessableEntity, http.Json{
			"message": "入力内容を確認してください。",
			"errors":  http.Json{inputError.Field: http.Json{"type": inputError.Message}},
		})
	}

	facades.Log().WithContext(ctx.Context()).With(map[string]any{"error": err}).Error("Task APIのDB操作に失敗しました。")
	return ctx.Response().Json(nethttp.StatusInternalServerError, http.Json{"message": "処理に失敗しました。"})
}
