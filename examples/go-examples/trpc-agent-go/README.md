# tRPC-Agent API Example

This minimal example exposes a tool-using `tRPC-Agent-Go` agent through the open-source tRPC-Go framework, exports traces and metrics over OTLP/HTTP, and generates traffic every three seconds.

It exposes:

- `GET /trpc-agent/v1/apps/calculator/structure`
- `POST /trpc-agent/v1/apps/calculator/runs`

## Run

Configure the model provider and BlueKing APM connection, then run:

```bash
docker build -t trpc-agent-go-apm:latest .

docker run --rm --name trpc-agent-go-demo \
  -p 8080:8080 \
  -e OTLP_ENDPOINT="<host:port>" \
  -e TOKEN="<your-apm-token>" \
  -e SERVICE_NAME="trpc-agent-go-demo" \
  -e OPENAI_API_KEY="<your-api-key>" \
  -e OPENAI_BASE_URL="<your-compatible-base-url>" \
  trpc-agent-go-apm:latest
```

After startup, `loopQuery` calls the local agent every three seconds so traces, metrics, and one LLM-to-tool call can be verified in BlueKing APM.
