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

var mutedchannelids []string
var muteduserids []string
var fileblockedids []string
var fileblockusers []string

const (
	trigger       string = "mod"
	reporttrigger string = "report"
	displayname   string = "System Admin"
)

func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		AutoComplete:     false,
		AutoCompleteDesc: "Moderation tool. If you can see this, something's broken.",
		DisplayName:      displayname,
	})
	p.API.RegisterCommand(&model.Command{
		Trigger:          reporttrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Report something.",
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
	reportChannel := strings.TrimSpace(p.getConfiguration().ReportChannel)
	if len(argumentArray) == 0 {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "huh?",
		}, nil
	}

	if strings.Contains(argumentArray[0], reporttrigger) {
		if argumentArray[1] == "bug" {
			if reportChannel == "" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "The report channel id is not yet set",
				}, nil
			}
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         ":bug: " + strings.Replace(args.Command, "/report bug ", "", 1),
				ChannelId:    reportChannel,
				Username:     "Reports",
			}, nil
		}
		if len(argumentArray) > 2 {
			if reportChannel == "" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "The report channel id is not yet set",
				}, nil
			}
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         ":warning: @all " + argumentArray[1] + ": " + strings.Replace(args.Command, "/report "+argumentArray[1]+" ", "", 1),
				ChannelId:    reportChannel,
				Username:     "Reports",
			}, nil
		}
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text: `/report <user> <report> 
			To report a bug use /report bug <report>
			Examples:
			/report bug I can't send anything!
			/report bob He stole my cookies :(`,
		}, nil
	}

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

			if targetuser.Id == args.UserId {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "You can't mute yourself.",
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

			if targetuser.Id == args.UserId {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "You can't mute yourself.",
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

	//mute
	if argumentArray[1] == "userfiles" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "/mod userfiles <user>",
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

			if stringInSlice(targetuser.Id, fileblockusers) == true {
				fileblockusers = remove(fileblockusers, targetuser.Id)
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
					Text:         "A moderator has enabled file uploads for " + targetuser.Username + ".",
					Username:     "System",
				}, nil
			}
			fileblockusers = append(fileblockusers, targetuser.Id)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has disabled file uploads for " + targetuser.Username + ".",
				Username:     "System",
			}, nil
		}
	}

	//togglefiles
	if argumentArray[1] == "togglefiles" {
		if len(argumentArray) == 2 {
			if stringInSlice(args.TeamId, fileblockedids) == true {
				fileblockedids = remove(fileblockedids, args.TeamId)
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
					Text:         "A moderator has enabled file uploads in this team.",
					Username:     "System",
				}, nil
			}
			fileblockedids = append(fileblockedids, args.TeamId)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has disabled file uploads in this team.",
				Username:     "System",
			}, nil
		}
	}

	if argumentArray[1] == "mutechannel" {
		if len(argumentArray) == 2 {
			if stringInSlice(args.ChannelId, mutedchannelids) == true {
				mutedchannelids = remove(mutedchannelids, args.ChannelId)
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
					Text:         "A moderator has enabled chat in this channel.",
					Username:     "System",
				}, nil
			}
			mutedchannelids = append(mutedchannelids, args.ChannelId)
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "A moderator has disabled chat in this channel",
				Username:     "System",
			}, nil
		}
	}

	if argumentArray[1] == "channelid" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Channel ID: " + args.ChannelId,
			}, nil
		}
	}

	if argumentArray[1] == "userid" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Your user ID: " + args.UserId,
			}, nil
		}
		if len(argumentArray) == 3 {
			if len(argumentArray) == 3 {
				var targetuser *model.User
				var err2 *model.AppError
				targetuser, err2 = p.API.GetUserByUsername(argumentArray[2])
				if err2 != nil {
					return &model.CommandResponse{
						ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
						Text:         "An error has occured. Likely the user does not exist.",
					}, nil
				}
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "Your user ID: " + targetuser.Id,
				}, nil
			}
		}
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         "That command doesn't exist.",
	}, nil

}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	moderatorList := strings.Split(strings.TrimSpace(p.getConfiguration().Moderators), ",")
	var targetchannel *model.Channel
	var err2 *model.AppError
	targetchannel, err2 = p.API.GetChannel(post.ChannelId)
	if err2 != nil {
		return nil, "An error has occured determining if the file could be uploaded or not. (2)"
	}
	var targetuser *model.User
	targetuser, err2 = p.API.GetUser(post.UserId)
	if err2 != nil {
		return nil, "An error has occured determining if the file could be uploaded or not. (2)"
	}

	if stringInSlice(targetchannel.TeamId, fileblockedids) || stringInSlice(targetuser.Id, fileblockusers) {
		if len(post.FileIds) != 0 && stringInSlice(targetuser.Username, moderatorList) == false {
			return nil, "File uploading is disabled."
		}
	}
	if stringInSlice(post.ChannelId, mutedchannelids) == true {
		if stringInSlice(targetuser.Username, moderatorList) == false {
			return nil, "You are not permitted to talk in this channel at this time."
		}
	}
	if stringInSlice(post.UserId, muteduserids) == true {
		return nil, "You are muted!"
	}
	return post, ""
}
