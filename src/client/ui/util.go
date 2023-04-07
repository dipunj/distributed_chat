// contains utility functions that are called by the ui/main.go file
package ui

import (
	"chat/client/network"
	"chat/client/state"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getPrompt() string {
	var prompt string
	if state.Current_group != "" {
		prompt = "(" + state.Current_group + ")"
	}
	if state.Current_user != "" {
		prompt += "[" + state.Current_user + "]"
	}
	return "\n" + prompt + "> "
}

// parses whatever the user types
// returns command, argument, [error]
func parseUserInput(trimmedInput string) (string, string, error) {

	if trimmedInput == HELP || trimmedInput == PRINT_ALL || trimmedInput == QUIT || trimmedInput == CLEAR_SCREEN {
		return trimmedInput, "", nil
	}

	command := trimmedInput[0:1]

	if len(trimmedInput) < 3 {
		return "", "", errors.New("Command " + command + "takes an argument. Enter h for help")
	}

	space, arg := trimmedInput[1:2], trimmedInput[2:]

	// command and arg must be separated by a space
	if space != " " {
		return "", "", errors.New("invalid command")
	}

	switch command {
	case CONNECT:
		{
			return command, arg, nil
		}
	case SEND_MESSAGE:
		{
			if len(arg) > MESSAGE_CHAR_LIMIT {
				err_msg := "Message too long. Limit to" + strconv.Itoa(MESSAGE_CHAR_LIMIT) + "characters"
				return "", "", errors.New(err_msg)
			}
			return command, arg, nil
		}
	case LIKE, UNLIKE:
		{
			message_line_no, err := strconv.Atoi(arg)

			if err != nil {
				return "", "", errors.New("invalid message ID")
			}
			return command, strconv.Itoa(message_line_no), nil
		}
	case SWITCH_GROUP, SWITCH_USER:
		{
			return command, strings.ToLower(arg), nil
		}
	}

	return HELP, "", nil
}

func showHelp() {
	fmt.Println("Usage:")
	fmt.Println("\t h: help")
	fmt.Println("\t q: quits the program")
	fmt.Println("\t c <server_address>: connects to the server at <server_address> (e.g. localhost:8080)")
	fmt.Println("\t x: clear the screen (might not work on some terminals)")
	fmt.Println("\t p: shows the entire message history of the current group")
	fmt.Println("\t a <message>: sends the <message> to current group")
	fmt.Println("\t u <user_name>: switches the user to <user_name>")
	fmt.Println("\t j <group_name>: exits the current group, and join <group_name>")
	fmt.Println("\t l <message_id>: sends a like for <message_id>")
	fmt.Println("\t u <message_id>: removes the like for <message_id>")
}

func handleSwitchUser(new_user_name string) {
	oldUser := state.Current_user

	if oldUser == new_user_name {
		fmt.Println("Already logged in with username: ", oldUser)
		return
	}

	// make a network call only if there actually is a change
	if network.RequestUserSwitchTo(new_user_name) {
		state.RenderMu.Lock()
		state.Current_user = new_user_name
		state.Current_group = ""
		fmt.Print("Username switched. Current username: ", new_user_name)
		state.RenderMu.Unlock()
	} else {
		log.Error("Failed to switch user to: ", new_user_name)
	}
}

func handleSwitchGroup(new_group_name string) {
	oldGroup := state.Current_group

	if new_group_name == oldGroup {
		fmt.Println("Already in group", oldGroup)
		return
	}

	if state.Current_user == "" {
		fmt.Println("Please select a user name before joining a group")
		return
	}

	// make a network call only if there actually is a change
	if network.RequestGroupSwitchTo(new_group_name) {
		// clear the recent messages

		state.RenderMu.Lock()
		state.RecentMessages.ClearAll()
		state.Current_group = new_group_name

		if new_group_name != "" {
			// changing to some group
			fmt.Print("Group switched to: ", new_group_name)
		} else if oldGroup != "" {
			fmt.Print("Exiting group:", oldGroup)
		}
		state.RenderMu.Unlock()
	} else {
		log.Error("Failed to switch group, current group:", oldGroup)
	}

}

// reaction_name = "like" | "unlike"

func handleReaction(reaction_name, line_no string) {
	int_line_no, err := strconv.ParseInt(line_no, 10, 64)
	if err == nil {
		message := state.RecentMessages.GetFromLineNo(int(int_line_no))
		if message != nil {
			message_id := *message.Id
			sender := message.SenderName
			if sender == state.Current_user {
				fmt.Println("Can't", reaction_name, "own message")
			} else {
				if !network.SendReaction(reaction_name, message_id) {
					state.RenderMu.Lock()
					log.Info("Failed to", reaction_name, "message")
					state.RenderMu.Unlock()
				}
			}
		}
	}
}

func RenderGroupDetails() {
	// backup std in
	if state.Current_group == "" || state.Current_user == "" {
		return
	}

	messages := state.RecentMessages.ListAll()
	os.Stdout.WriteString("\n=======================================\n")
	os.Stdout.WriteString("Participants: " + strings.Join(state.Current_group_participants, ", ") + "\n")
	os.Stdout.WriteString("=======================================\n")
	os.Stdout.WriteString("Recent Messages:\n")
	if len(messages) > 0 {
		os.Stdout.WriteString("\n")
		for line_no, m := range state.RecentMessages.ListAll() {
			left := fmt.Sprintf("\t%d. [%s] %s", line_no+1, m.SenderName, m.Content)
			right := fmt.Sprintf("likes: %d", m.LikedBy)
			fmt.Printf("%s %s\n", left, right)
		}
	}
	os.Stdout.WriteString("=======================================\n")
	os.Stdout.WriteString("\n" + getPrompt())
}

func commandController(command, arg string) {

	// the following two commands can be executed even if not connected to server

	switch command {
	case CONNECT:
		{
			network.EstablishConnection(arg)
			return

		}
	case QUIT:
		{
			network.GracefulDisconnect("User quit the program")
			os.Exit(0)
		}
	case HELP:
		{
			showHelp()
			return
		}
	case CLEAR_SCREEN:
		{
			fmt.Print("\033[H\033[2J")
			return
		}
	}

	// to execute any other command, the user must be connected to server
	if command != CONNECT && !network.IsConnected() {
		log.Error("You are not connected to server. Please connect to a server first")
		showHelp()
		return
	}

	// the following commands can only be executed if connected to server
	switch command {
	case LIKE:
		{
			handleReaction("like", arg)
			return
		}
	case UNLIKE:
		{
			handleReaction("unlike", arg)
			return
		}
	case SWITCH_USER:
		{
			handleSwitchUser(arg)
			return
		}
	case SWITCH_GROUP:
		{
			handleSwitchGroup(arg)
			return
		}
	case SEND_MESSAGE:
		{
			network.SendMessage(arg)
			return
		}
	case PRINT_ALL:
		{
			// TODO
			network.PrintGroupHistory()
			return
		}
	}
}
