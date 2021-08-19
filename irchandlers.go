package tmi

import (
	"errors"
	"fmt"
)

var (
	ErrUnrecognizedIRCCommand = errors.New("unrecognized IRC Command")
	ErrUnsetIRCCommand        = errors.New("unset IRC Command")
)

func (c *Client) tmiTwitchTvHandlers(ircData IRCData) error {
	switch ircData.Command {
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
			var clearChatMessage = parseClearChatMessage(ircData)
			c.handlers.onClearChatMessage(clearChatMessage)
		}
		return nil

	case "CLEARMSG":
		if c.handlers.onClearMsgMessage != nil {
			var clearMsgMessage = parseClearMsgMessage(ircData)
			c.handlers.onClearMsgMessage(clearMsgMessage)
		}
		return nil

	case "GLOBALUSERSTATE":
		if c.handlers.onGlobalUserstateMessage != nil {
			var globalUserstateMessage = parseGlobalUserstateMessage(ircData)
			c.handlers.onGlobalUserstateMessage(globalUserstateMessage)
		}
		return nil

	case "HOSTTARGET":
		if c.handlers.onHostTargetMessage != nil {
			var hostTargetMessage = parseHostTargetMessage(ircData)
			c.handlers.onHostTargetMessage(hostTargetMessage)
		}
		return nil

	case "NOTICE":
		var err error
		var noticeMessage, parseErr = parseNoticeMessage(ircData)
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
		return c.tmiTwitchTvCommandRECONNECT(ircData)

	case "ROOMSTATE":
		return c.tmiTwitchTvCommandROOMSTATE(ircData)

	case "USERNOTICE":
		return c.tmiTwitchTvCommandUSERNOTICE(ircData)

	case "USERSTATE":
		return c.tmiTwitchTvCommandUSERSTATE(ircData)

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}

func (c *Client) jtvHandlers(ircData IRCData) error {
	switch ircData.Command {
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

func (c *Client) otherHandlers(ircData IRCData) error {
	switch ircData.Command {
	// UNIMPLEMENTED
	// 366 ; RPL_ENDOFNAMES RFC1459 ; end of NAMES
	case "366":
		return ErrUnsetIRCCommand

	// IMPLEMENTED
	case "353": // RPL_NAMREPLY RFC1459 ; aka NAMES on twitch dev docs
		return c.otherCommand353(ircData)

	case "JOIN":
		return c.otherCommandJOIN(ircData)

	case "PART":
		return c.otherCommandPART(ircData)

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
		return c.otherCommandPRIVMSG(ircData)

	case "WHISPER":
		return c.otherCommandWHISPER(ircData)

	// NOT RECOGNIZED
	default:
		return ErrUnrecognizedIRCCommand
	}
}

func (c *Client) tmiTwitchTvCommandRECONNECT(ircData IRCData) error {
	fmt.Println("Got RECONNECT")
	return nil
}
func (c *Client) tmiTwitchTvCommandROOMSTATE(ircData IRCData) error {
	fmt.Println("Got ROOMSTATE")
	return nil
}
func (c *Client) tmiTwitchTvCommandUSERNOTICE(ircData IRCData) error {
	fmt.Println("Got USERNOTICE")
	return nil
}
func (c *Client) tmiTwitchTvCommandUSERSTATE(ircData IRCData) error {
	fmt.Println("Got USERSTATE")
	return nil
}

func (c *Client) otherCommand353(ircData IRCData) error {
	fmt.Println("Got 353: NAMES")
	return nil
}
func (c *Client) otherCommandJOIN(ircData IRCData) error {
	fmt.Println("Got JOIN")
	return nil
}
func (c *Client) otherCommandPART(ircData IRCData) error {
	fmt.Println("Got PART")
	return nil
}
func (c *Client) otherCommandPRIVMSG(ircData IRCData) error {
	fmt.Println("Got PRIVMSG")
	return nil
}
func (c *Client) otherCommandWHISPER(ircData IRCData) error {
	fmt.Println("Got WHISPER")
	return nil
}
