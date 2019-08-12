package main

import (
	"github.com/mattermost/mattermost-server/model"
)

//HandleMute Handle running mute command
func (p *Plugin) HandleMute(argumentArray []string, user *model.User, moderatorList []string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	//unmute
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

	//mutechannel
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

	return nil, nil
}
