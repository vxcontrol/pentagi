# RFC: Native Google Vertex AI Provider

> Status: RFC / planning. This document proposes a future provider. Native
> Vertex AI support does **not** exist in PentAGI today; every variable, type,
> and migration named here is a candidate, not a shipped feature. The intent is
> to agree on direction (especially the open questions) before any code is
> written.

## Summary

This RFC proposes a native **Google Vertex AI** provider (`vertex`) so users can
authenticate with GCP project credentials (Application Default Credentials or a
service-account key) instead of an AI Studio API key. It is motivated by #310
and #321: today the only Google option is the AI Studio `gemini` provider, which
accepts an API key against `https://generativelanguage.googleapis.com` and
cannot consume Vertex project/service-account credentials. Anthropic
Claude-on-Vertex is likewise not reachable through any current provider.

The RFC recommends a **staged** approach: a small v1 that adds Gemini-on-Vertex
with ADC / service-account authentication, and a separately-decided follow-up
for Claude-on-Vertex. It deliberately stops short of prescribing the final code
because two design questions (adapter strategy and Gemini-vs-Claude scope) need
maintainer direction first.

## Goals

- Let users authenticate to Vertex AI with GCP **Application Default
  Credentials** or a **service-account JSON** key, with explicit project ID and
  region/location, rather than an AI Studio API key.
- Support Gemini models served through Vertex AI in a first iteration.
- Keep the new path additive: existing providers and their configuration are
  untouched.
- Reuse existing request-shaping logic where it is safe to do so, to minimize
  new surface area.

## Non-Goals

- Replacing or changing the existing AI Studio `gemini` provider. It stays as-is.
- Per-request dynamic credentials, multi-project routing, or credential rotation
  (possible future work, explicitly out of scope here).
- Committing to Claude-on-Vertex in v1. Whether and how to add it is an open
  question below, not a decision in this RFC.
- Any change to flow lifecycle, queueing, or persisted state. This is a provider
  proposal only and introduces no hidden background state.

## Current Provider Landscape

PentAGI currently registers the following provider types: `openai`,
`anthropic`, `gemini`, `bedrock`, `ollama`, `custom`, `deepseek`, `glm`,
`kimi`, and `qwen`. The Google- and Anthropic-relevant options today are:

- **Google AI Studio (`gemini`)**: API-key auth against
  `https://generativelanguage.googleapis.com`. This is the consumer AI Studio
  surface, not Vertex AI. It cannot accept a GCP project or service-account
  credential.
- **Direct Anthropic (`anthropic`)**: `ANTHROPIC_API_KEY` /
  `ANTHROPIC_SERVER_URL` against Anthropic's own API. Not Vertex.
- **AWS Bedrock (`bedrock`)**: Anthropic and other models via AWS, with a
  multi-mode auth model (default AWS credential chain, bearer token, or static
  access/secret keys). This is the closest existing precedent for
  cloud-IAM-style provider auth.
- **Custom OpenAI-compatible (`custom`, `LLM_SERVER_*`)**: the present
  workaround for Vertex is to front it with a **LiteLLM** proxy that exposes an
  OpenAI-compatible endpoint, then point the `custom` provider at it. This works
  but requires running and securing extra infrastructure, and it relies on the
  proxy to translate Vertex auth and message schemas correctly.

A native `vertex` provider would remove the need for the LiteLLM workaround for
the common Gemini-on-Vertex case.

## Proposed v1 Scope

Proposed for the first iteration:

- A new `vertex` provider type that serves **Gemini models on Vertex AI**.
- Authentication via **ADC** or a **service-account JSON file**, plus explicit
  **project ID** and **location**.
- Wiring through the same registration and validation path every other provider
  uses, so the provider is selectable in flows and accepted by the REST API.

Proposed to defer:

- **Claude-on-Vertex** (Anthropic models through Vertex). See Open Questions Q2.
- Bearer-token / workload-identity auth beyond ADC and service-account file.
- Settings-UI configuration if the credential-file requirement makes env-only
  configuration the safer starting point (Open Questions Q4).

## Authentication Model

Vertex AI uses GCP IAM (OAuth2 access tokens minted from ADC or a service
account), not a static API key. The AWS Bedrock provider already demonstrates a
multi-mode auth pattern in PentAGI, and a Vertex auth model could mirror its
shape:

- **ADC (default)**: use Application Default Credentials resolved from the
  environment (for example `GOOGLE_APPLICATION_CREDENTIALS`, a mounted metadata
  service, or `gcloud auth application-default login`). Analogous to
  `BEDROCK_DEFAULT_AUTH` using the AWS default credential chain.
- **Explicit service-account file**: a candidate `VERTEX_CREDENTIALS_FILE`
  pointing at a mounted JSON key, used when ADC is not available. Analogous to
  Bedrock static credentials.
- **Project and location**: candidate `VERTEX_PROJECT_ID` and `VERTEX_LOCATION`
  (for example `us-central1`), which Vertex requires and which have no AI Studio
  equivalent.
- **Optional regional/private endpoint**: a candidate `VERTEX_SERVER_URL` for
  regional or private Service endpoints, analogous to `BEDROCK_SERVER_URL`.

All of the above names are **candidate** keys for discussion, not shipped
configuration.

## Provider Architecture Options

The existing `gemini` provider is built on the langchaingo `googleai` client
configured for the AI Studio REST surface with an API key. Vertex AI changes
both the transport (GCP IAM auth, `aiplatform.googleapis.com` regional
endpoints) and, for Claude, the message schema. Two broad options:

