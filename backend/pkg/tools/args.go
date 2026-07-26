package tools

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FileOp is a type alias (not a distinct type) for String: it shares the
// exact same lenient UnmarshalJSON, so every FileOp value - including the
// ReadFile/WriteFile/EditFile constants below - automatically benefits from
// the quote-unwrapping behaviour documented on String, with zero changes
// needed at any existing call site (switches, comparisons, assignments)
// elsewhere in the codebase.
type FileOp = String

const (
	ReadFile  FileOp = "read_file"
	WriteFile FileOp = "write_file"
	EditFile  FileOp = "edit_file"
)

// FileAction's field descriptions are layered on purpose, each one adding
// only what the level above doesn't already say: the tool description (see
// registry.go) gives the read/write/edit decision; Action's enum values name
// which payload field each one consumes; Content/Diff each spec out the
// format of their own payload. Path/Message never vary by action, so they're
// described once, action-agnostically.
type FileAction struct {
	Action  FileOp `json:"action" jsonschema:"required,type=string,enum=read_file,enum=write_file,enum=edit_file" jsonschema_description:"'read_file' reads the file (no other field needed). 'write_file' overwrites it with 'content' (the whole file). 'edit_file' applies the patch in 'diff' (existing content elsewhere is untouched)."`
	Content string `json:"content,omitempty" jsonschema_description:"write_file only: the complete new file content (not a diff, not a partial update)."`
	Diff    String `json:"diff,omitempty" jsonschema:"type=string" jsonschema_description:"edit_file only: unified-diff hunk(s) - '@@ -old +new @@' header, then ' '/'-'/'+' lines. Always keep at least one unchanged context line so the location is unambiguous; context/removed lines must match the file's current content verbatim (read_file first). Header line numbers are only a hint - the line text is what must match."`
	Path    String `json:"path" jsonschema:"required,type=string" jsonschema_description:"Absolute path to the file"`
	Message string `json:"message" jsonschema:"required,title=File action message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary describing what you are reading, writing, or editing and why. Written in the engagement language declared by your system prompt."`
}

// BrowserAction is a type alias for String - see the FileOp comment above.
type BrowserAction = String

const (
	Markdown BrowserAction = "markdown"
	HTML     BrowserAction = "html"
	Links    BrowserAction = "links"
)

type Browser struct {
	Url     string        `json:"url" jsonschema:"required" jsonschema_description:"URL to open in the browser"`
	Action  BrowserAction `json:"action" jsonschema:"required,type=string,enum=markdown,enum=html,enum=links" jsonschema_description:"Action to perform in the browser. 'markdown' - Returns the content of the page in markdown format. 'html' - Returns the content of the page in html format. 'links' - Get the list of all URLs on the page to be used in later calls (e.g., open search results after the initial search lookup)."`
	Message string        `json:"message" jsonschema:"required,title=Browser action message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary describing what content you are fetching, in which format, and why. Written in the engagement language declared by your system prompt."`
}

type SubtaskInfo struct {
	Title       string `json:"title" jsonschema:"required,title=Subtask title" jsonschema_description:"Engagement-log plan entry — short title of this subtask capturing its main goal as it appears in the engagement plan. Written in the engagement language declared by your system prompt."`
	Description string `json:"description" jsonschema:"required,title=Subtask to complete" jsonschema_description:"Engagement-log plan entry — full description of this subtask (instructions, rules, requirements, success criteria) as it appears in the engagement plan. Written in the engagement language declared by your system prompt."`
}

type SubtaskList struct {
	Subtasks []SubtaskInfo `json:"subtasks" jsonschema:"required,title=Subtasks to complete" jsonschema_description:"Ordered list of subtasks produced by decomposing the task. Each subtask's title and description are engagement-log plan entries (see SubtaskInfo) — written in the engagement language declared by your system prompt."`
	Message  string        `json:"message" jsonschema:"required,title=Subtask generation result" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary on the generation result and the main goal of the plan. Written in the engagement language declared by your system prompt."`
}

// SubtaskOperationType defines the type of operation to perform on a subtask.
// It is a type alias for String - see the FileOp comment above.
type SubtaskOperationType = String

const (
	SubtaskOpAdd     SubtaskOperationType = "add"
	SubtaskOpRemove  SubtaskOperationType = "remove"
	SubtaskOpModify  SubtaskOperationType = "modify"
	SubtaskOpReorder SubtaskOperationType = "reorder"
)

