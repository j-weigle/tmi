package tmi

import (
	"errors"
	"fmt"
)

var (
	ErrUnrecognizedIRCCommand = errors.New("unrecognized IRC Command")
	ErrUnsetIRCCommand        = errors.New("unset IRC Command")
)

func (c *Client) tmiTwitchTvHandlers(data IRCData) error {
	switch data.Command {
	// UNIMPLEMENTED
	// 002 ; RPL_YOURHOST  RFC2812 ; "Your host is tmi.twitch.tv"
	// 003 ; RPL_CREATED   RFC2812 ; "This server is rather new"
	// 004 ; RPL_MYINFO    RFC2812 ; "-" ; part of post registration greeting
	// 375 ; RPL_MOTDSTART RFC1459 ; "-" ;  start message of the day
	// 372 ; RPL_MOTD      RFC1459 ; "You are in a maze of twisty passages, all alike" ; message of the day
	// 376 ; RPL_ENDOFMOTD RFC1459 ; "\u003e" which is a > ; end message of the day
	// CAP ; CAP * ACK ; acknowledgement of membership
	// 421 ; ERR_UNKNOWNCOMMAND RFC1459 ; invalid IRC Command
	case "002", "003", "004", "375", "372", "376", "CAP", "SERVERCHANGE", "421":
		return ErrUnsetIRCCommand

	// IMPLEMENTED
	case "001": // RPL_WELCOME        RFC2812 ; "Welcome, GLHF"
		c.connected.set(true)
		go c.onConnectedJoins()
		// successful connection, reset the reconnect counter
		c.reconnectCounter = 0

		if c.handlers.onConnected != nil {
			c.handlers.onConnected()
		}
		return nil

	case "CLEARCHAT":
		if c.handlers.onClearChatMessage != nil {
			c.handlers.onClearChatMessage(parseClearChatMessage(data))
		}
		return nil

	case "CLEARMSG":
		if c.handlers.onClearMsgMessage != nil {
			c.handlers.onClearMsgMessage(parseClearMsgMessage(data))
		}
		return nil

	case "GLOBALUSERSTATE":
		if c.handlers.onGlobalUserstateMessage != nil {
			c.handlers.onGlobalUserstateMessage(parseGlobalUserstateMessage(data))
		}
		return nil

	case "HOSTTARGET":
		if c.handlers.onHostTargetMessage != nil {
			c.handlers.onHostTargetMessage(parseHostTargetMessage(data))
		}
		return nil

	case "NOTICE":
		var err error
		var noticeMessage, parseErr = parseNoticeMessage(data)
		if parseErr != nil {
			c.warnUser(parseErr)
			if noticeMessage.MsgID == "login_failure" {
				err = ErrLoginFailure
			}
		}
		if c.handlers.onNoticeMessage != nil {
			c.handlers.onNoticeMessage(noticeMessage)
		}
		return err

	case "RECONNECT":
		if c.handlers.onReconnectMessage != nil {
			c.handlers.onReconnectMessage(parseReconnectMessage(data))
		}
		return errReconnect

	case "ROOMSTATE":
		if c.handlers.onRoomstateMessage != nil {
			c.handlers.onRoomstateMessage(parseRoomstateMessage(data))
		}
		return nil

	case "USERNOTICE":
		if c.handlers.onUserNoticeMessage != nil {
			c.handlers.onUserNoticeMessage(parseUsernoticeMessage(data))
		}
		return nil

	case "USERSTATE":
		if c.handlers.onUserstateMessage != nil {
			c.handlers.onUserstateMessage(parseUserstateMessage(data))
		}
		return nil

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}

func (c *Client) jtvHandlers(data IRCData) error {
	switch data.Command {
	// UNIMPLEMENTED
	case "MODE": // deprecated
		return ErrUnsetIRCCommand

	// IMPLEMENTED
	// nil

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}

func (c *Client) otherHandlers(data IRCData) error {
	switch data.Command {
	// UNIMPLEMENTED
	// 366 ; RPL_ENDOFNAMES RFC1459 ; end of NAMES
	case "366":
		return ErrUnsetIRCCommand

	// IMPLEMENTED
	case "353": // RPL_NAMREPLY RFC1459 ; aka NAMES on twitch dev docs
		// WARNING: deprecated, but not removed yet
		if c.handlers.onNamesMessage != nil {
			c.handlers.onNamesMessage(parseNamesMessage(data))
		}
		return nil

	case "JOIN":
		if c.handlers.onJoinMessage != nil {
			c.handlers.onJoinMessage(parseJoinMessage(data))
		}
		return nil

	case "PART":
		if c.handlers.onPartMessage != nil {
			c.handlers.onPartMessage(parsePartMessage(data))
		}
		return nil

	case "PING":
		c.send("PONG :tmi.twitch.tv")
		return nil

	case "PONG":
		select {
		case c.rcvdPong <- struct{}{}:
		default:
		}
		return nil

	case "PRIVMSG":
		if c.handlers.onPrivateMessage != nil {
			c.handlers.onPrivateMessage(parsePrivateMessage(data))
		}
		return nil

	case "WHISPER":
		return c.otherCommandWHISPER(data)

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}

func (c *Client) otherCommandWHISPER(ircData IRCData) error {
	fmt.Println("Got WHISPER")
	return nil
}
