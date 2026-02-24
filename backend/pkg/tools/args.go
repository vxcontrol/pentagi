package tools

import (
	"fmt"
	"strconv"
	"strings"
)

type CodeAction string

const (
	ReadFile   CodeAction = "read_file"
	UpdateFile CodeAction = "update_file"
)

type FileAction struct {
	Action  CodeAction `json:"action" jsonschema:"required,enum=read_file,enum=update_file" jsonschema_description:"Action to perform with the code. 'read_file' - Returns the content of the file. 'update_file' - Updates the content of the file"`
	Content string     `json:"content" jsonschema_description:"Content to write to the file"`
	Path    string     `json:"path" jsonschema:"required" jsonschema_description:"Path to the file to read or update"`
	Message string     `json:"message" jsonschema:"required,title=File action message" jsonschema_description:"Not so long message which explain what do you want to read or to write to the file and explain written content to send to the user in user's language only"`
}

type BrowserAction string

const (
	Markdown BrowserAction = "markdown"
	HTML     BrowserAction = "html"
	Links    BrowserAction = "links"
)

type Browser struct {
	Url     string        `json:"url" jsonschema:"required" jsonschema_description:"url to open in the browser"`
	Action  BrowserAction `json:"action" jsonschema:"required,enum=markdown,enum=html,enum=links" jsonschema_description:"action to perform in the browser. 'markdown' - Returns the content of the page in markdown format. 'html' - Returns the content of the page in html format. 'links' - Get the list of all URLs on the page to be used in later calls (e.g., open search results after the initial search lookup)."`
	Message string        `json:"message" jsonschema:"required,title=Task result message" jsonschema_description:"Not so long message which explain what do you want to get, what format do you want to get and why do you need this to send to the user in user's language only"`
}

type SubtaskInfo struct {
	Title       string `json:"title" jsonschema:"required,title=Subtask title" jsonschema_description:"Subtask title to show to the user which contains main goal of work result by this subtask"`
	Description string `json:"description" jsonschema:"required,title=Subtask to complete" jsonschema_description:"Detailed description and instructions and rules and requirements what have to do in the subtask"`
}

type SubtaskList struct {
	Subtasks []SubtaskInfo `json:"subtasks" jsonschema:"required,title=Subtasks to complete" jsonschema_description:"Ordered list of subtasks to execute after decomposing the task in the user language"`
	Message  string        `json:"message" jsonschema:"required,title=Subtask generation result" jsonschema_description:"Not so long message with the generation result and main goal of work to send to the user in user's language only"`
}

// SubtaskOperationType defines the type of operation to perform on a subtask
type SubtaskOperationType string

const (
	SubtaskOpAdd     SubtaskOperationType = "add"
	SubtaskOpRemove  SubtaskOperationType = "remove"
	SubtaskOpModify  SubtaskOperationType = "modify"
	SubtaskOpReorder SubtaskOperationType = "reorder"
)

// SubtaskOperation defines a single operation on the subtask list for delta-based refinement
type SubtaskOperation struct {
	Op          SubtaskOperationType `json:"op" jsonschema:"required,enum=add,enum=remove,enum=modify,enum=reorder" jsonschema_description:"Operation type: 'add' creates a new subtask, 'remove' deletes a subtask by ID, 'modify' updates title/description of existing subtask, 'reorder' moves a subtask to a different position"`
	ID          *int64               `json:"id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"ID of existing subtask (required for remove/modify/reorder operations)"`
	AfterID     *int64               `json:"after_id,omitempty" jsonschema:"title=Insert after ID" jsonschema_description:"For add/reorder: insert after this subtask ID (null/0 = insert at beginning)"`
	Title       string               `json:"title,omitempty" jsonschema:"title=New title" jsonschema_description:"New title (required for add, optional for modify)"`
	Description string               `json:"description,omitempty" jsonschema:"title=New description" jsonschema_description:"New description (required for add, optional for modify)"`
}

type SubtaskInfoPatch struct {
	ID int64 `json:"id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"ID of the subtask (populated by the system for existing subtasks)"`
	SubtaskInfo
}

