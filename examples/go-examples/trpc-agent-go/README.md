# tRPC-Agent API Example

This minimal example exposes a tool-using `tRPC-Agent-Go` agent through the open-source tRPC-Go framework, exports traces and metrics over OTLP/HTTP, and generates traffic every three seconds.

It exposes:

- `GET /trpc-agent/v1/apps/calculator/structure`
- `POST /trpc-agent/v1/apps/calculator/runs`

## Run

Configure the model provider and BlueKing APM connection, then run:

```bash
export OPENAI_API_KEY="<your-api-key>"
export OPENAI_BASE_URL="<your-compatible-base-url>"
export OTLP_ENDPOINT="<host:port>"
export TOKEN="<your-apm-token>"
export SERVICE_NAME="trpc-agent-go-demo"
go run .
```

After startup, `loopQuery` calls the local agent every three seconds so traces, metrics, and one LLM-to-tool call can be verified in BlueKing APM.
