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
	auditChannel := strings.TrimSpace(p.getConfiguration().AuditChannel)
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
			
			var reportPost *model.Post
			reportPost = &model.Post{
				UserId:    user.Id,
				ChannelId: reportChannel,
				Message:   ":bug: " + strings.Replace(args.Command, "/report bug ", "", 1),
			}

			p.API.CreatePost(reportPost)
			
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
				Text:         "Your report has been received.",
			}, nil
		}
		if len(argumentArray) > 2 {
			if reportChannel == "" {
				return &model.CommandResponse{
					ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
					Text:         "The report channel id is not yet set",
				}, nil
			}
			
			var reportPost *model.Post
			reportPost = &model.Post {
				UserId:    user.Id,
				ChannelId: reportChannel,
				Message:   ":warning: @all " + argumentArray[1] + ": " + strings.Replace(args.Command, "/report "+argumentArray[1]+ " ", "", 1),
			}

			p.API.CreatePost(reportPost)
			
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "Your report has been received.",
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

	var auditPost *model.Post
	auditPost = &model.Post{
		UserId:    user.Id,
		ChannelId: auditChannel,
		Message:   args.Command,
	}

	p.API.CreatePost(auditPost)
	
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

	var response *model.CommandResponse
	var error *model.AppError
	response, error = nil, nil

	if response == nil {
		response, error = p.HandleUtil(argumentArray, user, moderatorList, args)
	}
	
	if len(argumentArray) > 2 && response == nil {
		if stringInSlice(argumentArray[2], moderatorList) {
			return &model.CommandResponse{
				ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
				Text:         "You can't perform an action on a mod.",
			}, nil
		}
	}
	
	if response == nil {
		response, error = p.HandleFiles(argumentArray, user, moderatorList, args)
	}

	if response == nil {
		response, error = p.HandleMute(argumentArray, user, moderatorList, args)
	}

	if response == nil {
		response, error = p.HandleUsers(argumentArray, user, moderatorList, args)
	}

	if response != nil {
		return response, error
	}

	if error != nil {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "An error has occured. Please contact your adminstrator.",
		}, nil
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
