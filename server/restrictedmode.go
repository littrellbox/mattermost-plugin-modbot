package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/model"
)

func (p *Plugin) HandleRestrictedMode(argumentArray []string, user *model.User, moderatorList []string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if argumentArray[1] == "resmodeenable" {
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

			restrictedmodeusers = append(restrictedmodeusers, targetuser.Id)
			p.API.KVSet("modbot_resmodeusers", []byte(strings.Join(restrictedmodeusers, ",")))
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You put " + argumentArray[2] + " in to restricted mode.",
			}, nil
		}
	}
	if argumentArray[1] == "resmodedisable" {
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

			restrictedmodeusers = remove(restrictedmodeusers, targetuser.Id)
			p.API.KVSet("modbot_resmodeusers", []byte(strings.Join(restrictedmodeusers, ",")))
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You removed " + argumentArray[2] + " from restricted mode.",
			}, nil
		}
	}
	return nil, nil
}
