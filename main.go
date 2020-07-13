package main

import (
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/process"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"image/color"
	"os"
	"time"
)

var (
	help     bool
	version  bool
	pid      int64
	duration time.Duration
	interval time.Duration
	width    string
	height   string
	output   string
)

func init() {
	flag.BoolVar(&help, "help", false, "help info")
	flag.BoolVar(&version, "v", false, "version info")
	flag.Int64Var(&pid, "p", -1, "pid to get info from")
	flag.DurationVar(&duration, "d", 10*time.Second, "sampling duration (0 means sample until pid exits)")
	flag.DurationVar(&interval, "i", 50*time.Millisecond, "sampling interval")
	flag.StringVar(&width, "w", "10cm", "output image width (cm or inch)")
	flag.StringVar(&height, "h", "8cm", "output image height (cm or inch)")
	flag.StringVar(&output, "o", "pinfo.png", "output image file path")
	flag.Usage = usage
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `pinfoplot tool in golang to generate process info image
Version: 0.0.1
Usage: pinfoplot [-help help] [-v version] [-p pid] [-d sampling duration] [-i sampling interval] [-w output image width (cm or inch)] [-h output image height (cm or inch)] [-o output image file path]
Options
`)
	flag.PrintDefaults()
}

type ProcessInfo struct {
	Pid              int32
	StartTime        time.Time
	SamplingInterval time.Duration
	Samples          []sample
}

func New(pid int32, duration, interval time.Duration) (*ProcessInfo, error) {
	sampleNo := duration / interval
	if sampleNo < 2 {
		return nil, fmt.Errorf("need at least 2 samples, your sampling interval is too long or sampling duration is too short")
	}

	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	pi := &ProcessInfo{
		Pid:              pid,
		StartTime:        start,
		SamplingInterval: interval,
		Samples:          make([]sample, 0, sampleNo),
	}

	running, err := proc.IsRunning()
	if err != nil {
		return nil, err
	}

	// duration == 0 means sampling until pid exits
	for t := time.Since(start); duration == 0 || t <= duration && running; t = time.Since(start) {
		mem, err := proc.MemoryInfo()
		if err != nil {
			return nil, err
		}
		io, err := proc.IOCounters()
		if err != nil {
			return nil, err
		}
		c, err := proc.CPUPercent()
		if err != nil {
			return nil, err
		}

		sample := sample{
			mem:      mem,
			io:       io,
			cpu:      c,
			interval: t,
		}

		pi.Samples = append(pi.Samples, sample)
		time.Sleep(interval)
		running, err = proc.IsRunning()
		if err != nil {
			return nil, err
		}
	}
	return pi, nil
}

func (pi *ProcessInfo) MemPlot() (*plot.Plot, error) {
	pl, err := plot.New()
	if err != nil {
		return nil, err
	}
	pl.Title.Text = fmt.Sprintf("Memory Plot of PID %d", pi.Pid)
	pl.X.Label.Text = "t (s)"
	pl.Y.Label.Text = "KB"
	pl.Add(plotter.NewGrid())
	// RSS
	ptsRss := make(plotter.XYs, len(pi.Samples))
	// VMS
	ptsVms := make(plotter.XYs, len(pi.Samples))
	for i, s := range pi.Samples {
		ptsRss[i].X = s.interval.Seconds()
		ptsVms[i].X = s.interval.Seconds()
		ptsRss[i].Y = float64(pi.Samples[i].mem.RSS) / 1024
		ptsVms[i].Y = float64(pi.Samples[i].mem.VMS) / 1024
	}
	// RSS
	lineRss, err := plotter.NewLine(ptsRss)
	if err != nil {
		return nil, err
	}
	lineRss.LineStyle.Width = vg.Points(1)
	lineRss.LineStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	pl.Add(lineRss)
	pl.Legend.Add("RSS", lineRss)
	// VMS
	lineVms, err := plotter.NewLine(ptsVms)
	if err != nil {
		return nil, err
	}
	lineVms.LineStyle.Width = vg.Points(1)
	lineVms.LineStyle.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	pl.Add(lineVms)
	pl.Legend.Add("VMS", lineVms)
	return pl, nil
}

