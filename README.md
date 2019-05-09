# Modbot
Mattermost moderation tool.

# Contributing
```
go get github.com/mattermost/mattermost-server
go get github.com/pkg/errors
make dist
```
# Usage:
Make sure the moderator list and report channel are set in the settings. You can get your channel ID using /mod channelid.

/mod:
/mod teamkick \<user\> \[silent\] - Kicks a user from the channel you are currently in.
/mod globalban \<user\> \[silent\] - Deactivates a user's account.
/mod mute <user> \[silent\] - Mutes a user.
/mod unmute <user> - Unmutes a user. (you can't do this silently as the user has to know they were unmuted. This will likely be changed in the future) 
/mod togglefiles - Disables file uploading in the team you are in. (toggle) 
/mod mutechannel - Mutes the channel you are in (toggle) 
/mod channelid - (DEBUG) Prints the channel ID of the channel you are in /
/mod userid - (DEBUG) Prints your user ID 
/mod userid \<name\> - (DEBUG) Prints the user specified's user ID

/report:
/report \<user\> \<report\> - Reports a user
/report bug \<report\> - Reports a bug