// SubtaskOperation defines a single operation on the subtask list for delta-based refinement
type SubtaskOperation struct {
	Op          SubtaskOperationType `json:"op" jsonschema:"required,type=string,enum=add,enum=remove,enum=modify,enum=reorder" jsonschema_description:"Operation type: 'add' creates a new subtask, 'remove' deletes a subtask by ID, 'modify' updates title/description of existing subtask, 'reorder' moves a subtask to a different position"`
	ID          *int64               `json:"id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"ID of existing subtask (required for remove/modify/reorder operations)"`
	AfterID     *int64               `json:"after_id,omitempty" jsonschema:"title=Insert after ID" jsonschema_description:"For add/reorder: insert after this subtask ID (null/0 = insert at beginning)"`
	Title       string               `json:"title,omitempty" jsonschema:"title=New title" jsonschema_description:"Engagement-log plan entry — new subtask title (required for add, optional for modify). Written in the engagement language declared by your system prompt."`
	Description string               `json:"description,omitempty" jsonschema:"title=New description" jsonschema_description:"Engagement-log plan entry — new subtask description (required for add, optional for modify). Written in the engagement language declared by your system prompt."`
}

type SubtaskInfoPatch struct {
	ID int64 `json:"id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"ID of the subtask (populated by the system for existing subtasks)"`
	SubtaskInfo
}

// SubtaskPatch is the delta-based refinement output for modifying subtask lists
type SubtaskPatch struct {
	Operations []SubtaskOperation `json:"operations" jsonschema:"required" jsonschema_description:"List of operations to apply to the current subtask list. Empty array means no changes needed. Each operation's title/description, when present, is an engagement-log plan entry (see operations)."`
	Message    string             `json:"message" jsonschema:"required,title=Refinement summary" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary on the changes made and the justification for the modifications. Written in the engagement language declared by your system prompt."`
}

type TaskResult struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,type=boolean" jsonschema_description:"True if the task was executed successfully and its objective was reached"`
	Result  string `json:"result" jsonschema:"required,title=Task result description" jsonschema_description:"Engagement-log closing entry — fully detailed write-up of the task/subtask outcome (what was achieved or why it failed). Written in the engagement language declared by your system prompt."`
	Message string `json:"message" jsonschema:"required,title=Task result message" jsonschema_description:"Engagement-log closing summary — a concise 1-2 sentence recap of the outcome and the path taken to reach the goal. Written in the engagement language declared by your system prompt."`
}

type AskUser struct {
	Message string `json:"message" jsonschema:"required,title=Question for user" jsonschema_description:"Engagement-log entry — clarification question or any other information addressed to the engagement coordinator. Written in the engagement language declared by your system prompt."`
}

type Done struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,type=boolean" jsonschema_description:"True if the subtask was executed successfully and its objective was reached"`
	Result  string `json:"result" jsonschema:"required,title=Subtask result description" jsonschema_description:"Engagement-log closing entry — fully detailed write-up of the subtask outcome (what was achieved or why it failed). Written in the engagement language declared by your system prompt."`
	Message string `json:"message" jsonschema:"required,title=Subtask result message" jsonschema_description:"Engagement-log closing summary — a concise 1-2 sentence recap of the subtask outcome. Written in the engagement language declared by your system prompt."`
}

type TerminalAction struct {
	Input   string `json:"input" jsonschema:"required" jsonschema_description:"Command to be run in the docker container terminal according to the command-execution rules"`
	Cwd     string `json:"cwd" jsonschema:"required" jsonschema_description:"Custom current working directory to execute the command in, or the default directory if not specified"`
	Detach  Bool   `json:"detach" jsonschema:"required,type=boolean" jsonschema_description:"Set to true for INTERACTIVE or LONG-RUNNING commands: shells (msfconsole, bash, python), listeners (nc -lvnp, socat TCP-LISTEN), servers (python -m http.server, php -S), monitors (tcpdump, tail -f). These commands expect user input or run indefinitely. When true: command runs in background, you get immediate confirmation, no stdout/stderr captured. When false: command must complete within timeout and return output. For quick batch commands (nmap, curl, ls) use false"`
	Timeout Int64  `json:"timeout" jsonschema:"required,type=integer" jsonschema_description:"Execution time limit in seconds. Use 0 value to apply the configured server default timeout. Explicit positive values are accepted up to 10800 seconds (3 hours); any value outside the 1–10800 range or non-positive is replaced by the server default. For batch commands that may run long, use the 'timeout' shell utility INSIDE your command to ensure clean completion with full output: 'timeout 55 nmap -sV target' (set 5-10 seconds less than this parameter). For interactive/long-running commands, use detach=true instead of relying solely on timeout"`
	Message string `json:"message" jsonschema:"required,title=Terminal command message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining what you intend to achieve by running this command. Written in the engagement language declared by your system prompt."`
}

type AskAdvice struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — detailed question for the senior mentor describing the issue and what you need to figure out. Always written in English; never translated."`
	Code     string `json:"code" jsonschema_description:"Technical-channel excerpt — optional code snippet relevant to the question. Preserve verbatim; do not translate code or comments."`
	Output   string `json:"output" jsonschema_description:"Technical-channel excerpt — optional stdout/stderr excerpt relevant to the question. Preserve verbatim."`
	Message  string `json:"message" jsonschema:"required,title=Ask advice message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining what you are asking the mentor and why. Written in the engagement language declared by your system prompt."`
}

