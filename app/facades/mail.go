package facades

import (
	"github.com/goravel/framework/contracts/mail"
)

// Mail はmailを作成・送信する入口を返す。
func Mail() mail.Mail {
	return App().MakeMail()
}
