package caption

import "errors"

const (
	CmdStart = "start"
	CmdHome  = "home"
	CmdStop  = "stop"
)

const (
	CmdTMainMenu = "🔝 Main Menu"
)

const (
	TextCmdStart       = "Simple images Telegram bot"
	TextUnknownMessage = "I don't understand you, please type correct image category!"
)

var (
	ErrUnknownCommand = errors.New("unknown bot command")
)