type ComplexSearch struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — research question for the researcher specialist with full context, target content type, and required structure. Always written in English; never translated."`
	Message  string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the research question. Written in the engagement language declared by your system prompt."`
}

type SearchAction struct {
	Query      string `json:"query" jsonschema:"required" jsonschema_description:"Technical-channel payload — query for the specific search engine (google, duckduckgo, tavily, traversaal, perplexity, serper, etc.). ALWAYS written in English regardless of the engagement language: internet sources are predominantly indexed in English and non-English queries return poor or empty results. Short and exact queries return better results."`
	MaxResults Int64  `json:"max_results" jsonschema:"required,type=integer" jsonschema_description:"Maximum number of results to return (minimum 1; maximum 10; default 5)"`
	Message    string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining the expected result and how it advances the goal. Written in the engagement language declared by your system prompt."`
}

type SearchResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search result" jsonschema_description:"Technical-channel payload — fully detailed search report (or error explanation) returned to the calling agent for further reasoning. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Search result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with a short summary of the result. Written in the engagement language declared by your system prompt."`
}

type SploitusAction struct {
	Query       string `json:"query" jsonschema:"required" jsonschema_description:"Technical-channel payload — search query for Sploitus (e.g. 'ssh', 'apache 2.4', 'CVE-2021-44228'). ALWAYS written in English; the Sploitus index is English-only. Short and precise queries return the best results."`
	ExploitType String `json:"exploit_type,omitempty" jsonschema:"type=string,enum=exploits,enum=tools" jsonschema_description:"What to search for: 'exploits' (default) for exploit code and PoCs, 'tools' for offensive security tools"`
	Sort        String `json:"sort,omitempty" jsonschema:"type=string,enum=default,enum=date,enum=score" jsonschema_description:"Result ordering: 'default' (relevance), 'date' (newest first), 'score' (highest CVSS first)"`
	MaxResults  Int64  `json:"max_results" jsonschema:"required,type=integer" jsonschema_description:"Maximum number of results to return (minimum 1; maximum 25; default 10)"`
	Message     string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining the expected result and how it advances the goal. Written in the engagement language declared by your system prompt."`
}

// WebSearchAction is the LLM-facing schema for the unified web_search tool. The agent
// supplies a query and an intent `mode`; web_search selects the concrete engine,
// retries transient failures, and falls back across providers automatically.
type WebSearchAction struct {
	Query       string `json:"query" jsonschema:"required" jsonschema_description:"Technical-channel payload — the search query. ALWAYS written in English regardless of the engagement language: internet sources and the underlying engines are indexed in English and non-English queries return poor or empty results. Keep it short and keyword-focused."`
	Mode        String `json:"mode,omitempty" jsonschema:"type=string,enum=links,enum=answer,enum=research,enum=exploit" jsonschema_description:"What you need back (default 'answer'). 'links' — a list of source links with snippets (fastest/cheapest). 'answer' — a synthesized answer over live sources. 'research' — deep multi-source analysis with reasoning (most thorough). 'exploit' — exploit code, PoCs, and offensive tools. web_search picks the concrete engine for you and falls back automatically; you do NOT name an engine."`
	MaxResults  Int64  `json:"max_results,omitempty" jsonschema:"type=integer" jsonschema_description:"Maximum number of results to return (1–25; default 5). Ignored by answer/research engines that return a single synthesized answer."`
	ExploitType String `json:"exploit_type,omitempty" jsonschema:"type=string,enum=exploits,enum=tools" jsonschema_description:"exploit mode only: 'exploits' (default) for exploit code and PoCs, 'tools' for offensive-security tools."`
	Sort        String `json:"sort,omitempty" jsonschema:"type=string,enum=default,enum=date,enum=score" jsonschema_description:"exploit mode only: result ordering — 'default' (relevance), 'date' (newest first), 'score' (highest CVSS first)."`
	Message     string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining the expected result and how it advances the goal. Written in the engagement language declared by your system prompt."`
}

