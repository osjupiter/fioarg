# fioh

**fio** interactive **h**elper — a Go CLI that interactively builds a `fio` command from your selections and runs it.

## Setup

```bash
git clone <this> fioh && cd fioh
go mod init fioh
go mod tidy
go build -o fioh .
```

Install `fio` separately:

- Debian/Ubuntu: `sudo apt install fio`
- RHEL/Fedora: `sudo dnf install fio`
- macOS: `brew install fio`

## Usage

```bash
./fioh
```

Arrow keys to select, Enter to confirm, Tab/Shift+Tab to move between fields, Ctrl+C to abort.

Prompts, in order:

| Group | Options |
|---|---|
| Basics | `--name` / `--rw` / `--bs` |
| Scale & engine | `--size` / `--numjobs` / `--iodepth` / `--ioengine` / `--direct` |
| Time | `--time_based` / `--runtime` |
| Target | `--filename` or `--directory` and its path |
| Output / mixed | `--rwmixread` / `--output-format` / `--group_reporting` / extra options |

Finally the assembled `fio ...` command is shown, and you can choose **Run / Print only / Cancel**.

## Example

Selections like:

```
name       = randread-test
rw         = randread
bs         = 4k
size       = 1G
numjobs    = 4
iodepth    = 32
ioengine   = libaio
direct     = yes
time_based = yes
runtime    = 30
target     = filename: /tmp/fio-testfile
group_rep  = yes
```

produce:

```
fio --name=randread-test --rw=randread --bs=4k --size=1G --numjobs=4 \
    --iodepth=32 --ioengine=libaio --direct=1 --time_based --runtime=30 \
    --filename=/tmp/fio-testfile --output-format=normal --group_reporting
```

> **Warning:** Pointing `--filename` at a block device (e.g. `/dev/sdX`) with a write workload destroys its data.