- **Option A - parameterize an existing adapter.** Add a Vertex transport/auth
  mode to the Gemini path so request-shaping is shared and only auth + endpoint
  differ. Lower duplication, but couples two surfaces that authenticate very
  differently and risks regressing the stable AI Studio path.
- **Option B - a separate `vertex` package.** A dedicated provider that owns its
  auth and endpoint logic and reuses request-shaping helpers where practical.
  More code, cleaner separation, no risk to the existing `gemini` provider.

This RFC leans toward **Option B for v1** (separation first, extract shared
helpers later if duplication proves real), but defers to maintainer preference
(Open Questions Q1).

A key architectural note: **Claude-on-Vertex likely does not fit the same
adapter as Gemini-on-Vertex.** Gemini-on-Vertex uses the Gemini request/response
schema, while Claude-on-Vertex uses the Anthropic message schema over a Vertex
endpoint with GCP auth. That asymmetry argues for routing Claude-on-Vertex
through the **Anthropic** adapter with a Vertex auth/endpoint mode, rather than
bolting it onto a Gemini-shaped Vertex provider. Treating "Vertex" as one
monolithic provider for both model families would mix two schemas behind one
type.

## Config and Migration Considerations

A native provider would follow the repository's documented "Adding a New LLM
Provider" checklist (`CLAUDE.md`). At a high level that means, when
implementation is approved:

- A `ProviderVertex` type constant and default provider name.
- Registration in the provider factory functions.
- Addition of `vertex` to the REST `Valid()` whitelist (without this the REST
  API rejects the type with 422).
- Candidate config keys in the central config (the `VERTEX_*` names above).
- A goose migration adding `vertex` to the `PROVIDER_TYPE` enum, following the
  enum-swap pattern used by the existing provider migrations under
  `backend/migrations/sql/` (Up: recreate the enum including the new value;
  Down: remove rows of the new type, then recreate the enum without it).

The migration is the least reversible step and the Down path deletes any rows of
the new provider type, so it warrants explicit review. None of these changes are
part of this RFC; they are listed so the eventual implementation size is clear.

## Frontend / Installer Considerations

For parity with other providers, a future implementation would add a provider
icon and register it, and decide whether Vertex appears in the Settings UI. The
service-account-file requirement is the wrinkle: unlike an API key, a JSON key
is a file and should not be pasted into a web form or stored as plain settings
text. A reasonable starting point is **env/file-mounted configuration only**,
with Settings-UI support considered later (Open Questions Q4). The interactive
installer wizard could later grow a Vertex section that asks for project,
location, and credential-file path.

## Testing Strategy

- **Unit / config**: a future `vertex` provider would get the same provider
  unit tests other providers have, exercised through the existing config-loading
  path.
- **Provider validation**: the `ctester` utility (which tests LLM agent
  capabilities and tool-calling agent types) would be the pre-merge smoke test
  for the new provider once credentials are available.
- **Credentials caveat**: end-to-end testing requires real GCP credentials and a
  Vertex-enabled project, which maintainers would need to supply or stub. This
  is called out as a practical gating factor on any implementation PR.

## Open Questions

1. **Adapter strategy** - parameterize the existing Gemini adapter with a Vertex
   transport/auth mode (Option A), or ship a separate `vertex` package
   (Option B)?
2. **Scope** - Gemini-on-Vertex only in v1, or include Claude-on-Vertex? If
   Claude-on-Vertex is in scope, should it route through the Anthropic adapter
   (Vertex auth/endpoint mode) rather than a Gemini-shaped provider, given the
   schema difference?
3. **Auth surface** - are ADC and a service-account JSON file sufficient for v1,
   or is bearer-token / workload-identity auth (mirroring the Bedrock multi-auth
   approach) also wanted?
4. **Web settings** - should Vertex be configurable from the Settings UI like
   other providers, or env/file-mounted only at first, given the credential-file
   requirement?

## Security Considerations

- **Service-account JSON is a sensitive secret.** It should be **file-mounted or
  secret-managed**, never pasted into UI text, never committed, and never
  written to logs. Provider initialization and any error surface must avoid
  echoing credential contents or file paths beyond what is necessary to
  diagnose a misconfiguration.
- **Least privilege**: documentation for any implementation should recommend a
  dedicated service account scoped to the minimum Vertex AI prediction roles.
- **No hidden state**: this proposal adds a provider, not background lifecycle
  state; credentials are supplied explicitly via env/mounted file and are not
  cached or queued anywhere implicit.
- **Endpoint trust**: regional/private endpoint overrides should be validated so
  a misconfigured `VERTEX_SERVER_URL` cannot silently redirect traffic.

## Suggested First Milestone

If maintainers confirm Option B and a Gemini-on-Vertex-only v1, a minimal first
PR could add the `vertex` provider package, the type constant and registration,
the REST whitelist entry, the candidate `VERTEX_*` config keys, the enum
migration, and `.env.example` plus docs - with Claude-on-Vertex, extra auth
modes, and Settings-UI support tracked as explicit follow-ups. Confirmation on
Open Questions Q1 and Q2 is the blocker before any of that work begins.

## References

- #310 - original Vertex AI configuration request (clarified that Vertex is not
  natively supported today).
- #321 - native Vertex AI provider request and implementation outline.
- `CLAUDE.md` - "Adding a New LLM Provider" checklist.
- `backend/migrations/sql/` - existing provider enum-swap migrations that
  demonstrate the `PROVIDER_TYPE` pattern.
