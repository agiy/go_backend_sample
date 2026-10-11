// Package resources はModelから公開するHTTPレスポンスを組み立てる。
package resources

import (
	"time"

	"github.com/agiy/go_backend_sample/app/models"
)

// TaskResource はTask APIが公開する項目を明示する。
// Modelを埋め込まず、DB項目の追加がそのままAPIの公開項目へ混ざることを防ぐ。
type TaskResource struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskCollection はTask一覧とpagination情報を返す。
type TaskCollection struct {
	Data []TaskResource `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta は一覧取得の位置・件数・総件数を示す。
type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

// NewTaskResource は保存済みModelを公開用の型へ変換する。
func NewTaskResource(task models.Task) TaskResource {
	return TaskResource{
		ID: task.ID, Title: task.Title, Description: task.Description, Completed: task.Completed,
		// carbonのDB向け表現を境界で変換し、APIの日時をUTCのRFC 3339に揃える。
		CreatedAt: task.CreatedAt.StdTime().UTC(),
		UpdatedAt: task.UpdatedAt.StdTime().UTC(),
	}
}

// NewTaskCollection は空の一覧もnullではなく配列で返す。
func NewTaskCollection(tasks []models.Task, page, perPage int, total int64) TaskCollection {
	data := make([]TaskResource, 0, len(tasks))
	for _, task := range tasks {
		data = append(data, NewTaskResource(task))
	}

	return TaskCollection{
		Data: data,
		Meta: PaginationMeta{Page: page, PerPage: perPage, Total: total},
	}
}
