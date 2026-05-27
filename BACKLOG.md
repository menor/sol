# Sol Backlog

Items here are candidates for future work — not yet scheduled or prioritized.

---

## Agent-first CLI consistency (from Cloudflare CLI post)

**Source:** https://blog.cloudflare.com/cf-cli-local-explorer/

Cloudflare enforces strict naming rules at the schema layer to prevent agent confusion. Key finding: agents expect CLIs to be consistent. If one command uses `info` and another uses `get`, agents will call non-existent commands.

**Findings for Sol:**

1. **Audit command verb consistency.** Sol currently has both `auth:info` and could grow `project:get`-style names. Pick one verb (`info` or `get`) and enforce it everywhere.

2. **Enforce `--output` flag on every command.** The `--output json|toon` flag must work identically across all commands. No command should silently ignore it.

3. **Schema flag is the right instinct.** Sol's existing `--schema` flag aligns with their schema-first approach. Keep investing in it as command count grows.

**Not applicable to Sol:**
- Local/remote parity (`--local` flag) — Cloudflare-specific for Workers/Miniflare
- Auto-generation from schema — Sol has far fewer commands; overkill for now
