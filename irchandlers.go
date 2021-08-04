package tmi

import (
	"fmt"
	"strings"
)

func tmiTwitchTvCommands(cmd string) (func(*client, *IRCData), bool) {
	var f func(*client, *IRCData)
	var ok bool

	switch cmd {
	// UNIMPLEMENTED
	// 002 ; RPL_YOURHOST  RFC2812 ; "Your host is tmi.twitch.tv"
	// 003 ; RPL_CREATED   RFC2812 ; "This server is rather new"
	// 004 ; RPL_MYINFO    RFC2812 ; "-" ; part of post registration greeting
	// 375 ; RPL_MOTDSTART RFC1459 ; "-" ;  start message of the day
	// 372 ; RPL_MOTD      RFC1459 ; "You are in a maze of twisty passages, all alike" ; message of the day
	// 376 ; RPL_ENDOFMOTD RFC1459 ; "\u003e" which is a > ; end message of the day
	// CAP ; CAP * ACK ; acknowledgement of membership
	case "002", "003", "004", "375", "372", "376", "CAP", "SERVERCHANGE":
		ok = true

	// IMPLEMENTED
	case "001": // RPL_WELCOME        RFC2812 ; "Welcome, GLHF"
		f = (*client).tmiTwitchTvCommand001
	case "421": // ERR_UNKNOWNCOMMAND RFC1459 ; invalid IRC Command
		f = (*client).tmiTwitchTvCommand421
	case "CLEARCHAT":
		f = (*client).tmiTwitchTvCommandCLEARCHAT
	case "CLEARMSG":
		f = (*client).tmiTwitchTvCommandCLEARMSG
	case "GLOBALUSERSTATE":
		f = (*client).tmiTwitchTvCommandGLOBALUSERSTATE
	case "HOSTTARGET":
		f = (*client).tmiTwitchTvCommandHOSTTARGET
	case "NOTICE":
		f = (*client).tmiTwitchTvCommandNOTICE
	case "RECONNECT":
		f = (*client).tmiTwitchTvCommandRECONNECT
	case "ROOMSTATE":
		f = (*client).tmiTwitchTvCommandROOMSTATE
	case "USERNOTICE":
		f = (*client).tmiTwitchTvCommandUSERNOTICE
	case "USERSTATE":
		f = (*client).tmiTwitchTvCommandUSERSTATE

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func jtvCommands(cmd string) (func(*client, *IRCData), bool) {
	var f func(*client, *IRCData)
	var ok bool

	switch cmd {
	// UNIMPLEMENTED
	// nil

	// IMPLEMENTED
	case "MODE":
		f = (*client).jtvCommandMODE

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func otherCommands(cmd string) (func(*client, *IRCData), bool) {
	var f func(*client, *IRCData)
	var ok bool

	switch cmd {
	// UNIMPLEMENTED
	// 366 ; RPL_ENDOFNAMES RFC1459 ; end of NAMES
	case "366":
		ok = true

	// IMPLEMENTED
	case "353": // RPL_NAMREPLY RFC1459 ; aka NAMES on twitch dev docs
		f = (*client).otherCommand353
	case "JOIN":
		f = (*client).otherCommandJOIN
	case "PART":
		f = (*client).otherCommandPART
	case "PING":
		f = (*client).otherCommandPING
	case "PONG":
		f = (*client).otherCommandPONG
	case "PRIVMSG":
		f = (*client).otherCommandPRIVMSG
	case "WHISPER":
		f = (*client).otherCommandWHISPER

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func (c *client) tmiTwitchTvCommand001(ircData *IRCData) {
	// TODO: set ping loop to confirm still connected
	fmt.Println("Got 001: Welcome, GLHF") // "Welcome, GLHF"
}
func (c *client) tmiTwitchTvCommand421(ircData *IRCData) {
	fmt.Println("Got 421: invalid IRC command") // invalid IRC command
}
func (c *client) tmiTwitchTvCommandCLEARCHAT(ircData *IRCData) {
	fmt.Println("Got CLEARCHAT")
}
func (c *client) tmiTwitchTvCommandCLEARMSG(ircData *IRCData) {
	fmt.Println("Got CLEARMSG")
}
func (c *client) tmiTwitchTvCommandHOSTTARGET(ircData *IRCData) {
	fmt.Println("Got HOSTTARGET")
}
func (c *client) tmiTwitchTvCommandNOTICE(ircData *IRCData) {
	var noticeMessage = &Message{
		IRCType: ircData.Command,
		RawIRC:  ircData,
		//Info
		//Type
		//Userstate
	}
	if len(ircData.Params) >= 1 {
		noticeMessage.From = strings.TrimPrefix(ircData.Params[0], "#")
	}
	var msg string
	if len(ircData.Params) >= 2 {
		msg = ircData.Params[1]
		noticeMessage.Text = msg
	}
	if msgId, ok := ircData.Tags["msg-id"]; ok {
		noticeMessage.MsgId = msgId

		switch msgId {
		// TODO: organize this wall of cases
		case "already_banned":
		case "already_emote_only_off":
		case "already_emote_only_on":
		case "already_r9k_off":
		case "already_r9k_on":
		case "already_subs_off":
		case "already_subs_on":
		case "bad_ban_admin":
		case "bad_ban_anon":
		case "bad_ban_broadcaster":
		case "bad_ban_mod":
		case "bad_ban_self":
		case "bad_ban_staff":
		case "bad_commercial_error":
		case "bad_delete_message_broadcaster":
		case "bad_delete_message_mod":
		case "bad_host_error":
		case "bad_host_hosting":
		case "bad_host_rate_exceeded":
		case "bad_host_rejected":
		case "bad_host_self":
		case "bad_marker_client":
		case "bad_mod_banned":
		case "bad_mod_mod":
		case "bad_slow_duration":
		case "bad_timeout_admin":
		case "bad_timeout_anon":
		case "bad_timeout_broadcaster":
		case "bad_timeout_duration":
		case "bad_timeout_mod":
		case "bad_timeout_self":
		case "bad_timeout_staff":
		case "bad_unban_no_ban":
		case "bad_unhost_error":
		case "bad_unmod_mod":
		case "bad_unvip_grantee_not_vip":
		case "bad_vip_grantee_already_vip":
		case "bad_vip_grantee_banned":
		case "ban_success":
		case "cmds_available":
		case "color_changed":
		case "commercial_success":
		case "delete_message_success":
		case "emote_only_off":
		case "emote_only_on":
		case "followers_off":
		case "followers_on":
		case "followers_onzero":
		case "host_off":
		case "host_on":
		case "host_success":
		case "host_success_viewers":
		case "host_target_went_offline":
		case "hosts_remaining":
		case "invalid_user":
		case "mod_success":
		case "msg_banned":
		case "msg_bad_characters":
		case "msg_channel_blocked":
		case "msg_channel_suspended":
		case "msg_duplicate":
		case "msg_emoteonly":
		case "msg_facebook":
		case "msg_followersonly":
		case "msg_followersonly_followed":
		case "msg_followersonly_zero":
		case "msg_r9k":
		case "msg_ratelimit":
		case "msg_rejected":
		case "msg_rejected_mandatory":
		case "msg_room_not_found":
		case "msg_slowmode":
		case "msg_subsonly":
		case "msg_suspended":
		case "msg_timedout":
		case "msg_verified_email":
		case "no_help":
		case "no_mods":
		case "no_vips":
		case "not_hosting":
		case "no_permission":
		case "r9k_off":
		case "r9k_on":
		case "raid_error_already_raiding":
		case "raid_error_forbidden":
		case "raid_error_self":
		case "raid_error_too_many_viewers":
		case "raid_error_unexpected":
		case "raid_notice_mature":
		case "raid_notice_restricted_chat":
		case "room_mods":
		case "slow_off":
		case "slow_on":
		case "subs_off":
		case "subs_on":
		case "timeout_no_timeout":
		case "timeout_success":
		case "tos_ban":
		case "turbo_only_color":
		case "unban_success":
		case "unmod_success":
		case "unraid_error_no_active_raid":
		case "unraid_error_unexpected":
		case "unraid_success":
		case "unrecognized_cmd":
		case "unsupported_chatrooms_cmd":
		case "untimeout_banned":
		case "untimeout_success":
		case "unvip_success":
		case "usage_ban":
		case "usage_clear":
		case "usage_color":
		case "usage_commercial":
		case "usage_disconnect":
		case "usage_emote_only_off":
		case "usage_emote_only_on":
		case "usage_followers_off":
		case "usage_followers_on":
		case "usage_help":
		case "usage_host":
		case "usage_marker":
		case "usage_me":
		case "usage_mod":
		case "usage_mods":
		case "usage_r9k_off":
		case "usage_r9k_on":
		case "usage_raid":
		case "usage_slow_off":
		case "usage_slow_on":
		case "usage_subs_off":
		case "usage_subs_on":
		case "usage_timeout":
		case "usage_unban":
		case "usage_unhost":
		case "usage_unmod":
		case "usage_unraid":
		case "usage_untimeout":
		case "usage_vip":
		case "vip_success":
		case "whisper_banned":
		case "whisper_banned_recipient":
		case "whisper_invalid_args":
		case "whisper_invalid_login":
		case "whisper_invalid_self":
		case "whisper_limit_per_min":
		case "whisper_limit_per_sec":
		case "whisper_restricted":
		case "whisper_restricted_recipient":
		default:
			noticeMessage.Type = "notice"
		}
	} else {
		loginFailures := []string{
			"Login unsuccessful",
			"Login authentication failed",
			"Error logging in",
			"Improperly formatted auth",
			"Invalid NICK",
		}
		for _, failure := range loginFailures {
			if strings.Contains(msg, failure) {
				c.err <- fmt.Errorf("login authentication:\n" + msg)
				err := c.Disconnect()
				if err != nil {
					c.CloseConnection()
				}
				return
			}
		}
		c.err <- fmt.Errorf("could not parse NOTICE:\n" + ircData.Raw)
	}
}
func (c *client) tmiTwitchTvCommandRECONNECT(ircData *IRCData) {
	fmt.Println("Got RECONNECT")
}
func (c *client) tmiTwitchTvCommandROOMSTATE(ircData *IRCData) {
	fmt.Println("Got ROOMSTATE")
}
func (c *client) tmiTwitchTvCommandUSERNOTICE(ircData *IRCData) {
	fmt.Println("Got USERNOTICE")
}
func (c *client) tmiTwitchTvCommandUSERSTATE(ircData *IRCData) {
	fmt.Println("Got USERSTATE")
}
func (c *client) tmiTwitchTvCommandGLOBALUSERSTATE(ircData *IRCData) {
	fmt.Println("Got GLOBALUSERSTATE")
}

func (c *client) jtvCommandMODE(ircData *IRCData) {
	fmt.Println("Got MODE")
}

func (c *client) otherCommand353(ircData *IRCData) {
	fmt.Println("Got 353: NAMES")
}
func (c *client) otherCommandJOIN(ircData *IRCData) {
	fmt.Println("Got JOIN")
}
func (c *client) otherCommandPART(ircData *IRCData) {
	fmt.Println("Got PART")
}
func (c *client) otherCommandPING(ircData *IRCData) {
	c.send("PONG :tmi.twitch.tv")
}
func (c *client) otherCommandPONG(ircData *IRCData) {
	// TODO: handle responses from sent pings to check latency and timeouts
	fmt.Println("Got PONG")
}
func (c *client) otherCommandPRIVMSG(ircData *IRCData) {
	fmt.Println("Got PRIVMSG")
}
func (c *client) otherCommandWHISPER(ircData *IRCData) {
	fmt.Println("Got WHISPER")
}
