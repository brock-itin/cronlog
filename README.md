# cronlog

Lightweight wrapper that captures and rotates cron job output with structured logging.

## Installation

```bash
go install github.com/yourusername/cronlog@latest
```

## Usage

Wrap any cron command with `cronlog` to capture output and write structured JSON logs with automatic rotation.

**Basic example:**

```bash
# In your crontab
* * * * * cronlog --job="backup" --log="/var/log/cronlog/backup.log" /usr/local/bin/backup.sh
```

**With log rotation options:**

```bash
cronlog \
  --job="cleanup" \
  --log="/var/log/cronlog/cleanup.log" \
  --max-size=10MB \
  --max-files=5 \
  -- /usr/local/bin/cleanup.sh --verbose
```

**Sample log output:**

```json
{"time":"2024-01-15T02:00:01Z","job":"backup","level":"info","msg":"job started","pid":12345}
{"time":"2024-01-15T02:00:03Z","job":"backup","level":"info","msg":"job finished","exit_code":0,"duration":"2.1s"}
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--job` | required | Job name used in log entries |
| `--log` | required | Path to log file |
| `--max-size` | `100MB` | Max log file size before rotation |
| `--max-files` | `7` | Number of rotated files to retain |

## License

MIT © [yourusername](https://github.com/yourusername)