type GraphitiSearchAction struct {
	SearchType     String   `json:"search_type" jsonschema:"required,type=string,enum=temporal_window,enum=entity_relationships,enum=diverse_results,enum=episode_context,enum=successful_tools,enum=recent_context,enum=entity_by_label" jsonschema_description:"Type of search to perform: temporal_window (time-bounded search), entity_relationships (graph traversal from an entity), diverse_results (anti-redundancy search), episode_context (full agent reasoning and tool outputs), successful_tools (proven techniques), recent_context (latest findings), entity_by_label (type-specific entity search, REQUIRES node_labels)"`
	Query          string   `json:"query" jsonschema:"required" jsonschema_description:"Technical-channel payload — natural language query against the team's temporal knowledge graph. REQUIRED for EVERY search_type, including entity_relationships/entity_by_label: it re-ranks/filters the graph traversal, it is never optional just because center_node_uuid or node_labels is set. ALWAYS written in English regardless of the engagement language: the graph is indexed in English and shared across all engagements; non-English queries will fail to retrieve stored episodic memory."`
	MaxResults     *Int64   `json:"max_results,omitempty" jsonschema:"title=Maximum Results,type=integer" jsonschema_description:"Maximum number of results to return (default varies by search type)"`
	TimeStart      string   `json:"time_start,omitempty" jsonschema_description:"Start of time window (ISO 8601 format, required for temporal_window)"`
	TimeEnd        string   `json:"time_end,omitempty" jsonschema_description:"End of time window (ISO 8601 format, required for temporal_window)"`
	CenterNodeUUID string   `json:"center_node_uuid,omitempty" jsonschema_description:"REQUIRED for entity_relationships. Copy verbatim from the 'UUID:' field of an entity/community returned by an EARLIER graphiti_search call in this conversation (any search_type). NEVER invent one — not a flow/task ID, hostname, or title; if no prior result yielded a UUID, run recent_context/entity_by_label first"`
	MaxDepth       *Int64   `json:"max_depth,omitempty" jsonschema:"title=Maximum Depth,type=integer" jsonschema_description:"Maximum graph traversal depth (default: 2, max: 3, for entity_relationships)"`
	NodeLabels     []string `json:"node_labels,omitempty" jsonschema_description:"Filter to specific node types — EXACT taxonomy names, PascalCase singular (e.g., ['Host', 'Service', 'Vulnerability']). REQUIRED (non-empty) when search_type is entity_by_label; optional filter otherwise"`
	EdgeTypes      []string `json:"edge_types,omitempty" jsonschema_description:"Filter to specific relationship types — EXACT taxonomy names, UPPER_SNAKE_CASE (e.g., ['HAS_PORT', 'HAS_VULNERABILITY'])"`
	DiversityLevel String   `json:"diversity_level,omitempty" jsonschema:"type=string,enum=low,enum=medium,enum=high" jsonschema_description:"How much diversity to prioritize (default: medium, for diverse_results)"`
	MinMentions    *Int64   `json:"min_mentions,omitempty" jsonschema:"title=Minimum Mentions,type=integer" jsonschema_description:"Minimum episode mentions (default: 2, for successful_tools)"`
	RecencyWindow  String   `json:"recency_window,omitempty" jsonschema:"type=string,enum=1h,enum=6h,enum=24h,enum=7d" jsonschema_description:"How far back to search (default: 24h, for recent_context)"`
	Message        string   `json:"message" jsonschema:"required,title=Search message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the search query and expected results. Written in the engagement language declared by your system prompt."`
}

type EnricherResult struct {
	Result  string `json:"result" jsonschema:"required,title=Enricher result" jsonschema_description:"Technical-channel payload — fully detailed enrichment report (or error explanation) returned to the calling agent for further reasoning. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Enricher result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with a short view of the enriched data. Written in the engagement language declared by your system prompt."`
}

type MemoristAction struct {
	Question  string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — question for complex search in previous work, tasks, and calls; include full context of what happened and what you want to find. Always written in English; never translated."`
	TaskID    *Int64 `json:"task_id,omitempty" jsonschema:"title=Task ID,type=integer" jsonschema_description:"If you know task id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional)"`
	SubtaskID *Int64 `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID,type=integer" jsonschema_description:"If you know subtask id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional)"`
	Message   string `json:"message" jsonschema:"required,title=Search message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the question and the path to reach the goal. Written in the engagement language declared by your system prompt."`
}

type MemoristResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search in long-term memory result" jsonschema_description:"Technical-channel payload — fully detailed long-term memory search report (or error explanation) returned to the calling agent for further reasoning. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Search in long-term memory result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with a short answer summary. Written in the engagement language declared by your system prompt."`
}

type SearchInMemoryAction struct {
	Questions Strings `json:"questions" jsonschema:"required,type=array,minItems=1,maxItems=5" jsonschema_description:"Technical-channel payload — 1 to 5 detailed, context-rich semantic queries against the team's long-term vector store. Must be a real JSON array of strings, e.g. [\"query 1\",\"query 2\"] - NOT a JSON-encoded string containing an array. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements, so non-English queries will fail to retrieve relevant stored knowledge. Each query should provide context, intent, and specific details with descriptive phrases, synonyms, and related terms; multiple queries explore different semantic angles. Note: If TaskID or SubtaskID are provided, they will be used as strict filters in the search."`
	TaskID    *Int64  `json:"task_id,omitempty" jsonschema:"title=Task ID" jsonschema_description:"Optional. The Task ID to use as a strict filter, retrieving information specifically related to this task. Used to enhance relevance by narrowing down the search scope. Type: integer."`
	SubtaskID *Int64  `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"Optional. The Subtask ID to use as a strict filter, retrieving information specifically related to this subtask. Helps in refining search results for increased relevancy. Type: integer."`
	Message   string  `json:"message" jsonschema:"required,title=Search-in-memory message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the queries or the information retrieval process. Written in the engagement language declared by your system prompt."`
}