// SubtaskPatch is the delta-based refinement output for modifying subtask lists
type SubtaskPatch struct {
	Operations []SubtaskOperation `json:"operations" jsonschema:"required" jsonschema_description:"List of operations to apply to the current subtask list. Empty array means no changes needed."`
	Message    string             `json:"message" jsonschema:"required,title=Refinement summary" jsonschema_description:"Summary of changes made and justification for modifications to send to the user in user's language only"`
}

type TaskResult struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,type=boolean" jsonschema_description:"True if the task was executed successfully and the user task result was reached"`
	Result  string `json:"result" jsonschema:"required,title=Task result description" jsonschema_description:"Fully detailed report or error message of the task or subtask result what was achieved or not (in user's language only)"`
	Message string `json:"message" jsonschema:"required,title=Task result message" jsonschema_description:"Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type AskUser struct {
	Message string `json:"message" jsonschema:"required,title=Question for user" jsonschema_description:"Question or any other information that should be sent to the user for clarifications in user's language only"`
}

type Done struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,type=boolean" jsonschema_description:"True if the subtask was executed successfully and the user subtask result was reached"`
	Result  string `json:"result" jsonschema:"required,title=Task result description" jsonschema_description:"Fully detailed report or error message of the subtask result what was achieved or not (in user's language only)"`
	Message string `json:"message" jsonschema:"required,title=Task result message" jsonschema_description:"Not so long message with the result to send to the user in user's language only"`
}

type TerminalAction struct {
	Input   string `json:"input" jsonschema:"required" jsonschema_description:"Command to be run in the docker container terminal according to rules to execute commands"`
	Cwd     string `json:"cwd" jsonschema:"required" jsonschema_description:"Custom current working directory to execute commands in or default directory otherwise if it's not specified"`
	Detach  Bool   `json:"detach" jsonschema:"required,type=boolean" jsonschema_description:"True if the command should be executed in the background, use timeout argument to limit of the execution time and you can not get output from the command if you use detach"`
	Timeout Int64  `json:"timeout" jsonschema:"required,type=integer" jsonschema_description:"Limit in seconds for command execution in terminal to prevent blocking of the agent and it depends on the specific command (minimum 10; maximum 1200; default 60)"`
	Message string `json:"message" jsonschema:"required,title=Terminal command message" jsonschema_description:"Not so long message which explain what do you want to achieve and to execute in terminal to send to the user in user's language only"`
}

type AskAdvice struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question with detailed information about issue to much better understand what's happend that should be sent to the mentor for clarifications in English"`
	Code     string `json:"code" jsonschema_description:"If your request related to code you may send snippet with relevant part of this"`
	Output   string `json:"output" jsonschema_description:"If your request related to terminal problem you may send stdout or stderr part of this"`
	Message  string `json:"message" jsonschema:"required,title=Ask advice message" jsonschema_description:"Not so long message which explain what do you want to aks and solve and why do you need this and what do want to figure out to send to the user in user's language only"`
}

type ComplexSearch struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to search by researcher team member in the internet and long-term memory with full explanation of what do you want to find and why do you need this in English"`
	Message  string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Not so long message with the question to send to the user in user's language only"`
}

type SearchAction struct {
	Query      string `json:"query" jsonschema:"required" jsonschema_description:"Query to search in the the specific search engine (e.g. google duckduckgo tavily traversaal perplexity serper etc.) Short and exact query is much better for better search result in English"`
	MaxResults Int64  `json:"max_results" jsonschema:"required,type=integer" jsonschema_description:"Maximum number of results to return (minimum 1; maximum 10; default 5)"`
	Message    string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Not so long message with the expected result and path to reach goal to send to the user in user's language only"`
}

type SearchResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search result" jsonschema_description:"Fully detailed report or error message of the search result and as a answer for the user question in English"`
	Message string `json:"message" jsonschema:"required,title=Search result message" jsonschema_description:"Not so long message with the result and short answer to send to the user in user's language only"`
}

