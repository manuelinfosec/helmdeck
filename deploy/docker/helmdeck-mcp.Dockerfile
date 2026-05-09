# Distroless image for the helmdeck-mcp stdio bridge.
# Built by goreleaser; the helmdeck-mcp binary is supplied by the build context.
# See ADR 030.
FROM gcr.io/distroless/static:nonroot

# MCP Registry namespace verification — the official registry's OCI
# validator reads this label to confirm we own the namespace under
# which we're publishing (io.github.tosin2013/helmdeck). Required as
# of registry validator commit 2025-12; see
# https://github.com/modelcontextprotocol/registry/blob/main/internal/validators/registries/oci.go
LABEL io.modelcontextprotocol.server.name="io.github.tosin2013/helmdeck"

COPY helmdeck-mcp /usr/local/bin/helmdeck-mcp
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/helmdeck-mcp"]
