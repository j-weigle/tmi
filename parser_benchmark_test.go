package tmi

import (
	"archive/zip"
	"bufio"
	"runtime"
	"testing"
)

var messages []string

func init() {
	messages = readLogArchive("./benchmark-logs/privmsg2k.log.zip")
}

func BenchmarkParsePrivateMessageLog(b *testing.B) {
	client := NewClient(NewClientConfig("", ""))
	client.OnPrivateMessage(func(msg PrivateMessage) {})
	for n := 0; n < b.N; n++ {
		for _, msg := range messages {
			client.handleIRCMessage(msg)
		}
	}
}

func BenchmarkParseWhisperMessage(b *testing.B) {
	client := NewClient(NewClientConfig("", ""))
	client.OnWhisperMessage(func(msg WhisperMessage) {})
	testMessage := "@badges=;color=#FFFFFF;display-name=Anyone;emotes=;message-id=1;thread-id=11111111_11111111;turbo=0;user-id=12345678;user-type= :anyone!anyone@anyone.tmi.twitch.tv WHISPER someone :hello someone"
	for n := 0; n < b.N; n++ {
		client.handleIRCMessage(testMessage)
	}
}

func readLogArchive(logFile string) []string {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(logFile)
	if err != nil {
		runtime.Goexit()
	}
	defer r.Close()

	// Iterate through the files in the archive,
	logContent := []string{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			runtime.Goexit()
		}

		s := bufio.NewScanner(rc)
		for s.Scan() {
			logContent = append(logContent, s.Text())
		}
		rc.Close()
	}

	return logContent
}
