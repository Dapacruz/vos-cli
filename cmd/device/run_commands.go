//go:build !windows
// +build !windows

package device

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	expect "github.com/google/goexpect"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

var (
	cmds          []string
	port          string
	keyBasedAuth  bool
	passwordStdin bool
	promptRE      = regexp.MustCompile(`#$`)
	signer        ssh.Signer
	ignoreHostKey bool
	sortOutput    bool
	expectTimeout int
	sshTimeout    int
)

const SESSION_SETUP = ""

type sessionDetails struct {
	host    string
	results map[string]string
}

// runCommandsCmd represents the runCommands command
var runCommandsCmd = &cobra.Command{
	Use:   "commands [flags] <device> [device]...",
	Short: "Executes CLI commands via SSH",
	Long: `Executes CLI commands via SSH

Examples:
  # Execute the 'show software' and 'show ip routes' commands on hub01.example.com:

    > vos-cli device run commands --command "show software","show ip routes" hub01.example.com

  # Execute the 'show software' command on hub01.example.com and hub02.example.com, use key based auth, and ignore host key verification:

    > vos-cli device run commands --command "show software" --key-based-auth --insecure hub01.example.com hub02.example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// If the cmds flag is not set, exit and display usage
		fmt.Fprintln(os.Stderr)
		if len(cmds) == 0 {
			cmd.Help()
			fmt.Printf("\nno commands specified\n")
			os.Exit(1)
		}

		hosts = cmd.Flags().Args()
		if len(hosts) == 0 {
			if isInputFromPipe() {
				if !keyBasedAuth && viper.GetString("password") == "" && password == "" {
					log.Fatal("key based auth, saved password or password flag is required when reading hosts from stdin")
				}

				// Read hosts from stdin
				scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
				for scanner.Scan() {
					hosts = append(hosts, scanner.Text())
				}
			} else {
				cmd.Help()
				fmt.Printf("\nno hosts specified\n")
				os.Exit(1)
			}
		}

		// Retrieve password from standard input
		if passwordStdin {
			if isInputFromPipe() {
				// Read password from stdin
				scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
				for scanner.Scan() {
					password = scanner.Text()
				}
			} else {
				cmd.Help()
				fmt.Printf("\nunable to retrieve password from standard input\n")
				os.Exit(1)
			}
		}

		if !keyBasedAuth && viper.GetString("user") == "" && user == "" {
			fmt.Fprint(os.Stderr, "Cisco User: ")
			fmt.Scanln(&user)
		} else if user == "" {
			user = viper.GetString("user")
		}

		// If the password flag is not set, prompt for password
		if !keyBasedAuth && viper.GetString("password") == "" && password == "" {
			fmt.Fprintf(os.Stderr, "Password (%s): ", user)
			bytepw, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				panic(err)
			}
			password = string(bytepw)
			fmt.Fprintf(os.Stderr, "\n\n")
		} else if password == "" {
			password = viper.GetString("password")
		}

		if keyBasedAuth {
			file, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
			if err != nil {
				log.Fatal(err)
			}

			signer, err = ssh.ParsePrivateKey(file)
			if err != nil {
				log.Fatal(err)
			}
		}

		start := time.Now()

		ch := make(chan sessionDetails, 100)
		doneCh := make(chan struct{})

		go printCmdResults(ch, doneCh)

		for _, host := range hosts {
			wg.Add(1)
			go runCommands(ch, host)
		}
		wg.Wait()
		close(ch)
		<-doneCh

		elapsed := time.Since(start)

		fmt.Printf(" Complete: %d command(s) executed on %d host(s) in %.3f seconds\n", len(cmds), len(hosts), elapsed.Seconds())
	},
}

func init() {
	runCmd.AddCommand(runCommandsCmd)

	runCommandsCmd.Flags().StringVar(&user, "user", user, "Cisco admin user")
	runCommandsCmd.Flags().StringVar(&password, "password", password, "password for Cisco user")
	runCommandsCmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "receive password from standard input")
	runCommandsCmd.Flags().StringSliceVarP(&cmds, "command", "c", cmds, "comma separated set of commands to execute")
	runCommandsCmd.Flags().BoolVarP(&keyBasedAuth, "key-based-auth", "k", false, "use key-based authentication")
	runCommandsCmd.Flags().StringVarP(&port, "port", "p", "22", "port to connect to on host")
	runCommandsCmd.Flags().IntVarP(&expectTimeout, "expect-timeout", "e", 30, "expect timeout in seconds for each command")
	runCommandsCmd.Flags().IntVarP(&sshTimeout, "ssh-timeout", "S", 30, "expect timeout in seconds for each command")
	runCommandsCmd.Flags().BoolVarP(&ignoreHostKey, "insecure", "K", false, "ignore host key checking")
	runCommandsCmd.Flags().BoolVarP(&sortOutput, "sort-output", "s", false, "sort command output")
}

// runCommands executes commands on a host
func runCommands(ch chan<- sessionDetails, host string) {
	defer wg.Done()

	session := sessionDetails{
		host:    host,
		results: make(map[string]string),
	}
	sessionTimeout := time.Duration(expectTimeout) * time.Second

	// Set auth method
	var authMethod ssh.AuthMethod
	if keyBasedAuth {
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(password)
	}

	// Set host key callback
	var hostkeyCallback ssh.HostKeyCallback
	if ignoreHostKey {
		// allow any host key to be used (non-prod)
		hostkeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		var err error
		hostkeyCallback, err = knownhosts.New(path.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
		if err != nil {
			log.Fatalf("unable to load ssh known_hosts: %v", err)
		}
	}

	// Connect to host
	sshClt, err := ssh.Dial("tcp", net.JoinHostPort(host, port), &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: hostkeyCallback,
		Timeout:         time.Duration(sshTimeout) * time.Second,
		Config:          ssh.Config{Ciphers: []string{"aes256-ctr", "aes192-ctr", "aes128-ctr", "aes256-cbc", "aes192-cbc", "aes128-cbc"}},
	})
	if err != nil {
		log.Fatalf("ssh.Dial(%q) failed: %v", host, err)
	}
	defer sshClt.Close()

	// Start SSH session
	e, _, err := expect.SpawnSSH(sshClt, sessionTimeout)
	if err != nil {
		log.Fatalf("%s: %v", host, err)
	}
	defer e.Close()

	// Wait for prompt after login
	_, _, err = e.Expect(promptRE, sessionTimeout)
	if err != nil {
		log.Fatalf("%s: %v", host, err)
	}

	// Set up session
	err = e.Send(SESSION_SETUP + "\n")
	if err != nil {
		log.Fatal(err)
	}
	_, _, err = e.Expect(promptRE, sessionTimeout)
	if err != nil {
		log.Fatalf("%s: %v", host, err)
	}

	// Execute commands
	reConnectCmd := regexp.MustCompile(`^(connect|telnet)`)
	reConfigSaveCmd := regexp.MustCompile(`^(do\S* )?cop\S* ru\S* st\S*`)
	for _, cmd := range cmds {
		err = e.Send(cmd + "\n")
		if err != nil {
			log.Fatal(err)
		}

		if reConnectCmd.MatchString(cmd) {
			// Send keys to terminate session
			err = e.Send("\n\b\b\b\b\b\n")
			if err != nil {
				log.Fatal(err)
			}
		} else if reConfigSaveCmd.MatchString(cmd) {
			// Confirm saved configuration destination
			err = e.Send("\n")
			if err != nil {
				log.Fatal(err)
			}
		}

		result, _, err := e.Expect(promptRE, sessionTimeout)
		if err != nil {
			log.Fatalf("%s: %v", host, err)
		}

		session.results[cmd] = result
	}

	ch <- session
}

// printCmdResults prints the results
func printCmdResults(ch <-chan sessionDetails, doneCh chan<- struct{}) {
	for {
		if session, chanIsOpen := <-ch; chanIsOpen {
			green.Printf("\n*** %s ***\n\n\n", session.host)

			// Sort results by command
			if sortOutput {
				sort.Strings(cmds)
			}

			for _, cmd := range cmds {
				yellow.Printf("*** %s ***\n", cmd)
				fmt.Printf("%s\n\n", trimOutput(session.results[cmd]))
			}
			blue.Printf("################################################################################\n\n")
		} else {
			doneCh <- struct{}{}
			return
		}
	}
}

// trimOutput removes the echoed command and the prompt from the output
func trimOutput(output string) string {
	lines := strings.Split(output, "\n")
	return strings.Join(lines[1:len(lines)-1], "\n")
}
