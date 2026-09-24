package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"sync"
	"time"

	"p2pChat/config"
	"p2pChat/models"
	peer "p2pChat/peers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type pendingReq struct {
	Addr string
	ID   string
}

type appState struct {
	node *models.Node
	a    fyne.App

	mu         sync.Mutex
	discovered []models.DiscoveredPeer
	pending    []pendingReq
	connected  []string

	discoveredList *widget.List
	connectedList  *widget.List

	chatWindows map[string]fyne.Window
	chatLogs    map[string]*widget.Entry
}

func main() {
	os.Stderr.WriteString("main started\n")
	if len(os.Args) < 2 {
		fmt.Println("Usage: main.exe <listen-port>")
		return
	}
	listenPort, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid listening port:", err)
		return
	}

	a := app.New()
	setupWin := a.NewWindow("p2pChat - setup")

	state := &appState{
		a:           a,
		chatWindows: make(map[string]fyne.Window),
		chatLogs:    make(map[string]*widget.Entry),
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your display name")

	form := &widget.Form{Items: []*widget.FormItem{{Text: "Name", Widget: nameEntry}}}
	form.OnSubmit = func() {
		if nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("name cannot be empty"), setupWin)
			return
		}
		node, err := config.NewNode(nameEntry.Text, listenPort)
		if err != nil {
			dialog.ShowError(err, setupWin)
			return
		}
		state.node = node
		wireHooks(state)
		config.StartServices(node)

		setupWin.Close()
		buildMainWindow(state)
	}
	form.SubmitText = "Start"

	setupWin.SetContent(container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Listening on port %d", listenPort)),
		form,
	))
	setupWin.Resize(fyne.NewSize(320, 140))
	fmt.Println("about to show setup window")
	setupWin.Show()
	fmt.Println("setup window shown")

	a.Run()
}

