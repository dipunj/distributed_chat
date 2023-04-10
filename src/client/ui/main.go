package ui

import (
	"bufio"
	"chat/client/state"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const MESSAGE_CHAR_LIMIT = 82

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

var stdinReader = bufio.NewReader(os.Stdin)

func UserInputLoop() {

	for {
		os.Stdout.WriteString(getPrompt())
		rawInput, _ := stdinReader.ReadString('\n')
		input := strings.TrimSpace(rawInput)

		if len(input) > 0 {
			command, arg, err := parseUserInput(input)
			if err != nil {
				log.Error(err.Error())
			}
			commandController(command, arg)
		}
	}
}

func UserViewLoop() {
	for {
		if <-state.Rerender {
			state.RenderMu.Lock()
			RenderGroupDetails()
			state.RenderMu.Unlock()
		}
	}

}
