package ui

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type MessageType int

const (
	TypeLog MessageType = iota
	TypeProgress
	TypeLoader
	TypeSuccess
	TypeError
	TypeWarning
	TypeInfo
	TypeSnippet
	TypePackage
	TypeRealtime
	TypeSpinner
	TypeStopwatch
	TypeTable
)

type Element struct {
	elementType  MessageType
	content      string
	progress     float64
	progressText string
	showLoader   bool
	loaderText   string
	showSpinner  bool
	spinnerText  string
	startTime    time.Time
	tableData    [][]string
	tableHeaders []string
	status       string
	packageName  string
	action       string
}

var (
	instance *UI
	once     sync.Once
	spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

type UI struct {
	program     *tea.Program
	model       *uiModel
	enabled     bool
	updates     chan updateFunc
	closeCh     chan struct{}
	wg          sync.WaitGroup
	initialized bool
	ready       chan struct{}
}

func Init() {
	once.Do(func() {
		model := &uiModel{}
		ui := &UI{
			model:   model,
			enabled: true,
			updates: make(chan updateFunc, 100),
			closeCh: make(chan struct{}),
			ready:   make(chan struct{}),
		}

		go func() {
			ui.program = tea.NewProgram(
				model,
				tea.WithoutSignalHandler(),
				tea.WithInput(nil),
			)
			close(ui.ready)
			if _, err := ui.program.Run(); err != nil {
				fmt.Println("UI error:", err)
			}
		}()

		ui.wg.Add(1)
		go ui.processUpdates()

		instance = ui
		instance.initialized = true
	})
}

func Stop() {
	if instance != nil && instance.initialized {
		close(instance.closeCh)
		instance.wg.Wait()
		instance.enabled = false
		instance.initialized = false
		instance = nil
	}
}

func StopWithTimeout(timeout time.Duration) {
	if instance != nil && instance.initialized {
		time.AfterFunc(timeout, func() {
			Stop()
		})
	}
}

func IsEnabled() bool {
	return instance != nil && instance.enabled
}

func (ui *UI) queueUpdate(fn updateFunc) {
	select {
	case ui.updates <- fn:
	default:
	}
}

func (ui *UI) processUpdates() {
	defer ui.wg.Done()

	<-ui.ready

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case update := <-ui.updates:
			ui.model.mu.Lock()
			update(ui.model)
			ui.model.mu.Unlock()
			ui.requestRender()

		case <-ticker.C:
			ui.requestRender()

		case <-ui.closeCh:
			return
		}
	}
}

func (ui *UI) requestRender() {
	if ui.program != nil {
		go func() {
			ui.program.Send(forceRefresh{})
		}()
	}
}
