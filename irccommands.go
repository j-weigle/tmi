package tmi

import "fmt"

var tmiTwitchTvCommands = map[string]func(*client, *MessageData){
	"002":          nil, // RPL_YOURHOST  RFC2812 ; "Your host is tmi.twitch.tv"
	"003":          nil, // RPL_CREATED   RFC2812 ; "This server is rather new"
	"004":          nil, // RPL_MYINFO    RFC2812 ; "-" ; part of post registration greeting
	"375":          nil, // RPL_MOTDSTART RFC1459 ; "-" ;  start message of the day
	"372":          nil, // RPL_MOTD      RFC1459 ; "You are in a maze of twisty passages, all alike" ; message of the day
	"376":          nil, // RPL_ENDOFMOTD RFC1459 ; "\u003e" which is a > ; end message of the day
	"CAP":          nil, // CAP * ACK ; acknowledgement of membership
	"SERVERCHANGE": nil,

	"001":             (*client).tmiTwitchTvCommand001, // RPL_WELCOME        RFC2812 ; "Welcome, GLHF"
	"421":             (*client).tmiTwitchTvCommand421, // ERR_UNKNOWNCOMMAND RFC1459 ; invalid IRC Command
	"CLEARCHAT":       (*client).tmiTwitchTvCommandCLEARCHAT,
	"CLEARMSG":        (*client).tmiTwitchTvCommandCLEARMSG,
	"GLOBALUSERSTATE": (*client).tmiTwitchTvCommandGLOBALUSERSTATE,
	"HOSTTARGET":      (*client).tmiTwitchTvCommandHOSTTARGET,
	"NOTICE":          (*client).tmiTwitchTvCommandNOTICE,
	"RECONNECT":       (*client).tmiTwitchTvCommandRECONNECT,
	"ROOMSTATE":       (*client).tmiTwitchTvCommandROOMSTATE,
	"USERNOTICE":      (*client).tmiTwitchTvCommandUSERNOTICE,
	"USERSTATE":       (*client).tmiTwitchTvCommandUSERSTATE,
}

var jtvCommands = map[string]func(*client, *MessageData){
	"MODE": (*client).jtvCommandMODE,
}

var otherCommands = map[string]func(*client, *MessageData){
	"366": nil, // RPL_ENDOFNAMES RFC1459

	"353":     (*client).otherCommand353, // RPL_NAMREPLY RFC1459 ; aka NAMES on twitch dev docs
	"JOIN":    (*client).otherCommandJOIN,
	"PART":    (*client).otherCommandPART,
	"PING":    (*client).otherCommandPING,
	"PONG":    (*client).otherCommandPONG,
	"PRIVMSG": (*client).otherCommandPRIVMSG,
	"WHISPER": (*client).otherCommandWHISPER,
}

func (c *client) tmiTwitchTvCommand001(msgdata *MessageData) {
	fmt.Println("Got 001: Welcome, GLHF") // "Welcome, GLHF"
}
func (c *client) tmiTwitchTvCommand421(msgdata *MessageData) {
	fmt.Println("Got 421: invalid IRC command") // invalid IRC command
}
func (c *client) tmiTwitchTvCommandCLEARCHAT(msgdata *MessageData) {
	fmt.Println("Got CLEARCHAT")
}
func (c *client) tmiTwitchTvCommandCLEARMSG(msgdata *MessageData) {
	fmt.Println("Got CLEARMSG")
}
func (c *client) tmiTwitchTvCommandHOSTTARGET(msgdata *MessageData) {
	fmt.Println("Got HOSTTARGET")
}
func (c *client) tmiTwitchTvCommandNOTICE(msgdata *MessageData) {
	fmt.Println("Got NOTICE")
}
func (c *client) tmiTwitchTvCommandRECONNECT(msgdata *MessageData) {
	fmt.Println("Got RECONNECT")
}
func (c *client) tmiTwitchTvCommandROOMSTATE(msgdata *MessageData) {
	fmt.Println("Got ROOMSTATE")
}
func (c *client) tmiTwitchTvCommandUSERNOTICE(msgdata *MessageData) {
	fmt.Println("Got USERNOTICE")
}
func (c *client) tmiTwitchTvCommandUSERSTATE(msgdata *MessageData) {
	fmt.Println("Got USERSTATE")
}
func (c *client) tmiTwitchTvCommandGLOBALUSERSTATE(msgdata *MessageData) {
	fmt.Println("Got GLOBALUSERSTATE")
}

func (c *client) jtvCommandMODE(msgdata *MessageData) {
	fmt.Println("Got MODE")
}

func (c *client) otherCommand353(msgdata *MessageData) {
	fmt.Println("Got 353: NAMES")
}
func (c *client) otherCommandJOIN(msgdata *MessageData) {
	fmt.Println("Got JOIN")
}
func (c *client) otherCommandPART(msgdata *MessageData) {
	fmt.Println("Got PART")
}
func (c *client) otherCommandPING(msgdata *MessageData) {
	c.send("PONG :tmi.twitch.tv")
}
func (c *client) otherCommandPONG(msgdata *MessageData) {
	fmt.Println("Got PONG")
}
func (c *client) otherCommandWHISPER(msgdata *MessageData) {
	fmt.Println("Got WHISPER")
}
func (c *client) otherCommandPRIVMSG(msgdata *MessageData) {
	fmt.Println("Got PRIVMSG")
}
