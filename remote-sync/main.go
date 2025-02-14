package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/fogoid/remote-sync/config"
	"github.com/neovim/go-client/nvim"
	"golang.org/x/crypto/ssh"
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

	if err := config.ReadConfig(); err != nil {
		log.Fatal(err)
	}

	// Register function with the client.
	v.RegisterHandler("selectConnection", SelectConnection)
	v.RegisterHandler("sayHello", CreateAndSendHello)

	// Run the RPC message loop. The Serve function returns when
	// nvim closes.
	if err := v.Serve(); err != nil {
		log.Fatal(err)
	}
}

func SelectConnection(v *nvim.Nvim, args []string) error {
	if len(config.Conf) == 0 {
		return v.WriteOut("No existing connections. Please add connections to .sync.json file and reload them\n")
	}

	// Create slice and map with names and positions, respectively.
	// Will be later used for choices and for connection to use
	csName := make([]string, 0, len(config.Conf))
	choices := make(map[string]int, len(config.Conf))
	for i, c := range config.Conf {
		if c.Name == "" {
			err := v.WriteOut("WARN: Name is not present. Using host:port as name\n")
			if err != nil {
				return err
			}

			c.Name = fmt.Sprintf("%s:%d", c.Host, c.Port)
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
                vim.g.rs_connection = choice
	        end
	    )
	`
	selectCode = fmt.Sprintf(selectCode, luaChoices)

	err := v.ExecLua(selectCode, nil)
	if err != nil {
		return v.WriteOut("ERROR: Executing select block\n")
	}

	var userChoice string
	if err = v.Var("rs_connection", &userChoice); err != nil {
		return v.WriteOut("ERROR: getting user selected connection\n")
	}

	config.CurrentConnection = choices[userChoice]
	return v.WriteOut(fmt.Sprintf("INFO: connection %s will be used for further syncs\n", userChoice))
}

func CreateAndSendHello(v *nvim.Nvim, args []string) error {
	if config.CurrentConnection == -1 {
		return v.WriteOut("No connection selected. Please select first a connection using the SelectConnection command\n")
	}

	ipConn := net.IPConn{}
	authMethod := ssh.Password(config.Conf[config.CurrentConnection].Password)
	clientConfig := ssh.ClientConfig{
		User: config.Conf[config.CurrentConnection].Username,
		Auth: []ssh.AuthMethod{authMethod},
	}
	address := fmt.Sprintf("%s:%d", config.Conf[config.CurrentConnection].Host, config.Conf[config.CurrentConnection].Port)
	sshConn, newChan, reqChan, err := ssh.NewClientConn(&ipConn, address, &clientConfig)
	if err != nil {
		return v.WriteOut(fmt.Sprintf("ERROR: Creating client connection: %w", err))
	}

	return nil
}
