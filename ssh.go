package main

import (
	"context"
	"image/color"
	"io"
	"net"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
	"golang.org/x/crypto/ssh"
)

type remote struct {
	widget.BaseWidget

	terminal *terminal.Terminal

	session *ssh.Session

	win fyne.Window

	disconnected func()

	err error
}

var _ fyne.Widget = (*remote)(nil)
var _ io.Closer = (*remote)(nil)

func (r *router) NewSSH(win fyne.Window, dial func(ctx context.Context, network, address string) (net.Conn, error)) (*remote, error) {
	config := ssh.ClientConfig{
		User: r.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(r.password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := dial(context.Background(), "tcp", r.host+":22")
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, r.host+":22", &config)
	if err != nil {
		conn.Close()
		return nil, err
	}
	client := ssh.NewClient(c, chans, reqs)

	session, err := client.NewSession()
	if err != nil {
		c.Close()
		return nil, err
	}

	rssh := &remote{
		terminal: terminal.New(),
		session:  session,
		win:      win,
	}
	rssh.ExtendBaseWidget(rssh)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	cellSize := guessCellSize()
	if err := session.RequestPty("xterm-256color", int(rssh.Size().Height/cellSize.Height), int(rssh.Size().Width/cellSize.Width), modes); err != nil {
		_ = session.Close()
		return nil, err
	}

	in, _ := session.StdinPipe()
	out, _ := session.StdoutPipe()

	go session.Run("")

	go func() {
		rssh.err = rssh.terminal.RunWithConnection(in, out)

		if rssh.disconnected != nil {
			rssh.disconnected()
		}
	}()

	return rssh, nil
}

func (r *remote) OnDisconnected(f func()) {
	r.disconnected = f
}

func (r *remote) Tapped(_ *fyne.PointEvent) {
	r.win.Canvas().Focus(r.terminal)
}

func (r *remote) Resize(s fyne.Size) {
	cellSize := guessCellSize()
	r.err = r.session.WindowChange(int(s.Height/cellSize.Height), int(s.Width/cellSize.Width))
	r.terminal.Resize(s)
}

func (r *remote) Close() error {
	if r.session == nil {
		return nil
	}
	err := r.session.Close()
	r.session = nil
	return err
}

func (r *remote) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.terminal)
}

func guessCellSize() fyne.Size {
	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true

	return cell.MinSize()
}
