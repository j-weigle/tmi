package tmi

import (
	"fmt"
)

func tmiTwitchTvHandlers(cmd string) (func(*Client, IRCData) error, bool) {
	var f func(*Client, IRCData) error
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
		f = (*Client).tmiTwitchTvCommand001
	case "421": // ERR_UNKNOWNCOMMAND RFC1459 ; invalid IRC Command
		f = (*Client).tmiTwitchTvCommand421
	case "CLEARCHAT":
		f = (*Client).tmiTwitchTvCommandCLEARCHAT
	case "CLEARMSG":
		f = (*Client).tmiTwitchTvCommandCLEARMSG
	case "GLOBALUSERSTATE":
		f = (*Client).tmiTwitchTvCommandGLOBALUSERSTATE
	case "HOSTTARGET":
		f = (*Client).tmiTwitchTvCommandHOSTTARGET
	case "NOTICE":
		f = (*Client).tmiTwitchTvCommandNOTICE
	case "RECONNECT":
		f = (*Client).tmiTwitchTvCommandRECONNECT
	case "ROOMSTATE":
		f = (*Client).tmiTwitchTvCommandROOMSTATE
	case "USERNOTICE":
		f = (*Client).tmiTwitchTvCommandUSERNOTICE
	case "USERSTATE":
		f = (*Client).tmiTwitchTvCommandUSERSTATE

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func jtvHandlers(cmd string) (func(*Client, IRCData) error, bool) {
	var f func(*Client, IRCData) error
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

func otherHandlers(cmd string) (func(*Client, IRCData) error, bool) {
	var f func(*Client, IRCData) error
	var ok bool

	switch cmd {
	// UNIMPLEMENTED
	// 366 ; RPL_ENDOFNAMES RFC1459 ; end of NAMES
	case "366":
		ok = true

	// IMPLEMENTED
	case "353": // RPL_NAMREPLY RFC1459 ; aka NAMES on twitch dev docs
		f = (*Client).otherCommand353
	case "JOIN":
		f = (*Client).otherCommandJOIN
	case "PART":
		f = (*Client).otherCommandPART
	case "PING":
		f = (*Client).otherCommandPING
	case "PONG":
		f = (*Client).otherCommandPONG
	case "PRIVMSG":
		f = (*Client).otherCommandPRIVMSG
	case "WHISPER":
		f = (*Client).otherCommandWHISPER

	// NOT HANDLED
	default:
		ok = false
	}

	if f != nil {
		return f, true
	}
	return f, ok
}

func (c *Client) tmiTwitchTvCommand001(ircData IRCData) error {
	c.connected.set(true)
	go c.onConnectedJoins()
	// successful connection, reset the reconnect counter
	c.reconnectCounter = 0

	var welcomeMessage = WelcomeMessage{
		Data:    ircData,
		IRCType: ircData.Command,
		Type:    WELCOME,
	}

	if c.handlers.onWelcomeMessage != nil {
		c.handlers.onWelcomeMessage(welcomeMessage)
	}
	return nil
}
func (c *Client) tmiTwitchTvCommand421(ircData IRCData) error {
	var invalidIRCMessage, parseErr = parseInvalidIRCMessage(ircData)
	if parseErr != nil {
		c.warnUser(parseErr)
	}

	if c.handlers.onInvalidIRCMessage != nil {
		c.handlers.onInvalidIRCMessage(invalidIRCMessage)
	}
	return nil
}
func (c *Client) tmiTwitchTvCommandCLEARCHAT(ircData IRCData) error {
	fmt.Println("Got CLEARCHAT")
	return nil
}
func (c *Client) tmiTwitchTvCommandCLEARMSG(ircData IRCData) error {
	fmt.Println("Got CLEARMSG")
	return nil
}
func (c *Client) tmiTwitchTvCommandHOSTTARGET(ircData IRCData) error {
	fmt.Println("Got HOSTTARGET")
	return nil
}
func (c *Client) tmiTwitchTvCommandNOTICE(ircData IRCData) error {
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
func (c *Client) tmiTwitchTvCommandGLOBALUSERSTATE(ircData IRCData) error {
	fmt.Println("Got GLOBALUSERSTATE")
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
func (c *Client) otherCommandPING(ircData IRCData) error {
	c.send("PONG :tmi.twitch.tv")
	return nil
}
func (c *Client) otherCommandPONG(ircData IRCData) error {
	select {
	case c.rcvdPong <- struct{}{}:
	default:
	}
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