type SearchGuideAction struct {
	Questions Strings `json:"questions" jsonschema:"required,type=array,minItems=1,maxItems=5" jsonschema_description:"Technical-channel payload — 1 to 5 detailed, context-rich semantic queries for the team's guide vector store. Must be a real JSON array of strings, e.g. [\"query 1\",\"query 2\"] - NOT a JSON-encoded string containing an array. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements, so non-English queries will fail to retrieve relevant guides. Each query should include scenario context, objectives, and specific intent. Note: The 'Type' field acts as a strict filter."`
	Type      String  `json:"type" jsonschema:"required,type=string,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other" jsonschema_description:"The specific type of guide you need. This required field acts as a strict filter to enhance the relevance of search results by narrowing down the scope to the specified guide type."`
	Message   string  `json:"message" jsonschema:"required,title=Guide search message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the queries and the type of guide needed. Written in the engagement language declared by your system prompt."`
}

type StoreGuideAction struct {
	Guide    string `json:"guide" jsonschema:"required" jsonschema_description:"Technical-channel payload — ready guide in markdown format that will be stored in the team's vector store for future retrieval. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements; non-English content becomes unreachable to future searches. Anonymize all sensitive data (IPs, domains, credentials, paths) using descriptive placeholders."`
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — question that was used to prepare this guide; co-indexed with the guide. Always written in English; never translated."`
	Type     String `json:"type" jsonschema:"required,type=string,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other" jsonschema_description:"Type of the guide to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store guide message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the guide. Written in the engagement language declared by your system prompt."`
}

type SearchAnswerAction struct {
	Questions Strings `json:"questions" jsonschema:"required,type=array,minItems=1,maxItems=5" jsonschema_description:"Technical-channel payload — 1 to 5 detailed, context-rich semantic queries for the team's answer vector store. Must be a real JSON array of strings, e.g. [\"query 1\",\"query 2\"] - NOT a JSON-encoded string containing an array. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements, so non-English queries will fail to retrieve relevant answers. Each query should include the context, what you want to find, what you intend to do with the information, and why you need it. Note: The 'Type' field acts as a strict filter."`
	Type      String  `json:"type" jsonschema:"required,type=string,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other" jsonschema_description:"The specific type of information or answer you are seeking. This required field acts as a strict filter to enhance the relevance of search results by narrowing down the scope to the specified type."`
	Message   string  `json:"message" jsonschema:"required,title=Answer search message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the queries and the type of answer needed. Written in the engagement language declared by your system prompt."`
}

type StoreAnswerAction struct {
	Answer   string `json:"answer" jsonschema:"required" jsonschema_description:"Technical-channel payload — ready answer in markdown format that will be stored in the team's vector store for future retrieval. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements; non-English content becomes unreachable to future searches. Anonymize all sensitive data (IPs, domains, credentials) using descriptive placeholders."`
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — question that was used to prepare this answer; co-indexed with the answer. Always written in English; never translated."`
	Type     String `json:"type" jsonschema:"required,type=string,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other" jsonschema_description:"Type of the search query and answer to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store answer message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the answer. Written in the engagement language declared by your system prompt."`
}

type SearchCodeAction struct {
	Questions Strings `json:"questions" jsonschema:"required,type=array,minItems=1,maxItems=5" jsonschema_description:"Technical-channel payload — 1 to 5 detailed, context-rich semantic queries for the team's code vector store. Must be a real JSON array of strings, e.g. [\"query 1\",\"query 2\"] - NOT a JSON-encoded string containing an array. ALWAYS written in English regardless of the engagement language: the store is indexed in English and shared across all engagements, so non-English queries will fail to retrieve relevant code samples. Each query should include the context, what you intend to achieve with the code, and the functionality or content that should be included."`
	Lang      string  `json:"lang" jsonschema:"required" jsonschema_description:"The programming language of the code samples you need. Use the standard markdown code block language name (e.g., 'python', 'bash', 'golang'). This required field narrows down the search to code samples in the desired language."`
	Message   string  `json:"message" jsonschema:"required,title=Code search message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the queries and the programming language of the code samples. Written in the engagement language declared by your system prompt."`
}

