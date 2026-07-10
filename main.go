// fioh: an interactive wrapper that builds and runs a fio command
//
// Usage:
//   go mod init fioh
//   go mod tidy
//   go run .
//   # or build it:
//   go build -o fioh . && ./fioh
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

type fioConfig struct {
	Name       string
	RW         string
	BS         string
	Size       string
	NumJobs    string
	IODepth    string
	IOEngine   string
	Direct     bool
	TimeBased  bool
	Runtime    string
	Target     string
	TargetKind string // "filename" | "directory"
	RWMixRead  string
	OutputFmt  string
	GroupRep   bool
	Extra      string
}

// buildArgs assembles the fio argument list from the selections
func (c *fioConfig) buildArgs() []string {
	var args []string

	kv := func(k, v string) {
		if v != "" {
			args = append(args, fmt.Sprintf("--%s=%s", k, v))
		}
	}
	kvBool := func(k string, b bool) {
		v := "0"
		if b {
			v = "1"
		}
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}
	flag := func(k string, b bool) {
		if b {
			args = append(args, "--"+k)
		}
	}

	kv("name", c.Name)
	kv("rw", c.RW)
	kv("bs", c.BS)
	kv("size", c.Size)
	kv("numjobs", c.NumJobs)
	kv("iodepth", c.IODepth)
	kv("ioengine", c.IOEngine)
	kvBool("direct", c.Direct)

	flag("time_based", c.TimeBased)
	if c.TimeBased {
		kv("runtime", c.Runtime)
	}

	switch c.TargetKind {
	case "filename":
		kv("filename", c.Target)
	case "directory":
		kv("directory", c.Target)
	}

	// add rwmixread only for mixed workloads
	if c.RW == "randrw" || c.RW == "readwrite" || c.RW == "rw" {
		kv("rwmixread", c.RWMixRead)
	}

	kv("output-format", c.OutputFmt)
	flag("group_reporting", c.GroupRep)

	if strings.TrimSpace(c.Extra) != "" {
		args = append(args, strings.Fields(c.Extra)...)
	}
	return args
}

func main() {
	// defaults
	cfg := fioConfig{
		Name:       "benchtest",
		RW:         "randread",
		BS:         "4k",
		Size:       "1G",
		NumJobs:    "1",
		IODepth:    "32",
		IOEngine:   "libaio",
		Direct:     true,
		TimeBased:  true,
		Runtime:    "30",
		OutputFmt:  "normal",
		GroupRep:   true,
		TargetKind: "filename",
		Target:     "/tmp/fio-testfile",
		RWMixRead:  "70",
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("fioh — fio interactive helper").
				Description("Answer the prompts to build a fio command.\nPress Ctrl+C to abort."),
		),

		// --- basics ---
		huh.NewGroup(
			huh.NewInput().
				Title("Job name (--name)").
				Value(&cfg.Name),

			huh.NewSelect[string]().
				Title("I/O pattern (--rw)").
				Options(
					huh.NewOption("randread   (random read)", "randread"),
					huh.NewOption("randwrite  (random write)", "randwrite"),
					huh.NewOption("read       (sequential read)", "read"),
					huh.NewOption("write      (sequential write)", "write"),
					huh.NewOption("randrw     (mixed random read/write)", "randrw"),
					huh.NewOption("readwrite  (mixed sequential read/write)", "readwrite"),
				).
				Value(&cfg.RW),

			huh.NewSelect[string]().
				Title("Block size (--bs)").
				Options(
					huh.NewOption("4k", "4k"),
					huh.NewOption("8k", "8k"),
					huh.NewOption("16k", "16k"),
					huh.NewOption("32k", "32k"),
					huh.NewOption("64k", "64k"),
					huh.NewOption("128k", "128k"),
					huh.NewOption("1M", "1M"),
				).
				Value(&cfg.BS),
		),

		// --- scale / engine ---
		huh.NewGroup(
			huh.NewInput().
				Title("Transfer/file size (--size)").
				Description("e.g. 1G, 512M, 100M").
				Value(&cfg.Size),

			huh.NewInput().
				Title("Parallel jobs (--numjobs)").
				Value(&cfg.NumJobs),

			huh.NewInput().
				Title("I/O depth (--iodepth)").
				Description("Only meaningful with async engines").
				Value(&cfg.IODepth),

			huh.NewSelect[string]().
				Title("I/O engine (--ioengine)").
				Options(
					huh.NewOption("libaio    (Linux async, the usual choice)", "libaio"),
					huh.NewOption("io_uring  (newer Linux)", "io_uring"),
					huh.NewOption("psync     (pread/pwrite)", "psync"),
					huh.NewOption("sync      (read/write)", "sync"),
					huh.NewOption("posixaio  (POSIX AIO)", "posixaio"),
					huh.NewOption("mmap", "mmap"),
				).
				Value(&cfg.IOEngine),

			huh.NewConfirm().
				Title("Enable direct I/O (--direct=1)?").
				Value(&cfg.Direct),
		),

		// --- runtime ---
		huh.NewGroup(
			huh.NewConfirm().
				Title("Time based (--time_based)?").
				Description("If enabled, keeps running until runtime even after size is exhausted").
				Value(&cfg.TimeBased),

			huh.NewInput().
				Title("Runtime in seconds (--runtime)").
				Description("Ignored when time_based=false").
				Value(&cfg.Runtime),
		),

		// --- target ---
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Target type").
				Options(
					huh.NewOption("filename  (single file / device)", "filename"),
					huh.NewOption("directory (create files inside a directory)", "directory"),
				).
				Value(&cfg.TargetKind),

			huh.NewInput().
				Title("Target path").
				Description("e.g. /tmp/fio-testfile, /dev/nvme0n1  <- writing to a raw device destroys its data!").
				Value(&cfg.Target),
		),

		// --- mixed workload / output / misc ---
		huh.NewGroup(
			huh.NewInput().
				Title("Read ratio % (--rwmixread)").
				Description("Applied only for randrw / readwrite").
				Value(&cfg.RWMixRead),

			huh.NewSelect[string]().
				Title("Output format (--output-format)").
				Options(
					huh.NewOption("normal", "normal"),
					huh.NewOption("json", "json"),
					huh.NewOption("terse", "terse"),
				).
				Value(&cfg.OutputFmt),

			huh.NewConfirm().
				Title("Add --group_reporting?").
				Value(&cfg.GroupRep),

			huh.NewInput().
				Title("Extra options (optional)").
				Description("Appended as-is. e.g. --numa_cpu_nodes=0 --thread").
				Value(&cfg.Extra),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
		os.Exit(1)
	}

	args := cfg.buildArgs()
	cmdline := "fio " + strings.Join(args, " ")

	fmt.Println()
	fmt.Println("── Assembled command ─────────────────────────────────")
	fmt.Println(cmdline)
	fmt.Println("──────────────────────────────────────────────────────")
	fmt.Println()

	// final confirmation before running
	action := "run"
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What do you want to do?").
				Options(
					huh.NewOption("Run it", "run"),
					huh.NewOption("Print the command only (for copy/paste)", "print"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&action),
		),
	).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "confirmation error:", err)
		os.Exit(1)
	}

	switch action {
	case "cancel":
		fmt.Println("Canceled.")
		return
	case "print":
		fmt.Println(cmdline)
		return
	case "run":
		if _, err := exec.LookPath("fio"); err != nil {
			fmt.Fprintln(os.Stderr, "error: fio not found in PATH")
			os.Exit(1)
		}
		cmd := exec.Command("fio", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "fio execution error:", err)
			os.Exit(1)
		}
	}
}
