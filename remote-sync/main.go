package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fogoid/remote-sync/config"
	"github.com/neovim/go-client/nvim"
)

func main() {
	// Turn off timestamps in output.
	log.SetFlags(0)

	// Direct writes by the application to stdout garble the RPC stream.
	// Redirect the application's direct use of stdout to stderr.
	stdout := os.Stdout
	os.Stdout = os.Stderr

	// Create a client connected to stdio. Configure the client to use the
	// standard log package for logging.
	v, err := nvim.New(os.Stdin, stdout, stdout, log.Printf)
	if err != nil {
		log.Fatal(err)
	}

	// Register function with the client.
	v.RegisterHandler("SelectConnection", SelectConnection)

	// Run the RPC message loop. The Serve function returns when
	// nvim closes.
	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
}

func SelectConnection(v *nvim.Nvim, args []string) error {
	log.Println("HELLO WORLD")

	if len(config.Conf.Connections) == 0 {
		return v.WriteOut("No existing connections. Please add connections to .sync.json file and reload them")
	}

	// Create slice and map with names and positions, respectively.
	// Will be later used for choices and for connection to use
	csName := make([]string, 0, len(config.Conf.Connections))
	choices := make(map[string]int, len(config.Conf.Connections))
	for i, c := range config.Conf.Connections {
		if c.Name == "" {
			err := v.WriteOut("WARN: Name is not present. Using host:port as name")
			if err != nil {
				return err
			}

			c.Name = fmt.Sprintf("%s:%s", c.Host, c.Port)
		}

		luaName := fmt.Sprintf("'%s'", c.Name)
		csName = append(csName, luaName)
		choices[c.Name] = i
	}

	luaChoices := strings.Join(csName, ",")
	selectCode := `
        vim.ui.select(
            {%s},
            {
                prompt = 'Select connection to use when syncing workspace',
                format_item  = function(item)
                    return "Connection: " .. item
                end,
            },
            function(choice) 
            end
        )
    `
	selectCode = fmt.Sprintf(selectCode, luaChoices)

	err := v.ExecLua(selectCode, struct{}{})
	if err != nil {
		return v.WriteOut("ERROR: Executing select block")
	}

	var userChoice string
	v.Var("remote_sync.connection.choice", &userChoice)
	return v.WriteOut(fmt.Sprintf("INFO: connection %s will be used for further syncs", userChoice))
}
