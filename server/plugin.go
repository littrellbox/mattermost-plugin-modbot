package main

import (
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

var muteduserids []string

const (
	trigger     string = "mod"
	displayname string = "System Admin"
)

func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		AutoComplete:     false,
		AutoCompleteDesc: "wip mod command.",
		DisplayName:      displayname,
	})

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	argumentArray := strings.Split(args.Command, " ")
	moderatorList := strings.Split(strings.TrimSpace(p.getConfiguration().Moderators), ",")
	var user *model.User
	var err *model.AppError
	user, err = p.API.GetUser(args.UserId)

	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "An error has occured. Please contact your adminstrator.",
		}, nil
	}

	if stringInSlice(user.Username, moderatorList) == false {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Nice try bud." + args.UserId,
		}, nil
	}

	if len(argumentArray) == 1 {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "No action specified.",
		}, nil
	}

	//globalban
	if argumentArray[1] == "globalban" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "/mod globalban <user> [silent]",
			}, nil
		}
		//no silent argument
		if len(argumentArray) == 3 {
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			p.API.DeleteUser(targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has banned user " + argumentArray[2],
				Username:     "System",
			}, nil
		}
		//silent argument
		if len(argumentArray) == 4 {
			if argumentArray[3] != "silent" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Invalid fourth argument",
				}, nil
			}
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			p.API.DeleteUser(targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You banned " + argumentArray[2],
			}, nil
		}
	}

	//teamkick
	if argumentArray[1] == "teamkick" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "/mod teamkick <user> [silent]",
			}, nil
		}
		//no silent argument
		if len(argumentArray) == 3 {
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			p.API.DeleteTeamMember(args.TeamId, targetuser.Id, args.UserId)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has banned user " + argumentArray[2] + " from the team",
				Username:     "System",
			}, nil
		}
		//silent argument
		if len(argumentArray) == 4 {
			if argumentArray[3] != "silent" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Invalid fourth argument",
				}, nil
			}
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			p.API.DeleteTeamMember(args.TeamId, targetuser.Id, args.UserId)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You banned " + argumentArray[2] + " from the team",
			}, nil
		}
	}

	//teamkick
	if argumentArray[1] == "unmute" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "/mod unmute <user>",
			}, nil
		}
		//no silent argument
		if len(argumentArray) == 3 {
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			muteduserids = remove(muteduserids, targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has unmuted user " + argumentArray[2],
				Username:     "System",
			}, nil
		}
	}

	//mute
	if argumentArray[1] == "mute" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "/mod mute <user> [silent]",
			}, nil
		}
		//no silent argument
		if len(argumentArray) == 3 {
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			muteduserids = append(muteduserids, targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has muted user " + argumentArray[2],
				Username:     "System",
			}, nil
		}
		//silent argument
		if len(argumentArray) == 4 {
			if argumentArray[3] != "silent" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Invalid fourth argument",
				}, nil
			}
			var targetuser *model.User
			var err2 *model.AppError
			targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])

			if err2 != nil {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "An error has occured. Please contact your adminstrator.",
				}, nil
			}
			muteduserids = append(muteduserids, targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You muted " + argumentArray[2],
			}, nil
		}
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "That subcommand doesn't exist.",
	}, nil

}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	if stringInSlice(post.UserId, muteduserids) == true {
		return nil, "You are muted!"
	}
	return post, ""
}