func (pi *ProcessInfo) IoPlot() (*plot.Plot, error) {
	pl, err := plot.New()
	if err != nil {
		return nil, err
	}
	pl.Title.Text = fmt.Sprintf("IO Plot of PID %d", pi.Pid)
	pl.X.Label.Text = "t (s)"
	pl.Y.Label.Text = "op"
	pl.Add(plotter.NewGrid())
	// read
	ptsR := make(plotter.XYs, len(pi.Samples))
	// write
	ptsW := make(plotter.XYs, len(pi.Samples))
	for i, s := range pi.Samples {
		ptsR[i].X = s.interval.Seconds()
		ptsW[i].X = s.interval.Seconds()
		ptsR[i].Y = float64(pi.Samples[i].io.ReadCount)
		ptsW[i].Y = float64(pi.Samples[i].io.WriteCount)
	}
	// read
	lineR, err := plotter.NewLine(ptsR)
	if err != nil {
		return nil, err
	}
	lineR.LineStyle.Width = vg.Points(1)
	lineR.LineStyle.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	pl.Add(lineR)
	pl.Legend.Add("IO Read", lineR)
	// write
	lineW, err := plotter.NewLine(ptsR)
	if err != nil {
		return nil, err
	}
	lineW.LineStyle.Width = vg.Points(1)
	lineW.LineStyle.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	pl.Add(lineW)
	pl.Legend.Add("IO Write", lineW)
	return pl, nil
}

func (pi *ProcessInfo) CpuPlot() (*plot.Plot, error) {
	pl, err := plot.New()
	if err != nil {
		return nil, err
	}
	pl.Title.Text = fmt.Sprintf("CPU Plot of PID %d", pi.Pid)
	pl.X.Label.Text = "t (s)"
	pl.Y.Label.Text = "%"
	pl.Add(plotter.NewGrid())
	pts := make(plotter.XYs, len(pi.Samples))
	for i, s := range pi.Samples {
		pts[i].X = s.interval.Seconds()
		pts[i].Y = pi.Samples[i].cpu * 100
	}
	line, err := plotter.NewLine(pts)
	if err != nil {
		return nil, err
	}
	line.LineStyle.Width = vg.Points(1)
	line.LineStyle.Color = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	pl.Add(line)
	pl.Legend.Add("CPU", line)

	return pl, nil
}

func (pi *ProcessInfo) SavePlot(pl *plot.Plot, width, height vg.Length, filename string) error {
	return pl.Save(width, height, filename)
}

func (pi *ProcessInfo) Save(plots [][]*plot.Plot, width, height vg.Length, filename string) error {
	rows := len(plots)
	cols := len(plots[0])
	img := vgimg.New(width, height)
	dc := draw.New(img)
	t := draw.Tiles{
		Rows: rows,
		Cols: cols,
	}
	canvases := plot.Align(plots, t, dc)
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			if plots[j][i] != nil {
				plots[j][i].Draw(canvases[j][i])
			}
		}
	}
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(w); err != nil {
		return err
	}
	return nil
}

type sample struct {
	mem      *process.MemoryInfoStat
	io       *process.IOCountersStat
	cpu      float64
	interval time.Duration
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	if help {
		flag.Usage()
	} else if version {
		fmt.Println("version: 0.0.1")
	} else if pid <= 0 {
		fmt.Println("invalid pid")
	} else {
		imgWidth, err := vg.ParseLength(width)
		check(err)
		imgHeight, err := vg.ParseLength(height)
		check(err)
		fmt.Println("Collecting info from pid:", pid)
		if duration == 0 {
			fmt.Println("Your sampling duration is 0, which means sample pid until it exits")
		}
		pi, err := New(int32(pid), duration, interval)
		check(err)
		memPlot, err := pi.MemPlot()
		check(err)
		ioPlot, err := pi.IoPlot()
		check(err)
		cpuPlot, err := pi.CpuPlot()
		check(err)
		plots := make([][]*plot.Plot, 3)
		plots[0] = []*plot.Plot{memPlot}
		plots[1] = []*plot.Plot{ioPlot}
		plots[2] = []*plot.Plot{cpuPlot}
		err = pi.Save(plots, imgWidth, imgHeight, output)
		check(err)
	}
}
