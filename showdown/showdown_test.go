// showdown2irc - use Showdown chat with an IRC client
// Copyright (C) 2016 Konrad Borowski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package showdown

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var showdownCommand *exec.Cmd

func runShowdown() {
	// Oh dear, this is a ride, and all that just to test script in
	// realistic environment. That's fine, however. Relatively recent
	// version of Node.js required.

	// First, update a single submodule. Others don't matter, as the
	// tests somehow run already, and they don't really matter.
	updateCommand := exec.Command("git", "submodule", "update", "--init", "../vendor/github.com/Zarel/Pokemon-Showdown")
	if err := updateCommand.Run(); err != nil {
		log.Fatal(err)
	}

	// Then update npm modules used by Pokemon Showdown.
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	showdownPath := path.Join(currentPath, "..", "vendor", "github.com", "Zarel", "Pokemon-Showdown")

	npmUpdateCommand := exec.Command("npm", "install", "--production")
	npmUpdateCommand.Dir = showdownPath
	if err := npmUpdateCommand.Run(); err != nil {
		log.Fatal(err)
	}

	// Copy sample configuration file. For some reason current version of
	// Showdown has this broken.
	configPath := path.Join(showdownPath, "config")
	os.Link(path.Join(configPath, "config-example.js"), path.Join(configPath, "config.js"))

	// Now we can run Showdown. I hope? Please tell me that's it.
	showdownCommand = exec.Command(path.Join(showdownPath, "pokemon-showdown"))
	showdownCommand.Dir = showdownPath
	stdout, err := showdownCommand.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := showdownCommand.Start(); err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if line == "Test your server at http://localhost:8000\n" {
			// Apparently this message doesn't mean server can be connected to.
			// Wait a bit more, perhaps it will accept connections.
			time.Sleep(time.Second)
			return
		}
	}
}

func finishShowdown() {
	showdownCommand.Process.Signal(os.Interrupt)
	showdownCommand.Wait()
}

func TestShowdownServer(t *testing.T) {
	runShowdown()

	config := ServerAddress{
		Host: "localhost",
		Port: 8000,
	}

	loginData := LoginData{Rooms: []string{"lobby"}}

	firstMessage := make(chan string)
	bc, _, err := ConnectToKnownServer(loginData, config, func(command, arg string, room *Room) {
		// First message from a server is updateuser message
		firstMessage <- command
	})
	assert.NoError(t, err, "Showdown client connection failed (may need to wait more?)")
	assert.Equal(t, <-firstMessage, "updateuser", "First message from server should be updateserver")
	bc.Close()

	finishShowdown()
}
