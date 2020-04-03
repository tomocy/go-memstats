package memstats

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Run(gen func() Window, loader Loader) error {
	if err := termui.Init(); err != nil {
		return fmt.Errorf("failed to init terminal ui: %w", err)
	}
	defer termui.Close()

	win := gen()
	events := termui.PollEvents()
	ticks := time.Tick(time.Second)
	for {
		select {
		case <-ticks:
			stats, err := loader.Load(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load memory statics: %w", err)
			}

			win.Render(stats)
		case e := <-events:
			switch e.Type {
			case termui.KeyboardEvent:
				return nil
			case termui.ResizeEvent:
				win.Resize()
			}
		}
	}
}

type Window interface {
	Render(*runtime.MemStats)
	Resize()
}

func NewGrid() *Grid {
	g := &Grid{
		Grid:   termui.NewGrid(),
		widget: newWidget(),
	}
	g.init()

	return g
}

type Grid struct {
	*termui.Grid
	widget *widget
}

func (g *Grid) init() {
	g.resize()
	g.Set(
		termui.NewRow(1, g.widget.gcCPUFraction),
	)
}

func (g *Grid) Render(stats *runtime.MemStats) {
	g.widget.render(stats)
	termui.Render(g)
}

func (g *Grid) Resize() {
	g.resize()
	termui.Render(g)
}

func (g *Grid) resize() {
	w, h := termui.TerminalDimensions()
	g.SetRect(0, 0, w, h)
}

func newWidget() *widget {
	w := &widget{
		gcCPUFraction: widgets.NewGauge(),
	}
	w.init()

	return w
}

type widget struct {
	gcCPUFraction *widgets.Gauge
}

func (w *widget) init() {
	w.gcCPUFraction.Title = "GCCPUFraction 0%~100%"
	w.gcCPUFraction.BarColor = termui.Color(50)
}

func (w *widget) render(stat *runtime.MemStats) {
	w.updateGCCPUFraction(stat.GCCPUFraction)
}

func (w *widget) updateGCCPUFraction(f float64) {
	w.gcCPUFraction.Percent = int(f * 100)
	w.gcCPUFraction.Label = fmt.Sprintf("%.2f%%", f*100)
}

type Loader interface {
	Load(context.Context) (*runtime.MemStats, error)
}

type viaHTTP struct {
	*http.Client
	url string
}

func (l *viaHTTP) Load(ctx context.Context) (*runtime.MemStats, error) {
	resp, err := l.Get(l.url)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory statics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get memory statics successfully: %w", fmt.Errorf(resp.Status))
	}

	var loaded struct {
		Stats *runtime.MemStats `json:"memstats"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loaded); err != nil {
		return nil, fmt.Errorf("failed to decode body: %w", err)
	}

	return loaded.Stats, nil
}

type RandomLoader struct{}

func (l *RandomLoader) Load(context.Context) (*runtime.MemStats, error) {
	return &runtime.MemStats{
		GCCPUFraction: rand.Float64(),
	}, nil
}
