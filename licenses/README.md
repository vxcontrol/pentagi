# Third-Party Licenses

This directory contains license information for all PentAGI dependencies.

## Quick Start

Run the generator script to create/update license reports (run from project root):

```bash
./scripts/generate-licenses.sh
```

## Generated Files

### Backend (Go)
- `backend-dependencies.txt` - Complete list of Go modules
- `backend-licenses.csv` - Detailed license information (CSV format)

### Frontend (npm)
- `frontend-dependencies.json` - Complete npm dependency tree (JSON)
- `frontend-licenses.json` - Detailed license data (JSON)
- `frontend-licenses.csv` - License data (CSV)

**Note:** 
- Backend reports require `go-licenses` tool: `go install github.com/google/go-licenses@latest`
- Frontend reports require `npm ci` in the frontend directory first.

## License

PentAGI is licensed under **MIT License**.

All third-party dependencies use MIT-compatible licenses:
- MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0, 0BSD

## Docker Builds

License reports are automatically generated during Docker builds and included in the final image at `/opt/pentagi/licenses/`.

## More Information

- Project License: [../LICENSE](../LICENSE)
- Legal Notices: [../NOTICE](../NOTICE)
- Full Documentation: [../README.md](../README.md)
