package tmi

import (
	"errors"
)

var (
	ErrUnrecognizedIRCCommand = errors.New("unrecognized IRC Command")
	ErrUnsetIRCCommand        = errors.New("unset IRC Command")
)

func (c *Client) handleIRCMessage(rawMessage string) error {
	var data, errParseIRC = parseIRCMessage(rawMessage)
	if errParseIRC != nil {
		return c.unsetHandler(data)
	}

	var err = c.tmiHandlers(data)
	if err == ErrUnsetIRCCommand {
		return c.unsetHandler(data)
	}
	if err == ErrUnrecognizedIRCCommand {
		c.warnUser(errors.New("unrecognized message with { " + data.Prefix + " } prefix:\n" + rawMessage))
		return nil
	}
	return err
}

func (c *Client) unsetHandler(data IRCData) error {
	if c.handlers.onUnsetMessage != nil {
		c.handlers.onUnsetMessage(parseUnsetMessage(data))
	}
	return nil
}

func (c *Client) tmiHandlers(data IRCData) error {
	switch data.Command {
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
		var pingMessage = parsePingMessage(data)
		if pingMessage.Text != "" {
			c.send("PONG :" + pingMessage.Text)
		}
		if c.handlers.onPingMessage != nil {
			c.handlers.onPingMessage(parsePingMessage(data))
		}
		return nil

	case "PONG":
		var pongMessage = parsePongMessage(data)
		if pongMessage.Text == pingSignature {
			select {
			case c.rcvdPong <- struct{}{}:
			default:
			}
		}
		if c.handlers.onPongMessage != nil {
			c.handlers.onPongMessage(parsePongMessage(data))
		}
		return nil

	case "PRIVMSG":
		if c.handlers.onPrivateMessage != nil {
			c.handlers.onPrivateMessage(parsePrivateMessage(data))
		}
		return nil

	case "WHISPER":
		if c.handlers.onWhisperMessage != nil {
			c.handlers.onWhisperMessage(parseWhisperMessage(data))
		}
		return nil

	// UNIMPLEMENTED
	// ------------------
	// tmi.twitch.tv
	// ------------------
	// 002  ; RPL_YOURHOST  RFC2812 ; "Your host is tmi.twitch.tv"
	// 003  ; RPL_CREATED   RFC2812 ; "This server is rather new"
	// 004  ; RPL_MYINFO    RFC2812 ; "-" ; part of post registration greeting
	// 375  ; RPL_MOTDSTART RFC1459 ; "-" ;  start message of the day
	// 372  ; RPL_MOTD      RFC1459 ; "You are in a maze of twisty passages, all alike" ; message of the day
	// 376  ; RPL_ENDOFMOTD RFC1459 ; "\u003e" which is a > ; end message of the day
	// CAP  ; CAP * ACK ; acknowledgement of membership
	// 421  ; ERR_UNKNOWNCOMMAND RFC1459 ; invalid IRC Command
	//
	// ------------------
	// jtv
	// ------------------
	// MODE ; deprecated
	//
	// ------------------
	// other or no prefix
	// ------------------
	// 366 ; RPL_ENDOFNAMES RFC1459 ; end of NAMES
	case "002", "003", "004", "375", "372", "376", "CAP", "SERVERCHANGE", "421", "MODE", "366":
		return ErrUnsetIRCCommand

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}
