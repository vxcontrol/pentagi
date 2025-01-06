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
	Action  CodeAction `json:"action" jsonschema:"required,enum=read_file,enum=update_file,description=Action to perform with the code. 'read_file' - Returns the content of the file. 'update_file' - Updates the content of the file"`
	Content string     `json:"content" jsonschema:"description=Content to write to the file"`
	Path    string     `json:"path" jsonschema:"required,description=Path to the file to read or update"`
	Message string     `json:"message" jsonschema:"required,title=File action message,description=Not so long message which explain what do you want to read or to write to the file and explain written content to send to the user in user's language only"`
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
	Message string        `json:"message" jsonschema:"required,title=Task result message,description=Not so long message which explain what do you want to get, what format do you want to get and why do you need this to send to the user in user's language only"`
}

type SubtaskInfo struct {
	Title       string `json:"title" jsonschema:"required,title=Subtask title,description=Subtask title to show to the user which contains main goal of work result by this subtask"`
	Description string `json:"description" jsonschema:"required,title=Subtask to complete,description=Detailed description and instructions and rules and requirements what have to do in the subtask"`
}

type SubtaskList struct {
	Subtasks []SubtaskInfo `json:"subtasks" jsonschema:"required,title=Subtasks to complete,description=Ordered list of subtasks to execute after decomposing the task in the user language"`
	Message  string        `json:"message" jsonschema:"required,title=Subtask generation result,description=Not so long message with the generation result and main goal of work to send to the user in user's language only"`
}

type TaskResult struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,description=True if the task was executed successfully and the user task result was reached,type=boolean"`
	Result  string `json:"result" jsonschema:"required,title=Task result description,description=Fully detailed report or error message of the task or subtask result what was achieved or not (in user's language only)"`
	Message string `json:"message" jsonschema:"required,title=Task result message,description=Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type AskUser struct {
	Message string `json:"message" jsonschema:"required,title=Question for user,description=Question or any other information that should be sent to the user for clarifications in user's language only"`
}

type Done struct {
	Success Bool   `json:"success" jsonschema:"title=Execution result,description=True if the subtask was executed successfully and the user subtask result was reached,type=boolean"`
	Result  string `json:"result" jsonschema:"required,title=Task result description,description=Fully detailed report or error message of the subtask result what was achieved or not (in user's language only)"`
	Message string `json:"message" jsonschema:"required,title=Task result message,description=Not so long message with the result to send to the user in user's language only"`
}

type TerminalAction struct {
	Input   string `json:"input" jsonschema:"required,description=Command to be run in the docker container terminal according to rules to execute commands"`
	Cwd     string `json:"cwd" jsonschema:"required,description=Custom current working directory to execute commands in or default directory otherwise if it's not specified"`
	Detach  Bool   `json:"detach" jsonschema:"required,description=True if the command should be executed in the background, use timeout argument to limit of the execution time and you can not get output from the command if you use detach,type=boolean"`
	Timeout Int64  `json:"timeout" jsonschema:"required,description=Limit in seconds for command execution in terminal to prevent blocking of the agent and it depends on the specific command (minimum 10; maximum 1200; default 60),type=integer"`
	Message string `json:"message" jsonschema:"required,title=Terminal command message,description=Not so long message which explain what do you want to achieve and to execute in terminal to send to the user in user's language only"`
}

type AskAdvice struct {
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question with detailed information about issue to much better understand what's happend that should be sent to the mentor for clarifications in English"`
	Code     string `json:"code" jsonschema_description:"If your request related to code you may send snippet with relevant part of this"`
	Output   string `json:"output" jsonschema_description:"If your request related to terminal problem you may send stdout or stderr part of this"`
	Message  string `json:"message" jsonschema:"required,title=Ask advice message,description=Not so long message which explain what do you want to aks and solve and why do you need this and what do want to figure out to send to the user in user's language only"`
}

