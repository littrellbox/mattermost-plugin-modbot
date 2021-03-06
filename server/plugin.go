package main

import (
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

//Plugin The plugin itself
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
var restrictedmodeusers []string

const (
	trigger       string = "mod"
	reporttrigger string = "report"
	displayname   string = "System Admin"
)

//OnActivate Activates the plugin
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
	KVResModeUsers, err3 := p.API.KVGet("modbot_resmodeusers")

	if err3 != nil {
		return nil
	}

	if KVResModeUsers != nil {
		restrictedmodeusers = strings.Split(strings.TrimSpace(string(KVResModeUsers)), ",")
	}
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

//ExecuteCommand Runs the commands
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
			reportPost = &model.Post{
				UserId:    user.Id,
				ChannelId: reportChannel,
				Message:   ":warning: @all " + argumentArray[1] + ": " + strings.Replace(args.Command, "/report "+argumentArray[1]+" ", "", 1),
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
	
	if response == nil {
		response, error = p.HandleRestrictedMode(argumentArray, user, moderatorList, args)
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

func (p *Plugin) MessageHasBeenUpdated(c *plugin.Context, newPost, oldPost *model.Post) {

	//use plugin.getSession to get Session object then use the Session object to retrive User object
	moderatorList := strings.Split(strings.TrimSpace(p.getConfiguration().Moderators), ",")
	auditChannel := strings.TrimSpace(p.getConfiguration().AuditChannel)
	
	var targetsession *model.Session
	var err *model.AppError
	
	targetsession, err = p.API.GetSession(c.SessionId)
	if err != nil {
		p.API.SendEphemeralPost(newPost.UserId, &model.Post{
			ChannelId: newPost.ChannelId,
			Message:   "An error has occured determining if you are a moderator or not. (3):GetSession",
		})
		return
	}
	
	var targetuser *model.User
	var err2 *model.AppError
	
	targetuser, err2 = p.API.GetUser(targetsession.UserId)
	if err2 != nil {
		p.API.SendEphemeralPost(newPost.UserId, &model.Post{
			ChannelId: newPost.ChannelId,
			Message:   "An error has occured determining if you are a moderator or not. (3):GetUser",
		})
		return
	}
	
	var originaluser *model.User
	var err3 *model.AppError
	
	originaluser, err3 = p.API.GetUser(oldPost.UserId)
	if err3 != nil {
		p.API.SendEphemeralPost(newPost.UserId, &model.Post{
			ChannelId: newPost.ChannelId,
			Message:   "An error has occured determining the username of the original poster. (4):GetUser",
		})
		return
	}
	
	if stringInSlice(targetuser.Username, moderatorList) {
		var auditPost *model.Post
		auditPost = &model.Post{
			UserId:    newPost.UserId,
			ChannelId: auditChannel,
			Message:   targetuser.Username + " edited " + originaluser.Username + "'s post.",
		}
		p.API.CreatePost(auditPost)
	}
}

//MessageWillBePosted Handles mute and no-file settings
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	moderatorList := strings.Split(strings.TrimSpace(p.getConfiguration().Moderators), ",")
	var targetchannel *model.Channel
	var err2 *model.AppError
	targetchannel, err2 = p.API.GetChannel(post.ChannelId)
	if err2 != nil {
		p.API.SendEphemeralPost(post.UserId, &model.Post{
			ChannelId: post.ChannelId,
			Message:   "An error has occured determining if the file could be uploaded or not. (2):GetChannel",
		})
		return nil, ""
	}
	var targetuser *model.User
	targetuser, err2 = p.API.GetUser(post.UserId)
	if err2 != nil {
		p.API.SendEphemeralPost(post.UserId, &model.Post{
			ChannelId: post.ChannelId,
			Message:   "An error has occured determining if the file could be uploaded or not. (2):GetUser",
		})
		return nil, ""
	}
	
	if stringInSlice(targetuser.Id, restrictedmodeusers) {
        if len(post.FileIds) != 0 && stringInSlice(targetuser.Username, moderatorList) == false {
			p.API.SendEphemeralPost(post.UserId, &model.Post{
				ChannelId: post.ChannelId,
				Message:   "You are currently in restricted mode. Users in restricted mode can't send files.",
			})
			return nil, "plugin.message_will_be_posted.dismiss_post"
		}
    }

	if stringInSlice(targetchannel.TeamId, fileblockedids) || stringInSlice(targetuser.Id, fileblockusers) {
		if len(post.FileIds) != 0 && stringInSlice(targetuser.Username, moderatorList) == false {
			p.API.SendEphemeralPost(post.UserId, &model.Post{
				ChannelId: post.ChannelId,
				Message:   "File uploading is currently disabled.",
			})
			return nil, "plugin.message_will_be_posted.dismiss_post"
		}
	}
	if stringInSlice(post.ChannelId, mutedchannelids) == true {
		if stringInSlice(targetuser.Username, moderatorList) == false {
			p.API.SendEphemeralPost(post.UserId, &model.Post{
				ChannelId: post.ChannelId,
				Message:   "Only moderators can send messages in this channel.",
			})
			return nil, "plugin.message_will_be_posted.dismiss_post"
		}
	}
	if stringInSlice(post.UserId, muteduserids) == true {
		p.API.SendEphemeralPost(post.UserId, &model.Post{
			ChannelId: post.ChannelId,
			Message:   "You are muted!",
		})
		return nil, "plugin.message_will_be_posted.dismiss_post"
	}
	return post, ""
}