type SploitusAction struct {
	Query       string `json:"query" jsonschema:"required" jsonschema_description:"Search query for Sploitus (e.g. 'ssh', 'apache 2.4', 'CVE-2021-44228'). Short and precise queries return the best results."`
	ExploitType string `json:"exploit_type,omitempty" jsonschema:"enum=exploits,enum=tools" jsonschema_description:"What to search for: 'exploits' (default) for exploit code and PoCs, 'tools' for offensive security tools"`
	Sort        string `json:"sort,omitempty" jsonschema:"enum=default,enum=date,enum=score" jsonschema_description:"Result ordering: 'default' (relevance), 'date' (newest first), 'score' (highest CVSS first)"`
	MaxResults  Int64  `json:"max_results" jsonschema:"required,type=integer" jsonschema_description:"Maximum number of results to return (minimum 1; maximum 25; default 10)"`
	Message     string `json:"message" jsonschema:"required,title=Search query message" jsonschema_description:"Not so long message with the expected result and path to reach goal to send to the user in user's language only"`
}

type GraphitiSearchAction struct {
	SearchType     string   `json:"search_type" jsonschema:"required,enum=temporal_window,enum=entity_relationships,enum=diverse_results,enum=episode_context,enum=successful_tools,enum=recent_context,enum=entity_by_label" jsonschema_description:"Type of search to perform: temporal_window (time-bounded search), entity_relationships (graph traversal from an entity), diverse_results (anti-redundancy search), episode_context (full agent reasoning and tool outputs), successful_tools (proven techniques), recent_context (latest findings), entity_by_label (type-specific entity search)"`
	Query          string   `json:"query" jsonschema:"required" jsonschema_description:"Natural language query describing what to search for in English"`
	MaxResults     *Int64   `json:"max_results,omitempty" jsonschema:"title=Maximum Results,type=integer" jsonschema_description:"Maximum number of results to return (default varies by search type)"`
	TimeStart      string   `json:"time_start,omitempty" jsonschema_description:"Start of time window (ISO 8601 format, required for temporal_window)"`
	TimeEnd        string   `json:"time_end,omitempty" jsonschema_description:"End of time window (ISO 8601 format, required for temporal_window)"`
	CenterNodeUUID string   `json:"center_node_uuid,omitempty" jsonschema_description:"UUID of entity to search from (required for entity_relationships)"`
	MaxDepth       *Int64   `json:"max_depth,omitempty" jsonschema:"title=Maximum Depth,type=integer" jsonschema_description:"Maximum graph traversal depth (default: 2, max: 3, for entity_relationships)"`
	NodeLabels     []string `json:"node_labels,omitempty" jsonschema_description:"Filter to specific node types (e.g., ['IP_ADDRESS', 'SERVICE', 'VULNERABILITY'])"`
	EdgeTypes      []string `json:"edge_types,omitempty" jsonschema_description:"Filter to specific relationship types (e.g., ['HAS_PORT', 'EXPLOITS'])"`
	DiversityLevel string   `json:"diversity_level,omitempty" jsonschema:"enum=low,enum=medium,enum=high" jsonschema_description:"How much diversity to prioritize (default: medium, for diverse_results)"`
	MinMentions    *Int64   `json:"min_mentions,omitempty" jsonschema:"title=Minimum Mentions,type=integer" jsonschema_description:"Minimum episode mentions (default: 2, for successful_tools)"`
	RecencyWindow  string   `json:"recency_window,omitempty" jsonschema:"enum=1h,enum=6h,enum=24h,enum=7d" jsonschema_description:"How far back to search (default: 24h, for recent_context)"`
	Message        string   `json:"message" jsonschema:"required,title=Search message" jsonschema_description:"Not so long message with the summary of the search query and expected results to send to the user in user's language only"`
}

type EnricherResult struct {
	Result  string `json:"result" jsonschema:"required,title=Enricher result" jsonschema_description:"Fully detailed report or error message what you can enriches of the user's question from different sources to take advice according to this data in English"`
	Message string `json:"message" jsonschema:"required,title=Enricher result message" jsonschema_description:"Not so long message with the result and short view of the enriched data to send to the user in user's language only"`
}