func wireHooks(state *appState) {
	n := state.node

	n.OnPeerDiscovered = func(p models.DiscoveredPeer) {
		state.mu.Lock()
		state.discovered = append(state.discovered, p)
		state.mu.Unlock()
		fyne.Do(func() {
			if state.discoveredList != nil {
				state.discoveredList.Refresh()
			}
		})
	}

	n.OnConnectionRequest = func(addr, observedID string) {
		state.mu.Lock()
		state.pending = append(state.pending, pendingReq{Addr: addr, ID: observedID})
		state.mu.Unlock()
		fyne.Do(func() {
			if state.discoveredList != nil {
				state.discoveredList.Refresh()
			}
		})
	}

	n.OnPeerConnected = func(addr string) {
		state.mu.Lock()
		state.connected = append(state.connected, addr)
		filtered := state.pending[:0]
		for _, p := range state.pending {
			if p.Addr != addr {
				filtered = append(filtered, p)
			}
		}
		state.pending = filtered
		state.mu.Unlock()
		fyne.Do(func() {
			if state.connectedList != nil {
				state.connectedList.Refresh()
			}
			if state.discoveredList != nil {
				state.discoveredList.Refresh()
			}
		})
	}

	n.OnPeerDisconnected = func(addr string) {
		state.mu.Lock()
		filtered := state.connected[:0]
		for _, a := range state.connected {
			if a != addr {
				filtered = append(filtered, a)
			}
		}
		state.connected = filtered
		state.mu.Unlock()
		fyne.Do(func() {
			if state.connectedList != nil {
				state.connectedList.Refresh()
			}
		})
	}

	n.OnMessage = func(addr string, msg models.Message) {
		fyne.Do(func() {
			log, ok := state.chatLogs[addr]
			if !ok {
				return
			}
			log.SetText(log.Text + fmt.Sprintf("[%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.From, msg.Body))
		})
	}
}

func buildMainWindow(state *appState) {
	w := state.a.NewWindow(fmt.Sprintf("p2pChat - %s", state.node.Name))

	state.discoveredList = widget.NewList(
		func() int {
			state.mu.Lock()
			defer state.mu.Unlock()
			return len(state.pending) + len(state.discovered)
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			lbl := widget.NewLabel("template")
			return container.NewStack(bg, lbl)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			bg := c.Objects[0].(*canvas.Rectangle)
			lbl := c.Objects[1].(*widget.Label)

			state.mu.Lock()
			defer state.mu.Unlock()
			if id < len(state.pending) {
				p := state.pending[id]
				lbl.SetText(fmt.Sprintf("Connection request: %s", p.ID))
				bg.FillColor = color.NRGBA{R: 0, G: 180, B: 0, A: 70}
			} else {
				d := state.discovered[id-len(state.pending)]
				lbl.SetText(fmt.Sprintf("%s (%s)", d.Name, d.Addr))
				bg.FillColor = color.Transparent
			}
			bg.Refresh()
		},
	)

	state.discoveredList.OnSelected = func(id widget.ListItemID) {
		state.mu.Lock()
		var isPending bool
		var addr, obsID, name string
		if id < len(state.pending) {
			isPending = true
			addr = state.pending[id].Addr
			obsID = state.pending[id].ID
		} else {
			d := state.discovered[id-len(state.pending)]
			addr, obsID, name = d.Addr, d.ID, d.Name
		}
		state.mu.Unlock()
		state.discoveredList.UnselectAll()

		if isPending {
			showAcceptRejectDialog(state, w, addr, obsID)
		} else {
			showConnectDialog(state, w, addr, obsID, name)
		}
	}

	state.connectedList = widget.NewList(
		func() int {
			state.mu.Lock()
			defer state.mu.Unlock()
			return len(state.connected)
		},
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			state.mu.Lock()
			defer state.mu.Unlock()
			obj.(*widget.Label).SetText(state.connected[id])
		},
	)

	state.connectedList.OnSelected = func(id widget.ListItemID) {
		state.mu.Lock()
		addr := state.connected[id]
		state.mu.Unlock()
		state.connectedList.UnselectAll()
		openChatWindow(state, addr)
	}

	left := container.NewBorder(widget.NewLabel("Discovered / requests"), nil, nil, nil, state.discoveredList)
	right := container.NewBorder(widget.NewLabel("Connections"), nil, nil, nil, state.connectedList)

	split := container.NewHSplit(left, right)
	split.Offset = 0.5

	w.SetContent(split)
	w.Resize(fyne.NewSize(700, 450))
	w.Show()
}

func showAcceptRejectDialog(state *appState, w fyne.Window, addr, obsID string) {
	msg := fmt.Sprintf(
		"Incoming connection request\n\nFrom address: %s\nTheir fingerprint: %s\n\nYour fingerprint: %s\n\nConfirm this matches what the other person sees before accepting.",
		addr, obsID, state.node.NodeId,
	)
	dialog.ShowConfirm("Connection request", msg, func(accept bool) {
		if accept {
			acceptRequest(state.node, addr)
			return
		}
		rejectRequest(state.node, addr)
		state.mu.Lock()
		filtered := state.pending[:0]
		for _, p := range state.pending {
			if p.Addr != addr {
				filtered = append(filtered, p)
			}
		}
		state.pending = filtered
		state.mu.Unlock()
		state.discoveredList.Refresh()
	}, w)
}

func showConnectDialog(state *appState, w fyne.Window, addr, obsID, name string) {
	msg := fmt.Sprintf(
		"Peer: %s\nAddress: %s\n\nYour fingerprint: %s\nPeer fingerprint: %s\n\nConfirm this matches what the other person sees before connecting.",
		name, addr, state.node.NodeId, obsID,
	)
	dialog.ShowConfirm("Connect to peer?", msg, func(ok bool) {
		if !ok {
			return
		}
		if err := connectToDiscovered(state.node, addr, obsID); err != nil {
			dialog.ShowError(err, w)
		}
	}, w)
}

func openChatWindow(state *appState, addr string) {
	if win, ok := state.chatWindows[addr]; ok {
		win.RequestFocus()
		return
	}

	win := state.a.NewWindow("Chat - " + addr)

	log := widget.NewMultiLineEntry()
	log.Disable()
	logScroll := container.NewVScroll(log)

	input := widget.NewEntry()
	input.SetPlaceHolder("Type a message...")

	sendFunc := func() {
		text := input.Text
		if text == "" {
			return
		}
		if err := state.node.SendTo(addr, text); err != nil {
			dialog.ShowError(err, win)
			return
		}
		log.SetText(log.Text + fmt.Sprintf("[%s] Me: %s\n", time.Now().Format("15:04:05"), text))
		input.SetText("")
	}
	input.OnSubmitted = func(string) { sendFunc() }
	send := widget.NewButton("Send", sendFunc)

	content := container.NewBorder(nil, container.NewBorder(nil, nil, nil, send, input), nil, nil, logScroll)
	win.SetContent(content)
	win.Resize(fyne.NewSize(400, 500))

	state.chatWindows[addr] = win
	state.chatLogs[addr] = log
	win.SetOnClosed(func() {
		delete(state.chatWindows, addr)
		delete(state.chatLogs, addr)
	})

	win.Show()
}

func acceptRequest(node *models.Node, addr string) {
	conn, ok := node.AcceptPending(addr)
	if ok {
		go peer.ReadFromPeer(addr, conn, node)
	}
}

func rejectRequest(node *models.Node, addr string) {
	node.RejectPending(addr)
}

func connectToDiscovered(node *models.Node, addr, id string) error {
	conn, err := peer.ConnectToPeer(addr, node, id)
	if err != nil {
		return err
	}
	if node.AddPeer(addr, conn) {
		node.RememberPeer(addr, id)
		go peer.ReadFromPeer(addr, conn, node)
	}
	return nil
}
