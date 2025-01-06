package tools

import (
	"pentagi/pkg/database"

	"github.com/invopop/jsonschema"
	"github.com/tmc/langchaingo/llms"
)

const (
	FinalyToolName            = "done"
	AskUserToolName           = "ask"
	MaintenanceToolName       = "maintenance"
	MaintenanceResultToolName = "maintenance_result"
	CoderToolName             = "coder"
	CodeResultToolName        = "code_result"
	PentesterToolName         = "pentester"
	HackResultToolName        = "hack_result"
	AdviceToolName            = "advice"
	MemoristToolName          = "memorist"
	MemoristResultToolName    = "memorist_result"
	BrowserToolName           = "browser"
	GoogleToolName            = "google"
	TavilyToolName            = "tavily"
	TraversaalToolName        = "traversaal"
	SearchToolName            = "search"
	SearchResultToolName      = "search_result"
	EnricherResultToolName    = "enricher_result"
	SearchInMemoryToolName    = "search_in_memory"
	SearchGuideToolName       = "search_guide"
	StoreGuideToolName        = "store_guide"
	SearchAnswerToolName      = "search_answer"
	StoreAnswerToolName       = "store_answer"
	SearchCodeToolName        = "search_code"
	StoreCodeToolName         = "store_code"
	ReportResultToolName      = "report_result"
	SubtaskListToolName       = "subtask_list"
	TerminalToolName          = "terminal"
	FileToolName              = "file"
)

var reflector = &jsonschema.Reflector{
	DoNotReference: true,
	ExpandedStruct: true,
}

var allowedSummarizingToolsResult = []string{
	TerminalToolName,
	BrowserToolName,
}

var allowedStoringInMemoryTools = []string{
	TerminalToolName,
	FileToolName,
	SearchToolName,
	GoogleToolName,
	TavilyToolName,
	TraversaalToolName,
	MaintenanceToolName,
	CoderToolName,
	PentesterToolName,
	AdviceToolName,
}

