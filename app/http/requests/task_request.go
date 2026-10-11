// Package requests はHTTP入力の型・正規化・検証規則を定義する。
package requests

import (
	"encoding/json/v2"
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// TaskRequest は作成とPUTによる置換で受け付ける項目だけを持つ。
// IDやtimestampをModelへ直接bindさせず、クライアントによる上書きを防ぐ。
type TaskRequest struct {
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
	Completed   bool   `form:"completed" json:"completed"`
}

// InputError はJSONの構文・型エラーをクライアントへ安全に伝える。
type InputError struct {
	Field   string
	Message string
}

// Error は入力エラーの説明を返す。
func (e *InputError) Error() string {
	return e.Message
}

// Authorize はローカル動作確認用の公開APIとして入力を許可する。
func (r *TaskRequest) Authorize(_ http.Context) error {
	return nil
}

// Rules はUnicodeの文字数でタイトルと説明の上限を検証する。
func (r *TaskRequest) Rules(_ http.Context) map[string]any {
	return map[string]any{
		"title":       "required|string|max:200",
		"description": "string|max:2000",
		"completed":   "boolean",
	}
}

// Messages は検証エラーを日本語で返す。
func (r *TaskRequest) Messages(_ http.Context) map[string]string {
	return map[string]string{
		"title.required":  "タイトルを入力してください。",
		"title.max":       "タイトルは200文字以内にしてください。",
		"description.max": "説明は2000文字以内にしてください。",
	}
}

// PrepareForValidation はJSON bodyだけを入力にし、型を確認して既定値を補う。
func (r *TaskRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// Goravelはbody・query・routeを統合するため、書き込み項目をbodyの値で置き換える。
	// JSON booleanを厳密に扱い、文字列の"false"や数値の0へ暗黙に変換しない。
	var body map[string]any
	if err := json.UnmarshalRead(ctx.Request().Origin().Body, &body); err != nil || body == nil {
		return &InputError{Message: "正しいJSON objectを指定してください。"}
	}

	for _, field := range []string{"title", "description", "completed"} {
		value, exists := body[field]
		value, err := taskInputValue(field, value, exists)
		if err != nil {
			return err
		}
		if err := data.Set(field, value); err != nil {
			return fmt.Errorf("入力したTaskの設定: %w", err)
		}
	}

	return nil
}

func taskInputValue(field string, value any, exists bool) (any, error) {
	if field == "completed" {
		if !exists {
			return false, nil
		}
		completed, ok := value.(bool)
		if !ok {
			return nil, &InputError{Field: field, Message: "trueまたはfalseを指定してください。"}
		}
		return completed, nil
	}
	if !exists {
		return "", nil
	}
	text, ok := value.(string)
	if !ok || strings.ContainsRune(text, '\x00') {
		return nil, &InputError{Field: field, Message: "NUL文字を含まない文字列を指定してください。"}
	}
	if field == "title" {
		text = strings.TrimSpace(text)
	}

	return text, nil
}