type ComplexSearch struct {
	Question string `json:"question" jsonschema:"required,description=Question to search by researcher team member in the internet and long-term memory with full explanation of what do you want to find and why do you need this in English"`
	Message  string `json:"message" jsonschema:"required,title=Search query message,description=Not so long message with the question to send to the user in user's language only"`
}

type SearchAction struct {
	Query      string `json:"query" jsonschema:"required,description=Query to search in the the specific search engine (e.g. google tavily traversaal serper etc.) Short and exact query is much better for better search result in English"`
	MaxResults Int64  `json:"max_results" jsonschema:"required,description=Maximum number of results to return (minimum 1; maximum 10; default 5),type=integer"`
	Message    string `json:"message" jsonschema:"required,title=Search query message,description=Not so long message with the expected result and path to reach goal to send to the user in user's language only"`
}

type SearchResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search result,description=Fully detailed report or error message of the search result and as a answer for the user question in English"`
	Message string `json:"message" jsonschema:"required,title=Search result message,description=Not so long message with the result and short answer to send to the user in user's language only"`
}

type EnricherResult struct {
	Result  string `json:"result" jsonschema:"required,title=Enricher result,description=Fully detailed report or error message what you can enriches of the user's question from different sources to take advice according to this data in English"`
	Message string `json:"message" jsonschema:"required,title=Enricher result message,description=Not so long message with the result and short view of the enriched data to send to the user in user's language only"`
}

type MemoristAction struct {
	Question  string `json:"question" jsonschema:"required,description=Question to complex search in the previous work and tasks and calls what kind information you need with full explanation context what was happened and what you want to find in English"`
	TaskID    *Int64 `json:"task_id,omitempty" jsonschema:"title=Task ID,description=If you know task id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional),type=integer"`
	SubtaskID *Int64 `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID,description=If you know subtask id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional),type=integer"`
	Message   string `json:"message" jsonschema:"required,title=Search message,description=Not so long message with the summary of the question to send and path to reach goal to the user in user's language only"`
}

type MemoristResult struct {
	Result  string `json:"result" jsonschema:"required,title=Search in long-term memory result,description=Fully detailed report or error message of the long-term memory search result and as a answer for the user question in English"`
	Message string `json:"message" jsonschema:"required,title=Search in long-term memory result message,description=Not so long message with the result and short answer to send to the user in user's language only"`
}

type SearchInMemoryAction struct {
	Question  string `json:"question" jsonschema:"required,description=Question to search in the vector database what kind information you need with full explanation context what was happened and what you want to find in English"`
	TaskID    *Int64 `json:"task_id,omitempty" jsonschema:"title=Task ID,description=If you know task id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional),type=integer"`
	SubtaskID *Int64 `json:"subtask_id,omitempty" jsonschema:"title=Subtask ID,description=If you know subtask id you can use it to get more relevant information from the vector database; it will be used as a hard filter for search (optional),type=integer"`
	Message   string `json:"message" jsonschema:"required,title=Search in long-term memory message,description=Not so long message with the summary of the question to send and path to reach goal to the user in user's language only"`
}

type SearchGuideAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to search in the vector database what kind of guide you need with full explanation scenario what you want to do and to achieve for better search result in English"`
	Type     string `json:"type" jsonschema:"required,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other,description=Type of the guide what you need to get; it's a hard filter to get the most relevant guide"`
	Message  string `json:"message" jsonschema:"required,title=Search guide message,description=Not so long message with the question and type of the guide to send to the user in user's language only"`
}

type StoreGuideAction struct {
	Guide    string `json:"guide" jsonschema:"required,description=Ready guide to the question that will be stored as a guide in markdown format for future search in English"`
	Question string `json:"question" jsonschema:"required,description=Question to the guide which was used to prepare this guide in English"`
	Type     string `json:"type" jsonschema:"required,enum=install,enum=configure,enum=use,enum=pentest,enum=development,enum=other,description=Type of the guide what you need to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store guide message,description=Not so long message with the summary of the guide to send to the user in user's language only"`
}

type SearchAnswerAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to search in the vector database what kind of search and answer you need with full explanation context what you want to find and to do and why do you need this for better search result in English"`
	Type     string `json:"type" jsonschema:"required,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other,description=Type of the search query and answer what you need to find; it's a hard filter to get the most relevant answer"`
	Message  string `json:"message" jsonschema:"required,title=Search answer message,description=Not so long message with the question and type of the search query or answer to send to the user in user's language only"`
}

type StoreAnswerAction struct {
	Answer   string `json:"answer" jsonschema:"required,description=Ready answer to the question (search query) that will be stored as a answer in markdown format for future search in English"`
	Question string `json:"question" jsonschema:"required,description=Question to the answer which was used to prepare this answer in English"`
	Type     string `json:"type" jsonschema:"required,enum=guide,enum=vulnerability,enum=code,enum=tool,enum=other,description=Type of the search query and answer what you need to store; it will be used as a hard filter for search"`
	Message  string `json:"message" jsonschema:"required,title=Store answer message,description=Not so long message with the summary of the answer to send to the user in user's language only"`
}

type SearchCodeAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to search in the vector database what kind of code you need with full explanation context what you want to do by this code and what should be inside this code for better search result in English"`
	Lang     string `json:"lang" jsonschema:"required,description=Programming language of the code sample; use markdown code block language name like python or bash or golang etc."`
	Message  string `json:"message" jsonschema:"required,title=Search code message,description=Not so long message with the question and programming language of the code sample to send to the user in user's language only"`
}

type StoreCodeAction struct {
	Code        string `json:"code" jsonschema:"required,description=Ready code sample that will be stored as a code for future search"`
	Question    string `json:"question" jsonschema:"required,description=Question to the code which was used to prepare or to write this code in English"`
	Lang        string `json:"lang" jsonschema:"required,description=Programming language of the code sample; use markdown code block language name like python or bash or golang etc."`
	Explanation string `json:"explanation" jsonschema:"required,description=Fully detailed explanation of the code sample and what it does and how it works and why it's useful and list of libraries and tools used in English"`
	Description string `json:"description" jsonschema:"required,description=Short description of the code sample as a summary of explanation in English"`
	Message     string `json:"message" jsonschema:"required,title=Store code result message,description=Not so long message with the summary of the code sample to send to the user in user's language only"`
}

type MaintenanceAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to DevOps team member as a task to maintain local environment and tools inside the docker container in English"`
	Message  string `json:"message" jsonschema:"required,title=Maintenance task message,description=Not so long message with the task and question to maintain local environment to send to the user in user's language only"`
}

type MaintenanceResult struct {
	Result  string `json:"result" jsonschema:"required,title=Maintenance result description,description=Fully detailed report or error message of the maintenance result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Maintenance result message,description=Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type CoderAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to developer team member as a task to write a code for the specific task with detailed explanation of what do you want to achieve and how to do this if it's not obvious in English"`
	Message  string `json:"message" jsonschema:"required,title=Coder action message,description=Not so long message with the question and summary of the task to send to the user in user's language only"`
}

type CodeResult struct {
	Result  string `json:"result" jsonschema:"required,title=Code result description,description=Fully detailed report or error message of the writing code result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Code result message,description=Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type PentesterAction struct {
	Question string `json:"question" jsonschema:"required,description=Question to pentester team member as a task to perform a penetration test on the local environment and find vulnerabilities and weaknesses in the remote target in English"`
	Message  string `json:"message" jsonschema:"required,title=Pentester action message,description=Not so long message with the question and summary of the task to send to the user in user's language only"`
}

type HackResult struct {
	Result  string `json:"result" jsonschema:"required,title=Hack result description,description=Fully detailed report or error message of the penetration test result what was achieved or not with detailed explanation and guide how to use this result in English"`
	Message string `json:"message" jsonschema:"required,title=Hack result message,description=Not so long message with the result and path to reach goal to send to the user in user's language only"`
}

type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	sdata := strings.Trim(strings.ToLower(string(data)), "' \"\n\r\t")
	if sdata == "true" {
		*b = true
	} else if sdata == "false" {
		*b = false
	} else {
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