type MemoristAction struct {
	Question  string `json:"question" jsonschema:"required" jsonschema_description:"Question to complex search in the previous work and tasks and calls what kind information you need with full explanation context what was happened and what you want to find in English"`
	TaskID    *Int64 `json:"task_id,omitempty" jsonschema:"title=Task ID,type=integer" jsonschema_description:"If you know task id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional)"`
	SubtaskID *Int64 `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID,type=integer" jsonschema_description:"If you know subtask id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional)"`
	Message   string `json:"message" jsonschema:"required,title=Search message" jsonschema_description:"Not so long message with the summary of the question to send and path to reach goal to the user in user's language only"`
}

type MemoristResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search in long-term memory result" jsonschema_description:"Fully detailed report or error message of the long-term memory search result and as a answer for the user question in English"`
	Message string `json:"message" jsonschema:"required,title=Search in long-term memory result message" jsonschema_description:"Not so long message with the result and short answer to send to the user in user's language only"`
}

type SearchInMemoryAction struct {
	Question  string `json:"question" jsonschema:"required" jsonschema_description:"A detailed, context-rich natural language query describing the specific information you need to retrieve from the vector database. Provide sufficient context, intent, and specific details to optimize semantic search accuracy. Include descriptive phrases, synonyms, and related terms where appropriate. Note: If TaskID or SubtaskID are provided, they will be used as strict filters in the search."`
	TaskID    *Int64 `json:"task_id,omitempty" jsonschema:"title=Task ID" jsonschema_description:"Optional. The Task ID to use as a strict filter, retrieving information specifically related to this task. Used to enhance relevance by narrowing down the search scope. Type: integer."`
	SubtaskID *Int64 `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID" jsonschema_description:"Optional. The Subtask ID to use as a strict filter, retrieving information specifically related to this subtask. Helps in refining search results for increased relevancy. Type: integer."`
	Message   string `json:"message" jsonschema:"required,title=User-Facing Message" jsonschema_description:"A concise summary of the query or the information retrieval process to be presented to the user, in the user's language only. This message should guide the user towards their goal in a clear and approachable manner."`
}

type SearchGuideAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"A detailed, context-rich natural language query describing the specific guide you need. Include a full explanation of the scenario, your objectives, and what you aim to achieve. Incorporate sufficient context, intent, and specific details to enhance semantic search accuracy. Use descriptive phrases, synonyms, and related terms where appropriate. Formulate your query in English. Note: The 'Type' field acts as a strict filter to retrieve the most relevant guide."`
	Type     string `json:"type" jsonschema:"required,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other" jsonschema_description:"The specific type of guide you need. This required field acts as a strict filter to enhance the relevance of search results by narrowing down the scope to the specified guide type."`
	Message  string `json:"message" jsonschema:"required,title=User-Facing Guide Search Message" jsonschema_description:"A concise summary of your query and the type of guide needed, to be presented to the user in the user's language. This message should guide the user toward their goal in a clear and approachable manner."`
}

type StoreGuideAction struct {
	Guide    string `json:"guide" jsonschema:"required" jsonschema_description:"Ready guide to the question that will be stored as a guide in markdown format for future search in English"`
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to the guide which was used to prepare this guide in English"`
	Type     string `json:"type" jsonschema:"required,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other" jsonschema_description:"Type of the guide what you need to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store guide message" jsonschema_description:"Not so long message with the summary of the guide to send to the user in user's language only"`
}

type SearchAnswerAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"A detailed, context-rich natural language query describing the specific answer or information you need. Include a full explanation of the context, what you want to find, what you intend to do with the information, and why you need it. Incorporate sufficient context, intent, and specific details to enhance semantic search accuracy. Use descriptive phrases, synonyms, and related terms where appropriate. Formulate your query in English. Note: The 'Type' field acts as a strict filter to retrieve the most relevant answer."`
	Type     string `json:"type" jsonschema:"required,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other" jsonschema_description:"The specific type of information or answer you are seeking. This required field acts as a strict filter to enhance the relevance of search results by narrowing down the scope to the specified type."`
	Message  string `json:"message" jsonschema:"required,title=User-Facing Answer Search Message" jsonschema_description:"A concise summary of your query and the type of answer needed, to be presented to the user in the user's language. This message should guide the user toward their goal in a clear and approachable manner."`
}

type StoreAnswerAction struct {
	Answer   string `json:"answer" jsonschema:"required" jsonschema_description:"Ready answer to the question (search query) that will be stored as a answer in markdown format for future search in English"`
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to the answer which was used to prepare this answer in English"`
	Type     string `json:"type" jsonschema:"required,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other" jsonschema_description:"Type of the search query and answer what you need to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store answer message" jsonschema_description:"Not so long message with the summary of the answer to send to the user in user's language only"`
}

type SearchCodeAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"A detailed, context-rich natural language query describing the specific code sample you need. Include a full explanation of the context, what you intend to achieve with the code, and the functionality or content that should be included. Incorporate sufficient context, intent, and specific details to enhance semantic search accuracy. Use descriptive phrases, relevant terminology, and related concepts where appropriate. Formulate your query in English."`
	Lang     string `json:"lang" jsonschema:"required" jsonschema_description:"The programming language of the code sample you need. Use the standard markdown code block language name (e.g., 'python', 'bash', 'golang'). This required field narrows down the search to code samples in the desired language."`
	Message  string `json:"message" jsonschema:"required,title=User-Facing Code Search Message" jsonschema_description:"A concise summary of your query and the programming language of the code sample, to be presented to the user in the user's language. This message should guide the user toward their goal in a clear and approachable manner."`
}

type StoreCodeAction struct {
	Code        string `json:"code" jsonschema:"required" jsonschema_description:"Ready code sample that will be stored as a code for future search"`
	Question    string `json:"question" jsonschema:"required" jsonschema_description:"Question to the code which was used to prepare or to write this code in English"`
	Lang        string `json:"lang" jsonschema:"required" jsonschema_description:"Programming language of the code sample; use markdown code block language name like python or bash or golang etc."`
	Explanation string `json:"explanation" jsonschema:"required" jsonschema_description:"Fully detailed explanation of the code sample and what it does and how it works and why it's useful and list of libraries and tools used in English"`
	Description string `json:"description" jsonschema:"required" jsonschema_description:"Short description of the code sample as a summary of explanation in English"`
	Message     string `json:"message" jsonschema:"required,title=Store code result message" jsonschema_description:"Not so long message with the summary of the code sample to send to the user in user's language only"`
}

type MaintenanceAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to DevOps team member as a task to maintain local environment and tools inside the docker container in English"`
	Message  string `json:"message" jsonschema:"required,title=Maintenance task message" jsonschema_description:"Not so long message with the task and question to maintain local environment to send to the user in user's language only"`
}

type MaintenanceResult struct {
	Result  string `json:"result" jsonschema:"required,title=Maintenance result description" jsonschema_description:"Fully detailed report or error message of the maintenance result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Maintenance result message" jsonschema_description:"Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type CoderAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to developer team member as a task to write a code for the specific task with detailed explanation of what do you want to achieve and how to do this if it's not obvious in English"`
	Message  string `json:"message" jsonschema:"required,title=Coder action message" jsonschema_description:"Not so long message with the question and summary of the task to send to the user in user's language only"`
}

type CodeResult struct {
	Result  string `json:"result" jsonschema:"required,title=Code result description" jsonschema_description:"Fully detailed report or error message of the writing code result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Code result message" jsonschema_description:"Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type PentesterAction struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question to pentester team member as a task to perform a penetration test on the local environment and find vulnerabilities and weaknesses in the remote target in English"`
	Message  string `json:"message" jsonschema:"required,title=Pentester action message" jsonschema_description:"Not so long message with the question and summary of the task to send to the user in user's language only"`
}

type HackResult struct {
	Result  string `json:"result" jsonschema:"required,title=Hack result description" jsonschema_description:"Fully detailed report or error message of the penetration test result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Hack result message" jsonschema_description:"Not so long message with the result and path to reach goal to send to the user in user's language only"`
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
