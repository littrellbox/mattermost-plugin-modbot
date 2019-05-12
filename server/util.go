package main

import (
	"github.com/mattermost/mattermost-server/model"
)

func (p *Plugin) HandleUtil(argumentArray []string, user *model.User, moderatorList []string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	//channelid
	if argumentArray[1] == "channelid" {
		if len(argumentArray) == 2 {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Channel ID: " + args.ChannelId,
			}, nil
		}
	}

	//userid
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
					Text:         "User ID: " + targetuser.Id,
				}, nil
			}
		}
	}
	return nil, nil
}