var registryDefinitions = map[string]llms.FunctionDefinition{
	TerminalToolName: {
		Name: TerminalToolName,
		Description: "Calls a terminal command in blocking mode with hard limit timeout 1200 seconds and " +
			"optimum timeout 60 seconds, only one command can be executed at a time",
		Parameters: reflector.Reflect(&TerminalAction{}),
	},
	FileToolName: {
		Name:        FileToolName,
		Description: "Modifies or reads local files",
		Parameters:  reflector.Reflect(&FileAction{}),
	},
	ReportResultToolName: {
		Name:        ReportResultToolName,
		Description: "Send the report result to the user with execution status and description",
		Parameters:  reflector.Reflect(&TaskResult{}),
	},
	SubtaskListToolName: {
		Name:        SubtaskListToolName,
		Description: "Send new generated subtask list to the user",
		Parameters:  reflector.Reflect(&SubtaskList{}),
	},
	SearchToolName: {
		Name: SearchToolName,
		Description: "Search in a different search engines in the internet and long-term memory " +
			"by your complex question to the researcher team member, also you can add some instructions to get result " +
			"in a specific format or structure or content type like " +
			"code or command samples, manuals, guides, exploits, vulnerability details, repositories, libraries, etc.",
		Parameters: reflector.Reflect(&ComplexSearch{}),
	},
	SearchResultToolName: {
		Name:        SearchResultToolName,
		Description: "Send the complex search result as a answer for the user question to the user",
		Parameters:  reflector.Reflect(&SearchResult{}),
	},
	BrowserToolName: {
		Name:        BrowserToolName,
		Description: "Opens a browser to look for additional information from the web site",
		Parameters:  reflector.Reflect(&Browser{}),
	},
	GoogleToolName: {
		Name: GoogleToolName,
		Description: "Search in the google search engine, it's a fast query and the shortest content " +
			"to check some information or collect public links by short query",
		Parameters: reflector.Reflect(&SearchAction{}),
	},
	TavilyToolName: {
		Name: TavilyToolName,
		Description: "Search in the tavily search engine, it's a more complex query and more detailed content " +
			"with answer by query and detailed information from the web sites",
		Parameters: reflector.Reflect(&SearchAction{}),
	},
	TraversaalToolName: {
		Name: TraversaalToolName,
		Description: "Search in the traversaal search engine, presents you answer and web-links " +
			"by your query according to relevant information from the web sites",
		Parameters: reflector.Reflect(&SearchAction{}),
	},
	EnricherResultToolName: {
		Name:        EnricherResultToolName,
		Description: "Send the enriched user's question with additional information to the user",
		Parameters:  reflector.Reflect(&EnricherResult{}),
	},
	SearchInMemoryToolName: {
		Name:        SearchInMemoryToolName,
		Description: "Search in the vector database (long-term memory) some relevant information parts by the exact question or by the part of original information",
		Parameters:  reflector.Reflect(&SearchInMemoryAction{}),
	},
	SearchGuideToolName: {
		Name:        SearchGuideToolName,
		Description: "Search in the vector database some ready guides to use them as a base for your task or subtask or issue or question",
		Parameters:  reflector.Reflect(&SearchGuideAction{}),
	},
	StoreGuideToolName: {
		Name:        StoreGuideToolName,
		Description: "Store the guide to the vector database for future use",
		Parameters:  reflector.Reflect(&StoreGuideAction{}),
	},
	SearchAnswerToolName: {
		Name:        SearchAnswerToolName,
		Description: "Search in the vector database some ready answers to use them as a base for your task or subtask or issue or question",
		Parameters:  reflector.Reflect(&SearchAnswerAction{}),
	},
	StoreAnswerToolName: {
		Name:        StoreAnswerToolName,
		Description: "Store the question answer to the vector database for future use",
		Parameters:  reflector.Reflect(&StoreAnswerAction{}),
	},
	SearchCodeToolName: {
		Name:        SearchCodeToolName,
		Description: "Search in the vector database some ready code samples to use them as a base for your task or subtask or issue or question",
		Parameters:  reflector.Reflect(&SearchCodeAction{}),
	},
	StoreCodeToolName: {
		Name:        StoreCodeToolName,
		Description: "Store the code sample to the vector database for future use. It's should be a sample like a one source code file for some question",
		Parameters:  reflector.Reflect(&StoreCodeAction{}),
	},
	MemoristToolName: {
		Name:        MemoristToolName,
		Description: "Call to Archivist team member who remember all the information about the past work and made tasks and can answer your question about it",
		Parameters:  reflector.Reflect(&MemoristAction{}),
	},
	MemoristResultToolName: {
		Name:        MemoristResultToolName,
		Description: "Send the search in long-term memory result as a answer for the user question to the user",
		Parameters:  reflector.Reflect(&MemoristResult{}),
	},
	MaintenanceToolName: {
		Name:        MaintenanceToolName,
		Description: "Call to DevOps team member to maintain local environment and tools inside the docker container",
		Parameters:  reflector.Reflect(&MaintenanceAction{}),
	},
	MaintenanceResultToolName: {
		Name:        MaintenanceResultToolName,
		Description: "Send the maintenance result to the user with task status and fully detailed report about using the result",
		Parameters:  reflector.Reflect(&TaskResult{}),
	},
	CoderToolName: {
		Name:        CoderToolName,
		Description: "Call to developer team member to write a code for the specific task",
		Parameters:  reflector.Reflect(&CoderAction{}),
	},
	CodeResultToolName: {
		Name:        CodeResultToolName,
		Description: "Send the code result to the user with execution status and fully detailed report about using the result",
		Parameters:  reflector.Reflect(&CodeResult{}),
	},
	PentesterToolName: {
		Name:        PentesterToolName,
		Description: "Call to pentester team member to perform a penetration test or looking for vulnerabilities and weaknesses",
		Parameters:  reflector.Reflect(&PentesterAction{}),
	},
	HackResultToolName: {
		Name:        HackResultToolName,
		Description: "Send the penetration test result to the user with detailed report",
		Parameters:  reflector.Reflect(&HackResult{}),
	},
	AdviceToolName: {
		Name:        AdviceToolName,
		Description: "Get more complex answer from the mentor about some issue or difficult situation",
		Parameters:  reflector.Reflect(&AskAdvice{}),
	},
	AskUserToolName: {
		Name:        AskUserToolName,
		Description: "If you need to ask user for input, use this tool",
		Parameters:  reflector.Reflect(&AskUser{}),
	},
	FinalyToolName: {
		Name:        FinalyToolName,
		Description: "If you need to finish the task with success or failure, use this tool",
		Parameters:  reflector.Reflect(&Done{}),
	},
}

func getMessageType(name string) database.MsglogType {
	switch name {
	case TerminalToolName:
		return database.MsglogTypeTerminal
	case FileToolName:
		return database.MsglogTypeFile
	case BrowserToolName:
		return database.MsglogTypeBrowser
	case MemoristToolName, SearchToolName, GoogleToolName, TavilyToolName, TraversaalToolName,
		SearchGuideToolName, SearchAnswerToolName, SearchCodeToolName, SearchInMemoryToolName:
		return database.MsglogTypeSearch
	case AdviceToolName:
		return database.MsglogTypeAdvice
	case AskUserToolName:
		return database.MsglogTypeAsk
	case FinalyToolName:
		return database.MsglogTypeDone
	default:
		return database.MsglogTypeThoughts
	}
}

func getMessageResultFormat(name string) database.MsglogResultFormat {
	switch name {
	case TerminalToolName:
		return database.MsglogResultFormatTerminal
	case FileToolName, BrowserToolName:
		return database.MsglogResultFormatPlain
	default:
		return database.MsglogResultFormatMarkdown
	}
}
