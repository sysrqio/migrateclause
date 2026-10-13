# migrateclause

EU **Data Act** (Regulation (EU) 2023/2854) cloud **switching & porting** contract evidence CLI.

Validate a structured contract manifest (YAML) against Chapter VI limits: notice period (≤ 60 days),
transition (≤ 30 days), data retrieval (≥ 30 days), exportable-data schedule, termination assistance,
and the 2027 switching-fee ban.

## Quick start

```bash
go install github.com/sysrqio/migrateclause/cmd/migrateclause@latest

migrateclause audit --fixture compliant
migrateclause audit --manifest my-contract.yaml --format json --output report.json
```

## Manifest

See `fixtures/compliant.yaml`. Fields map to a legal review checklist — not a substitute for counsel.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Pass |
| 1 | I/O or parse error |
| 2 | Policy violation (one or more high/critical findings) |

## License

Apache-2.0
