package main

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

func (ct *CommandTool) Complete(d prompt.Document) []prompt.Suggest {

	args := strings.Split(d.TextBeforeCursor(), " ")
	argsMap := make(map[string]interface{})

	for _, a := range args {
		argsMap[a] = nil
	}
	w := d.GetWordBeforeCursor()

	// If PIPE is in text before the cursor, returns empty suggestions.
	for i := range args {
		if args[i] == "|" {
			return []prompt.Suggest{}
		}
	}

	validOptions := make([]prompt.Suggest, 0)

	if len(args) <= 1 {
		validOptions = append(validOptions, prompt.FilterHasPrefix([]prompt.Suggest{
			{"logs", "view service logs"},
			{"storage", "query/modify starket storage"},
		}, w, false)...)
	}

	switch args[0] {
	case "storage":
		validOptions = append(validOptions, storage(args[1:], w)...)
	case "logs":
		validOptions = append(validOptions, logsAutocomplete(args[1:], w)...)
	}
	return validOptions
}

func storage(args []string, current string) []prompt.Suggest {
	if len(args) <= 1 {
		return prompt.FilterHasPrefix([]prompt.Suggest{
			{"reset", "reset ddb to a fresh state"},
			{"seed", "reset and seed the database"},
		}, current, false)
	}
	return make([]prompt.Suggest, 0)
}

func logsAutocomplete(args []string, current string) []prompt.Suggest {
	if len(args) <= 1 {
		return prompt.FilterHasPrefix([]prompt.Suggest{
			{"gql-lambda", "gql lambda logs"},
		}, current, false)
	}

	if len(args) > 1 && len(args)%2 == 1 {
		return prompt.FilterHasPrefix(
			[]prompt.Suggest{
				{"--start", "when to start (absolute or relative) (default -5m)"},
				{"--end", "when to start (absolute or relative)"},
				{"--filter", "filter logs"},
			}, current, false)
	}
	return make([]prompt.Suggest, 0)
}

func deleteAutocomplete(args []string, current string) []prompt.Suggest {
	if len(args) <= 1 {
		return prompt.FilterHasPrefix([]prompt.Suggest{
			{"get-outstanding", "get the outstanding deletion requests for a service id"},
			{"get-userdao-status", "get the deletion status for a userdao-id"},
			{"create-deletion", "create a deletion request"},
		}, current, false)
	}
	switch args[0] {
	case "get-outstanding":
		return []prompt.Suggest{{"serviceID", "the service id to get outstanding requests"}}
	case "get-userdao-status":
		return []prompt.Suggest{{"userID", "the userdao id to get the status for"}}
	case "create-deletion":
		switch len(args[1:]) {
		case 0:
			return []prompt.Suggest{{"", "userID"}}
		case 1:
			return []prompt.Suggest{{"", "login"}}
		case 2:
			return []prompt.Suggest{{"", "email"}}
		}
	}
	return make([]prompt.Suggest, 0)
}
