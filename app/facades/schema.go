package facades

import (
	"github.com/goravel/framework/contracts/database/schema"
)

// Schema はmigration用のschema builderを返す。
func Schema() schema.Schema {
	return App().MakeSchema()
}
