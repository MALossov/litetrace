package ui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/malossov/lite-tracer-mygo/internal/ftrace"
)

type Dashboard struct {
	app         *tview.Application
	engine      *ftrace.Engine
	statusView  *tview.TextView
	traceView   *tview.TextView
	ctx         context.Context
	cancel      context.CancelFunc
	traceBuffer []string
	bufferSize  int
	paused      bool
}

func NewDashboard(engine *ftrace.Engine) *Dashboard {
	ctx, cancel := context.WithCancel(context.Background())
	return &Dashboard{
		engine:      engine,
		ctx:         ctx,
		cancel:      cancel,
		traceBuffer: make([]string, 0, 1000),
		bufferSize:  0,
	}
}

func (d *Dashboard) Run() error {
	d.app = tview.NewApplication()

	// Status panel (top-left)
	d.statusView = tview.NewTextView()
	d.statusView.SetDynamicColors(true).SetScrollable(true).SetBorder(true).SetTitle(" Status ")

	// Live trace panel (top-right)
	d.traceView = tview.NewTextView()
	d.traceView.SetDynamicColors(true).SetScrollable(true).SetBorder(true).SetTitle(" Live Trace ")

	// Help/Controls panel (bottom)
	helpView := tview.NewTextView()
	helpView.SetDynamicColors(true).SetBorder(true).SetTitle(" Controls ")
	helpView.SetText(`[yellow]Keyboard Controls:[white]
  [green]Ctrl+C[white] or [green]q[white] - Quit and cleanup
  [green]p[white]            - Pause/Resume tracing
  [green]s[white]            - Save current trace to file
  [green]c[white]            - Clear trace buffer
  [green]↑/↓[white]          - Scroll trace view`)

	// Create top flex (status + trace)
	topFlex := tview.NewFlex().
		AddItem(d.statusView, 0, 1, false).
		AddItem(d.traceView, 0, 2, false)

	// Create main flex with help at bottom
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 0, 4, false).
		AddItem(helpView, 6, 0, false)

	d.app.SetRoot(mainFlex, true)

	// Keyboard handler
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				d.Shutdown()
				return nil
			case 'p', 'P':
				d.togglePause()
				return nil
			case 'c', 'C':
				d.clearBuffer()
				return nil
			case 's', 'S':
				d.saveTrace()
				return nil
			}
		case tcell.KeyCtrlC:
			d.Shutdown()
			return nil
		}
		return event
	})

	go d.updateStatus()
	go d.startTraceStream()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		d.Shutdown()
	}()

	if err := d.app.Run(); err != nil {
		return err
	}
	return nil
}

func (d *Dashboard) updateStatus() {
	ticker := NewTicker(2)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			status, err := d.engine.GetStatus()
			if err != nil {
				continue
			}

			statusText := fmt.Sprintf(`[yellow]Ftrace Kernel Subsystem Status[white]
=========================================
- Engine Status : %s
- Current Tracer: %s
- Active Filters: %s
- Buffer Size   : %d KB (Per CPU)
- Trace Clock   : %s
=========================================`,
				func() string {
					if status.Enabled {
						return "[green]RUNNING[white] (tracing_on = 1)"
					}
					return "[red]STOPPED[white] (tracing_on = 0)"
				}(),
				status.Tracer,
				status.Filter,
				status.BufferSize,
				status.TraceClock,
			)

			d.app.QueueUpdateDraw(func() {
				d.statusView.SetText(statusText)
			})
		}
	}
}

func (d *Dashboard) startTraceStream() {
	pipeChan := d.engine.ReadTracePipe(d.ctx)
	maxBufferSize := 5000

	for {
		select {
		case <-d.ctx.Done():
			return
		case line, ok := <-pipeChan:
			if !ok {
				return
			}
			d.app.QueueUpdateDraw(func() {
				d.traceBuffer = append(d.traceBuffer, line)
				d.bufferSize += len(line)
				
				if len(d.traceBuffer) > maxBufferSize {
					overflow := len(d.traceBuffer) - maxBufferSize
					d.traceBuffer = d.traceBuffer[overflow:]
					d.bufferSize = 0
					for _, s := range d.traceBuffer {
						d.bufferSize += len(s)
					}
					d.traceView.Clear()
				}
				
				fmt.Fprint(d.traceView, line)
			})
		}
	}
}

func (d *Dashboard) Shutdown() {
	d.cancel()
	d.engine.SafeShutdown()
	d.app.Stop()
	os.Exit(0)
}

func (d *Dashboard) togglePause() {
	d.paused = !d.paused
	if d.paused {
		d.engine.Disable()
		d.app.QueueUpdateDraw(func() {
			fmt.Fprint(d.traceView, "\n[yellow]*** TRACING PAUSED ***[white]\n\n")
		})
	} else {
		d.engine.Enable()
		d.app.QueueUpdateDraw(func() {
			fmt.Fprint(d.traceView, "\n[green]*** TRACING RESUMED ***[white]\n\n")
		})
	}
}

func (d *Dashboard) clearBuffer() {
	d.traceBuffer = d.traceBuffer[:0]
	d.bufferSize = 0
	d.app.QueueUpdateDraw(func() {
		d.traceView.Clear()
		fmt.Fprintln(d.traceView, "[yellow]*** BUFFER CLEARED ***[white]")
	})
}

func (d *Dashboard) saveTrace() {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("/tmp/litetrace_%s.txt", timestamp)
	
	content := strings.Join(d.traceBuffer, "")
	err := os.WriteFile(filename, []byte(content), 0644)
	
	d.app.QueueUpdateDraw(func() {
		if err != nil {
			fmt.Fprintf(d.traceView, "\n[red]*** Failed to save: %v ***[white]\n", err)
		} else {
			fmt.Fprintf(d.traceView, "\n[green]*** Trace saved to %s ***[white]\n", filename)
		}
	})
}

type Ticker struct {
	C     chan struct{}
	stop  chan struct{}
	timer *time.Timer
}

func NewTicker(intervalSec int) *Ticker {
	t := &Ticker{
		C:    make(chan struct{}),
		stop: make(chan struct{}),
	}
	d := time.Duration(intervalSec) * time.Second
	t.timer = time.NewTimer(d)
	go func() {
		for {
			select {
			case <-t.timer.C:
				select {
				case t.C <- struct{}{}:
				default:
				}
				t.timer.Reset(d)
			case <-t.stop:
				return
			}
		}
	}()
	return t
}

func (t *Ticker) Stop() {
	close(t.stop)
	t.timer.Stop()
}