type StoreCodeAction struct {
	Code        string `json:"code" jsonschema:"required" jsonschema_description:"Ready code sample that will be stored for future retrieval (raw source code, not a localized message). Anonymize all sensitive data (IPs, domains, credentials, API keys) using descriptive placeholders."`
	Question    string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — question that was used to prepare or to write this code; co-indexed with the code in the team's vector store. Always written in English; never translated."`
	Lang        string `json:"lang" jsonschema:"required" jsonschema_description:"Programming language of the code sample; use markdown code block language name like python or bash or golang etc."`
	Explanation string `json:"explanation" jsonschema:"required" jsonschema_description:"Technical-channel payload — fully detailed explanation of the code sample (what it does, how it works, why it is useful, libraries/tools used). ALWAYS written in English; the explanation is co-indexed with the code in the team's vector store and shared across all engagements; non-English content becomes unreachable to future searches."`
	Description string `json:"description" jsonschema:"required" jsonschema_description:"Technical-channel payload — short description of the code sample as a summary of the explanation; co-indexed in the team's vector store. Always written in English; never translated."`
	Message     string `json:"message" jsonschema:"required,title=Store code message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the code sample. Written in the engagement language declared by your system prompt."`
}

type MaintenanceAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — task description for the DevOps team member to maintain the local environment and tools inside the docker container. Always written in English; never translated."`
	Message  string `json:"message" jsonschema:"required,title=Maintenance task message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the maintenance task. Written in the engagement language declared by your system prompt."`
}

type MaintenanceResult struct {
	Result  string `json:"result" jsonschema:"required,title=Maintenance result description" jsonschema_description:"Technical-channel payload — fully detailed maintenance report (or error explanation) returned to the calling agent, with usage instructions for the result. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Maintenance result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with the result and the path to reach the goal. Written in the engagement language declared by your system prompt."`
}

type CoderAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — task description for the developer team member to write code, with detailed explanation of the goal and how to achieve it when not obvious. Always written in English; never translated."`
	Message  string `json:"message" jsonschema:"required,title=Coder action message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the question and the task. Written in the engagement language declared by your system prompt."`
}

type CodeResult struct {
	Result  string `json:"result" jsonschema:"required,title=Code result description" jsonschema_description:"Technical-channel payload — fully detailed code-development report (or error explanation) returned to the calling agent, with usage instructions for the result. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Code result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with the result and the path to reach the goal. Written in the engagement language declared by your system prompt."`
}

type PentesterAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Technical-channel payload — task description for the pentester team member to perform a penetration test on the local environment and find vulnerabilities and weaknesses in the remote target. Always written in English; never translated."`
	Message  string `json:"message" jsonschema:"required,title=Pentester action message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary summarizing the question and the task. Written in the engagement language declared by your system prompt."`
}

type HackResult struct {
	Result  string `json:"result" jsonschema:"required,title=Hack result description" jsonschema_description:"Technical-channel payload — fully detailed penetration-test report (or error explanation) returned to the calling agent, with usage instructions for the result. Always written in English; never translated."`
	Message string `json:"message" jsonschema:"required,title=Hack result message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary with the result and the path to reach the goal. Written in the engagement language declared by your system prompt."`
}

// FlowStatusDetail controls the level of detail returned by get_flow_status.
// It is a type alias for String - see the FileOp comment above.
type FlowStatusDetail = String

const (
	FlowStatusDetailSummary  FlowStatusDetail = "summary"
	FlowStatusDetailTasks    FlowStatusDetail = "tasks"
	FlowStatusDetailSubtasks FlowStatusDetail = "subtasks"
	FlowStatusDetailRunning  FlowStatusDetail = "running"
	FlowStatusDetailPlanned  FlowStatusDetail = "planned"
)

// GetFlowStatusAction defines arguments for the get_flow_status tool.
type GetFlowStatusAction struct {
	Detail  FlowStatusDetail `json:"detail" jsonschema:"required,type=string,enum=summary,enum=tasks,enum=subtasks,enum=running,enum=planned" jsonschema_description:"Level of detail: 'summary' - flow health snapshot with status and counts; 'tasks' - all tasks with ID/status/title; 'subtasks' - all subtasks optionally filtered by task_id; 'running' - full Task→Subtask execution chain including task input and recent agent messages; 'planned' - only subtasks with status 'created' (not yet started), optionally filtered by task_id"`
	TaskID  *Int64           `json:"task_id,omitempty" jsonschema:"title=Task ID,type=integer" jsonschema_description:"Optional task ID filter. Applies to detail=subtasks and detail=planned to narrow results to a specific task."`
	Verbose Bool             `json:"verbose,omitempty" jsonschema:"type=boolean" jsonschema_description:"Set to true for deeper investigation: includes descriptions, inputs, results, and execution context per entry; shows up to 50 recent agent messages instead of 10."`
	Message string           `json:"message" jsonschema:"required,title=Status message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary describing what status information you are requesting. Written in the engagement language declared by your system prompt."`
}

