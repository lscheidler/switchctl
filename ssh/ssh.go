package ssh

import (
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Ssh struct {
	hostname string
	username string
	port     string

	sshc *ssh.Client
}

func New(hostname string, username string, port string) *Ssh {
	return &Ssh{
		hostname: hostname,
		username: username,
		port:     port,
		sshc:     nil,
	}
}

func (s *Ssh) Connect() error {
	// ssh-agent(1) provides a UNIX socket at $SSH_AUTH_SOCK.
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		//log.Fatalf("Failed to open SSH_AUTH_SOCK: %v", err)
		return err
	}

	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: s.username,
		Auth: []ssh.AuthMethod{
			// Use a callback rather than PublicKeys so we only consult the
			// agent once the remote server wants it.
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if sshc, err := ssh.Dial("tcp", s.hostname+":"+s.port, config); err != nil {
		//log.Println("Failed to create connection: ", err)
		//log.Fatal("Failed to create connection: ", err)
		return err
	} else {
		s.sshc = sshc
	}
	// Use sshc...
	//defer sshc.Close()

	return nil
}

func (s *Ssh) Close() {
	if s.sshc != nil {
		defer s.sshc.Close()
	}
}

func (s *Ssh) Execute(command string, stdout io.Writer, stderr io.Writer) error {
	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	if session, err := s.sshc.NewSession(); err != nil {
		log.Println("Failed to create session: ", err)
		return err
	} else {
		defer session.Close()

		// Once a Session is created, you can execute a single command on
		// the remote side using the Run method.
		session.Stdout = stdout
		session.Stderr = stderr
		if err := session.Run(command); err != nil {
			//log.Fatal("Failed to run: " + err.Error())
			return err
		}
		return nil
	}
}
