package tmi

import (
	"fmt"
)

func tmiTwitchTvHandlers(cmd string) (func(*client, *IRCData) error, bool) {
	var f func(*client, *IRCData) error
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

func jtvHandlers(cmd string) (func(*client, *IRCData) error, bool) {
	var f func(*client, *IRCData) error
	var ok bool

	switch cmd {
	// UNIMPLEMENTED
	case "MODE": // deprecated
		ok = true

	// IMPLEMENTED
	// nil

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func otherHandlers(cmd string) (func(*client, *IRCData) error, bool) {
	var f func(*client, *IRCData) error
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

func (c *client) tmiTwitchTvCommand001(ircData *IRCData) error {
	// successful connection, reset the reconnect counter
	c.reconnectCounter = 0

	var welcomeMessage = &WelcomeMessage{
		Data:    ircData,
		IRCType: ircData.Command,
		Type:    WELCOME,
	}

	c.callMessageHandler(WELCOME, welcomeMessage)
	return nil
}
func (c *client) tmiTwitchTvCommand421(ircData *IRCData) error {
	var invalidIRCMessage, parseErr = parseInvalidIRCMessage(ircData)
	if parseErr != nil {
		c.warnUser(parseErr)
	}

	c.callMessageHandler(INVALIDIRC, invalidIRCMessage)
	return nil
}
func (c *client) tmiTwitchTvCommandCLEARCHAT(ircData *IRCData) error {
	fmt.Println("Got CLEARCHAT")
	return nil
}
func (c *client) tmiTwitchTvCommandCLEARMSG(ircData *IRCData) error {
	fmt.Println("Got CLEARMSG")
	return nil
}
func (c *client) tmiTwitchTvCommandHOSTTARGET(ircData *IRCData) error {
	fmt.Println("Got HOSTTARGET")
	return nil
}
func (c *client) tmiTwitchTvCommandNOTICE(ircData *IRCData) error {
	var err error
	var noticeMessage, parseErr = parseNoticeMessage(ircData)
	if parseErr != nil {
		c.warnUser(parseErr)
		if noticeMessage.MsgID == "login_failure" {
			err = ErrLoginFailure
		}
	}
	c.callMessageHandler(NOTICE, noticeMessage)
	return err
}
func (c *client) tmiTwitchTvCommandRECONNECT(ircData *IRCData) error {
	fmt.Println("Got RECONNECT")
	return nil
}
func (c *client) tmiTwitchTvCommandROOMSTATE(ircData *IRCData) error {
	fmt.Println("Got ROOMSTATE")
	return nil
}
func (c *client) tmiTwitchTvCommandUSERNOTICE(ircData *IRCData) error {
	fmt.Println("Got USERNOTICE")
	return nil
}
func (c *client) tmiTwitchTvCommandUSERSTATE(ircData *IRCData) error {
	fmt.Println("Got USERSTATE")
	return nil
}
func (c *client) tmiTwitchTvCommandGLOBALUSERSTATE(ircData *IRCData) error {
	fmt.Println("Got GLOBALUSERSTATE")
	return nil
}

func (c *client) otherCommand353(ircData *IRCData) error {
	fmt.Println("Got 353: NAMES")
	return nil
}
func (c *client) otherCommandJOIN(ircData *IRCData) error {
	fmt.Println("Got JOIN")
	return nil
}
func (c *client) otherCommandPART(ircData *IRCData) error {
	fmt.Println("Got PART")
	return nil
}
func (c *client) otherCommandPING(ircData *IRCData) error {
	c.send("PONG :tmi.twitch.tv")
	return nil
}
func (c *client) otherCommandPONG(ircData *IRCData) error {
	select {
	case c.rcvdPong <- struct{}{}:
	default:
	}
	return nil
}
func (c *client) otherCommandPRIVMSG(ircData *IRCData) error {
	fmt.Println("Got PRIVMSG")
	return nil
}
func (c *client) otherCommandWHISPER(ircData *IRCData) error {
	fmt.Println("Got WHISPER")
	return nil
}