// StopFlowAction defines arguments for the stop_flow tool.
type StopFlowAction struct {
	Reason  string `json:"reason" jsonschema:"required" jsonschema_description:"Engagement-log entry — brief explanation of why the engagement is being halted; surfaced in the engagement record. Written in the engagement language declared by your system prompt."`
	Message string `json:"message" jsonschema:"required,title=Stop message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining that the automation is being stopped. Written in the engagement language declared by your system prompt."`
}

// SubmitFlowInputAction defines arguments for the submit_flow_input tool.
type SubmitFlowInputAction struct {
	Input   string `json:"input" jsonschema:"required" jsonschema_description:"Engagement-log entry submitted on behalf of the engagement coordinator into the running automation. If a subtask is waiting at an 'ask' checkpoint, this is delivered as the coordinator's answer and execution resumes. If the flow is waiting with no active subtask, this becomes the goal for a new task — include complete context, targets, and constraints because the generator will decompose it into subtasks without further clarification. Written in the engagement language declared by your system prompt so downstream agents and the engagement record stay consistent."`
	Message string `json:"message" jsonschema:"required,title=Input message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary describing what is being submitted to the automation. Written in the engagement language declared by your system prompt."`
}

// WaitFlowCompletionAction defines arguments for the wait_flow_completion tool.
type WaitFlowCompletionAction struct {
	Timeout Int64  `json:"timeout" jsonschema:"required,type=integer" jsonschema_description:"How long to wait for the running automation task to finish, in seconds. Use 0 or a negative value to apply the default timeout of 60 seconds. Values above 3600 are capped at 3600 seconds (1 hour)."`
	Message string `json:"message" jsonschema:"required,title=Wait message" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary explaining why you are waiting for the automation to finish. Written in the engagement language declared by your system prompt."`
}

// PatchFlowSubtasksAction defines arguments for the patch_flow_subtasks tool.
type PatchFlowSubtasksAction struct {
	TaskID     int64              `json:"task_id" jsonschema:"required,type=integer" jsonschema_description:"ID of the task whose subtask plan to modify. Obtain this from get_flow_status with detail='tasks'."`
	Operations []SubtaskOperation `json:"operations" jsonschema:"required" jsonschema_description:"Delta operations to apply: add (insert new subtask at a position), remove (delete by ID), modify (update title/description), reorder (move to different position). Empty array returns the current plan unchanged. Each operation's title/description, when present, is an engagement-log plan entry (see operations)."`
	Message    string             `json:"message" jsonschema:"required,title=Patch summary" jsonschema_description:"Engagement-log entry — a 1-2 short sentence running commentary describing what changes are being made to the plan. Written in the engagement language declared by your system prompt."`
}

// ValidateSubtaskPatch validates the operations in a SubtaskPatch
func (sp SubtaskPatch) Validate() error {
	for i, op := range sp.Operations {
		switch op.Op {
		case SubtaskOpAdd:
			if op.Title == "" {
				return fmt.Errorf("operation %d: add requires title", i)
			}
			if op.Description == "" {
				return fmt.Errorf("operation %d: add requires description", i)
			}
		case SubtaskOpRemove:
			if op.ID == nil {
				return fmt.Errorf("operation %d: remove requires id", i)
			}
		case SubtaskOpModify:
			if op.ID == nil {
				return fmt.Errorf("operation %d: modify requires id", i)
			}
			if op.Title == "" && op.Description == "" {
				return fmt.Errorf("operation %d: modify requires at least title or description", i)
			}
		case SubtaskOpReorder:
			if op.ID == nil {
				return fmt.Errorf("operation %d: reorder requires id", i)
			}
		default:
			return fmt.Errorf("operation %d: unknown operation type %q", i, op.Op)
		}
	}
	return nil
}

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	sdata := strings.Trim(strings.ToLower(string(data)), "' \"\n\r\t")
	switch sdata {
	case "true":
		*b = true
	case "false":
		*b = false
	default:
		return fmt.Errorf("invalid bool value: %s", sdata)
	}
	return nil
}

