# MCP Temporal Server

An MCP server that integrates with Temporal to expose workflow history and additional resources.

## Tools

- `workflow_history`: Retrieve the history of a Temporal workflow by providing the required arguments.

## Resources

- `test://resource`: A guide for understanding workflow histories.

## Environment

- `TEMPORAL_ADDRESS`: The Temporal server address (default: `localhost:7233`).
- `TEMPORAL_NAMESPACE`: The Temporal namespace (default: `default`).
- `PORT`: The port for the MCP server (default: `8080`).

## Usage

### Running the Server

Start the MCP server:

```sh
go run cmd/server/main.go
```

### Using the `workflow_history` Tool

Send a request to retrieve a workflow's history:

```json
{
  "tool": "workflow_history",
  "arguments": {
    "workflow_id": "your-workflow-id",
    "run_id": "your-run-id"
  }
}
```

### Accessing the `test://resource` Resource

Retrieve the guide for understanding workflow histories:

```json
{
  "resource": "test://resource"
}
```