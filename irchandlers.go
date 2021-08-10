package tmi

import (
	"fmt"
)

func tmiTwitchTvHandlers(cmd string) (func(*client, *IRCData), bool) {
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

func jtvHandlers(cmd string) (func(*client, *IRCData), bool) {
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

func otherHandlers(cmd string) (func(*client, *IRCData), bool) {
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
	// successful connection, reset the reconnect counter
	c.reconnectCounter = 0

	c.spawnPinger()

	var welcomeMessage = &WelcomeMessage{
		IRCType: ircData.Command,
		Data:    ircData,
		Type:    WELCOME,
	}

	c.callMessageHandler(WELCOME, welcomeMessage)
	// TODO: for each handler, although probably not this one...
	// if err == nil {
	// 	c.callMessageHandler(WELCOME, welcomeMessage)
	// } else {
	// 	c.callMessageHandler(WELCOME, unsetMessage)
	// 	c.callMessageHandler(UNSET, unsetMessage)
	// }
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
	var noticeMessage, err = parseNoticeMessage(ircData)
	if err != nil {
		c.warnUser(err)
		if noticeMessage.MsgID == "login_failure" {
			c.notifFatal.notify()
			return
		}
	}
	c.callMessageHandler(NOTICE, noticeMessage)
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
	select {
	case c.rcvdPong <- true:
	default:
	}
}
func (c *client) otherCommandPRIVMSG(ircData *IRCData) {
	fmt.Println("Got PRIVMSG")
}
func (c *client) otherCommandWHISPER(ircData *IRCData) {
	fmt.Println("Got WHISPER")
}

func (c *client) callMessageHandler(mt MessageType, message Message) {
	if h, ok := c.handlers[mt]; ok {
		if h != nil {
			h(message)
		}
	}
}