func (b *Bool) MarshalJSON() ([]byte, error) {
	if b == nil || !*b {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}

func (b *Bool) Bool() bool {
	if b == nil {
		return false
	}
	return bool(*b)
}

func (b *Bool) String() string {
	if b == nil {
		return ""
	}
	return strconv.FormatBool(bool(*b))
}

type Int64 int64

func (i *Int64) UnmarshalJSON(data []byte) error {
	sdata := strings.Trim(strings.ToLower(string(data)), "' \"\n\r\t")
	if sdata == "" {
		*i = 0
		return nil
	}
	num, err := strconv.ParseInt(sdata, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid int value: %s", sdata)
	}
	*i = Int64(num)
	return nil
}

func (i *Int64) MarshalJSON() ([]byte, error) {
	if i == nil {
		return []byte("0"), nil
	}
	return []byte(strconv.FormatInt(int64(*i), 10)), nil
}

func (i *Int64) Int() int {
	if i == nil {
		return 0
	}
	return int(*i)
}

func (i *Int64) Int64() int64 {
	if i == nil {
		return 0
	}
	return int64(*i)
}

func (i *Int64) PtrInt64() *int64 {
	if i == nil {
		return nil
	}
	v := int64(*i)
	return &v
}

func (i *Int64) String() string {
	if i == nil {
		return ""
	}
	return strconv.FormatInt(int64(*i), 10)
}

// String is a lenient string for LLM-generated tool-call arguments, in
// particular enum-like fields (action/type/search_type/... - anywhere a
// value is compared or switched on rather than read as free text). Models
// occasionally wrap the value in an extra, literal pair of quote characters
// - e.g. the JSON string "\"write_file\"" decodes via a plain string
// UnmarshalJSON into the 12-character Go string `"write_file"` (quotes
// included), which then matches no known enum value, fails validation, and
// forces an unnecessary tool-call-fixer round-trip. UnmarshalJSON unwraps
// that redundant quoting instead of failing outright.
//
// FileOp, BrowserAction, SubtaskOperationType, and FlowStatusDetail are type
// aliases for String (not distinct types), so every field declared with one
// of those names gets this leniency automatically, with zero changes needed
// at any existing call site: switches, `==` comparisons, and assignments
// against their constants all keep working exactly as before, because the
// alias makes them the exact same type as String, not merely a similar one.
type String string

func (s *String) UnmarshalJSON(data []byte) error {
	// A bare JSON "null" unmarshals into an empty string with no error by
	// default, which would silently mask a required field being omitted -
	// treat it as invalid instead, consistent with Bool/Int64/Strings above.
	if trimmed := strings.TrimSpace(string(data)); trimmed == "null" {
		return fmt.Errorf("invalid string value: got null")
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = String(unwrapRedundantQuotes(raw))
	return nil
}

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// String implements fmt.Stringer, so %s/%v formatting and logrus fields show
// the unwrapped value, and callers that need a plain string (map keys,
// strings.* helpers, external struct fields, typed casts) can call it explicitly.
func (s String) String() string {
	return string(s)
}

// unwrapRedundantQuotes strips at most a few layers of a matching leading+
// trailing '"' pair from a decoded JSON string value. Bounded to guard
// against pathological input; the observed corruption is a single extra
// layer, but a small bound costs nothing and covers repeated wrapping too.
func unwrapRedundantQuotes(s string) string {
	for range 3 {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) < 2 || trimmed[0] != '"' || trimmed[len(trimmed)-1] != '"' {
			return s
		}
		s = trimmed[1 : len(trimmed)-1]
	}
	return s
}

// Strings is a lenient []string for LLM-generated tool-call arguments (e.g.
// the "questions" field of vector-store search tools). Models occasionally
// double-encode the array as a JSON string containing the array literal
// (e.g. "[\"a\", \"b\"]" instead of a real ["a","b"]), which fails a plain
// []string unmarshal with "cannot unmarshal string into ... []string" and
// forces an unnecessary tool-call-fixer round-trip. UnmarshalJSON recovers
// from that case instead of failing outright.
type Strings []string

func (s *Strings) UnmarshalJSON(data []byte) error {
	// A bare JSON "null" unmarshals into a nil slice with no error by default,
	// which would silently mask a required field being omitted - treat it as
	// invalid instead, consistent with Bool/Int64 above.
	if trimmed := strings.TrimSpace(string(data)); trimmed == "null" {
		return fmt.Errorf("invalid strings value: expected a JSON array of strings, got: null")
	}

	// Primary path: a real JSON array of strings - the common, schema-conforming case.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}

	// Recovery path: the array was sent as a JSON string. Try to decode its
	// content as a JSON array of strings first (the double-encoding case seen
	// in production), then fall back to treating the whole string as a single
	// element so one plain question doesn't fail either.
	var encoded string
	if err := json.Unmarshal(data, &encoded); err == nil {
		var nested []string
		if err := json.Unmarshal([]byte(encoded), &nested); err == nil {
			*s = nested
			return nil
		}
		if trimmed := strings.TrimSpace(encoded); trimmed != "" {
			*s = []string{trimmed}
			return nil
		}
	}

	return fmt.Errorf(
		"invalid strings value: expected a JSON array of strings, e.g. [\"question 1\",\"question 2\"], got: %s",
		strings.TrimSpace(string(data)),
	)
}

func (s *Strings) MarshalJSON() ([]byte, error) {
	if s == nil || *s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(*s))
}
