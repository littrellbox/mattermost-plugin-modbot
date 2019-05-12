package main

import (
	"github.com/mattermost/mattermost-server/model"
)

func (p *Plugin) HandleFiles(argumentArray []string, user *model.User, moderatorList []string, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	//userfiles
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

	return nil, nil
}
