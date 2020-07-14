package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	_log "log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	serve   = flag.Bool("serve", false, "run searver")
	cmd     = flag.String("remote", "", "send remote command")
	sock    = flag.String("sock", "vimremote.sock", "set sock file")
	logFile = flag.String("log", "", "logging debug log to the file")

	log = _log.New(os.Stderr, "", 0)
)

func main() {
	flag.Parse()

	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		defer f.Close()

		log = _log.New(f, "", 0)
	}

	switch {
	case *serve:
		if err := runServer(*sock); err != nil {
			log.Println(err)
			os.Exit(2)
		}

	case len(*cmd) > 0:
		if err := runClient(*sock, *cmd, flag.Args()); err != nil {
			log.Println(err)
			os.Exit(2)
		}

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func runServer(sock string) error {
	log.Println("run server")

	if err := os.Remove(sock); err != nil && !os.IsNotExist(err) {
		return err
	}

	log.Printf("listening %s", sock)

	l, err := net.Listen("unix", sock)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	defer func() {
		if err := os.Remove(sock); err != nil && !os.IsNotExist(err) {
			log.Println(err)
		}
	}()

	var (
		c net.Conn
		s = make(chan os.Signal, 1)
	)

	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-s
		log.Println("closing")
		l.Close()
	}()

	for {
		c, err = l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			return fmt.Errorf("accept error: %w", err)
		}

		log.Println("accepted")

		conn, ok := c.(*net.UnixConn)
		if !ok {
			return fmt.Errorf("unexpected connection")
		}

		if _, err := io.Copy(os.Stdout, conn); err != nil {
			return err
		}

		log.Println("excecuted")

		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}
}

func runClient(sock, cmd string, args []string) error {
	src := []string{}
	hasResult := false

	switch cmd {
	case "redraw", "ex":
		src = append(src, cmd, strings.Join(args, " "))
	default:
		return fmt.Errorf("unexpected cmd: %s", cmd)
	}

	data, err := json.Marshal(&src)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(data)

	c, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}

	defer c.Close()

	conn, ok := c.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("unexpected connection")
	}

	if _, err := io.Copy(conn, b); err != nil {
		return err
	}

	if err := conn.CloseWrite(); err != nil {
		return err
	}

	if hasResult {
		if _, err := io.Copy(b, conn); err != nil {
			return err
		}

		fmt.Println(b.String())
	}

	return nil
}
