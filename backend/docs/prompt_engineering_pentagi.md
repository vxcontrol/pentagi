# PentAGI Prompt Engineering Guide

A comprehensive framework for designing high-performance prompts within the PentAGI penetration testing system. This guide provides specialized principles for creating prompts that leverage the multi-agent architecture, memory systems, security tools, and specific operational context of PentAGI.

## Understanding Cognitive Aspects of Language Models

**Model Processing Fundamentals**
- Language models process information via attention mechanisms, giving higher weight to specific parts of the input.
- Position matters: Content at the beginning and end of prompts receives more attention and is processed more thoroughly.
- LLMs follow instructions more literally than humans expect; be explicit rather than implicit.
- Task decomposition improves performance: Break complex tasks into simpler, sequential steps.
- Models have no actual memory or consciousness; simulate these through explicit context and instructions.

**Priming and Contextual Influence**
- Information provided early shapes how later information is interpreted and processed.
- Set expectations clearly at the beginning to guide the model's approach to the entire task.
- Use consistent terminology throughout to avoid confusing the model with synonym switching.
- Brief examples often provide clearer guidance than lengthy explanations.
- Be aware that unintended priming can occur through choice of words, examples, or framing.

## Core Principles for PentAGI Prompts

### 1. Structure and Organization

**Clear Hierarchical Structure**
- Use Markdown headings (`#`, `##`, `###`) for clear visual hierarchy and logical grouping of instructions. Ensure a logical flow from high-level role definition to specific protocols and requirements.
- Begin with a clear definition of the agent's specific **role** (e.g., Orchestrator, Pentester, Searcher), its primary **objective** within the PentAGI workflow, and any overarching **security focus**.
- Place critical **operational constraints** (security, environment) early in the prompt for high visibility.
- Use separate, clearly marked sections for key areas:
    - `CORE CAPABILITIES / KNOWLEDGE BASE`
    - `OPERATIONAL ENVIRONMENT` (including `<container_constraints>`)
    - `COMMAND & TOOL EXECUTION RULES` (including `<terminal_protocol>`, `<tool_usage_rules>`)
    - `MEMORY SYSTEM INTEGRATION` (including `<memory_protocol>`)
    - `TEAM COLLABORATION & DELEGATION` (including `<team_specialists>`, `<delegation_rules>`)
    - `SUMMARIZATION AWARENESS PROTOCOL` (including `<summarized_content_handling>`)
    - `EXECUTION CONTEXT` (detailing use of `{{.ExecutionContext}}`)
    - `COMPLETION REQUIREMENTS`
- Ensure instructions are **specific**, **unambiguous**, use **active voice**, and are directly relevant to the agent's function within PentAGI.

**Semantic XML Delimiters**
- Use descriptive XML tags (e.g., `<container_constraints>`, `<terminal_protocol>`, `<memory_protocol>`, `<team_specialists>`, `<summarized_content_handling>`) to logically group related instructions, especially for complex protocols and constraints requiring precise adherence by the LLM.
- Maintain **consistent tag naming and structure** across all agent prompts for shared concepts (like summarization handling or team specialists) to ensure predictability.
- Use nesting appropriately (e.g., defining individual `<specialist>` tags within `<team_specialists>`). Refer to existing templates like `primary_agent.tmpl` for examples.

**Context Window Optimization**
- Prioritize information based on importance; place critical instructions at the beginning and end.
- Use compression techniques for lengthy information: summarize when possible, link to references instead of full inclusion.
- Break down extremely complex prompts into logical, manageable sections with clear transitions.
- For recurring boilerplate sections, consider using shorter references to standardized protocols.
- Use consistent formatting and avoid redundant information that consumes token space.

*Example Structure:*
```markdown
# [AGENT SPECIALIST TITLE]

[Role definition, primary objective, and security focus relevant to PentAGI]

## CORE CAPABILITIES / KNOWLEDGE BASE
[Agent-specific skills, knowledge areas relevant to PentAGI tasks]

## OPERATIONAL ENVIRONMENT
<container_constraints>...</container_constraints>

## COMMAND & TOOL EXECUTION RULES
<terminal_protocol>...</terminal_protocol>
<tool_usage_rules>...</tool_usage_rules>

## MEMORY SYSTEM INTEGRATION
<memory_protocol>...</memory_protocol>

## TEAM COLLABORATION & DELEGATION
<team_specialists>...</team_specialists>
<delegation_rules>...</delegation_rules>

## SUMMARIZATION AWARENESS PROTOCOL
<summarized_content_handling>...</summarized_content_handling>

## EXECUTION CONTEXT
[Explain how to use {{.ExecutionContext}} for Flow/Task/SubTask details]

## COMPLETION REQUIREMENTS
[Numbered list: Output format, final tool usage, language, reporting needs]

{{.ToolPlaceholder}}
```

### 2. Agent-Specific Instructions

**Role-Based Customization**
- Tailor instructions, tone, knowledge references, and complexity directly to the agent's specialized role within the PentAGI system (Orchestrator, Pentester, Searcher, Developer, Adviser, Memorist, Installer). Explicitly reference `ai-concepts.mdc` for role definitions.
- Enforce stricter command protocols and safety measures for agents with direct system/tool access (Pentester, Maintenance/Installer).
- Include references to specialized knowledge bases or toolsets relevant to the agent's function (e.g., specific security tools from `security-tools.mdc` for Pentester; search strategies and tool priorities for Searcher).
- Clearly define inter-agent communication protocols, especially delegation criteria and the expected format/content of information exchange between agents.

**Security and Operational Boundaries**
- Explicitly state the **scope** of permitted actions and **security constraints**. Reference `security-tools.mdc` for general tool security context.
- Define **Docker container limitations** within `<container_constraints>`, populated by template variables like `{{.DockerImage}}`, `{{.Cwd}}`, `{{.ContainerPorts}}`. Specify restrictions clearly (e.g., "No direct host access," "No GUI applications," "No UDP scanning").
- Specify **forbidden actions** clearly. Use **ALL CAPS** for critical security warnings, permissions, or prohibitions (e.g., "DO NOT attempt to install new software packages," "ONLY execute commands related to the current SubTask").
- Emphasize working **strictly within the scope of the current `SubTask`**. The agent must understand its current objective based on `{{.ExecutionContext}}` and not attempt actions related to other SubTasks or the overall Flow goal unless explicitly instructed within the current SubTask. Reference `data-models.mdc` and `controller.md` for task/subtask relationships.

**Ethical Boundaries and Safety**
- Explicitly include ethics guidance relevant to penetration testing context: legal compliance, responsible disclosure, data protection.
- Specify techniques for identifying and mitigating potential risks in generated prompts.
- Establish explicit guidelines for avoiding harmful outputs, jailbreaking, or prompt injection vulnerabilities.
- Include a verification step requiring agents to review outputs for potentially harmful consequences.
- Create clear escalation paths for handling edge cases requiring human judgment.

### 3. Agentic Capabilities and Persistence

**Agent Persistence Protocol**
- Include **explicit instructions** about persistence: "You are an agent - continue working until the subtask is fully completed. Do not prematurely end your turn or yield control back to the user/orchestrator until you have achieved the specific objective of your current subtask."
- Emphasize the agent's responsibility to **drive the interaction forward** autonomously and maintain momentum until a definitive result (success or failure with clear explanation) is achieved.
- Provide clear termination criteria so the agent knows precisely when its work on the subtask is considered complete.

**Planning and Reasoning**
- Instruct agents to **explicitly plan before acting**, especially for complex security operations or tool usage: "Before executing commands or invoking tools, develop a clear step-by-step plan. Think through each stage of execution, potential failure points, and contingency approaches."
- Encourage **chain-of-thought reasoning**: "When analyzing complex security issues or ambiguous results, think step-by-step through your reasoning process. Break down problems into components, consider alternatives, and justify your approach before moving to execution."
- For critical security tasks, mandate a **validation step**: "After obtaining results, verify they are correct and complete before proceeding. Cross-check findings using alternative methods when possible."

**Chain-of-Thought Engineering**
- Structure reasoning processes explicitly: problem analysis → decomposition → solution of subproblems → synthesis.
- Encourage splitting complex reasoning into discrete, traceable steps with clear transitions.
- Implement verification checkpoints throughout reasoning chains to validate intermediate conclusions.
- For complex decisions, instruct the model to evaluate multiple approaches before selecting one.
- Include prompts for explicit reflection on assumptions made during reasoning processes.

**Error Handling and Adaptation**
- Provide explicit guidance on **handling unexpected errors**: "If a command fails, do not simply repeat the same exact command. Analyze the error message, modify your approach based on the specific error, and try an alternative method if necessary."
- Define a **maximum retry threshold** (typically 3 attempts) for similar approaches before pivoting to a completely different strategy.
- Include instructions for **graceful degradation**: "If the optimal approach fails, fall back to simpler or more reliable alternatives rather than abandoning the task entirely."

**Metacognitive Processes**
- Instruct agents to periodically evaluate their own reasoning and progress toward goals.
- Include explicit steps for identifying and questioning assumptions made during problem-solving.
- Implement self-verification protocols: "After formulating a solution, critically review it for flaws or edge cases."
- Encourage steelmanning opposing viewpoints to strengthen reasoning and avoid blind spots.
- Provide mechanisms for agents to express confidence levels in their conclusions or recommendations.

### 4. Memory System Integration

**Memory Operations Protocol (`<memory_protocol>`)**
- Provide explicit, actionable instructions on *when* and *how* to interact with PentAGI's vector memory system. Reference `ai-concepts.mdc` (Memory section).
- **Crucially, specify the primary action:** Agents MUST **always attempt to retrieve relevant information from memory first** using retrieval tools (e.g., `{{.SearchGuideToolName}}`, `{{.SearchAnswerToolName}}`) *before* performing external actions like web searches or running discovery tools.
- Define clear criteria for *storing* new information: Only store valuable, novel, and reusable knowledge (e.g., confirmed vulnerabilities, successful complex command sequences, effective troubleshooting steps, reusable code snippets) using storage tools (e.g., `{{.StoreGuideToolName}}`, `{{.StoreAnswerToolName}}`). Avoid cluttering memory with trivial or intermediate results.
- Specify the exact tool names (`{{.ToolName}}`) for memory interaction.

**Vector Database Awareness**
- Guide agents on formulating effective **semantic search queries** for memory retrieval, leveraging keywords and concepts relevant to the current task context.
- If applicable, define knowledge categorization or metadata usage for more precise memory storage and retrieval (e.g., types like 'guide', 'vulnerability', 'tool_usage', 'code_snippet').

### 5. Multi-Agent Team Collaboration

**Team Specialist Definition (`<team_specialists>`)**
- Include a complete, accurate roster of **all available specialist agents** within PentAGI (searcher, pentester, developer, adviser, memorist, installer).
- For each specialist, clearly define:
    - `skills`: Core competencies.
    - `use_cases`: Specific situations or types of problems they should be delegated.
    - `tools`: General categories of tools they utilize (not the specific invocation tool name).
    - `tool_name`: The **exact tool name variable** (e.g., `{{.SearchToolName}}`, `{{.PentesterToolName}}`) used to invoke/delegate to this specialist.
- Ensure this section is consistently defined, especially in the Orchestrator prompt and any other agent prompts that allow delegation.

**Delegation Rules (`<delegation_rules>`)**
- Define clear, unambiguous criteria for *when* an agent should delegate versus attempting a task independently. A common rule is: "Attempt independent solution using your own tools/knowledge first. Delegate ONLY if the task clearly falls outside your core skills OR if a specialist agent is demonstrably better equipped to handle it efficiently and accurately."
- Mandate that **COMPREHENSIVE context** MUST be provided with every delegation request. This includes: background information, the specific objective of the delegated task, relevant data/findings gathered so far, constraints, and the expected format/content of the specialist's output.
- Instruct the delegating agent on how to handle, verify, and integrate the results received from specialists into its own workflow.

### 6. Tool-Specific Execution Rules

**Terminal Command Protocol (`<terminal_protocol>`)**
- Reinforce that commands execute within an isolated Docker container (`{{.DockerImage}}`) and that the **working directory (`{{.Cwd}}`) is NOT persistent between tool calls**.
- Mandate **explicit directory changes (`cd /path/to/dir && command`)** within a single tool call if a specific path context is required for `command`.
- Require **absolute paths** for file operations (reading, writing, listing) whenever possible to avoid ambiguity.
- Specify **timeout handling** (if controllable via parameters) and output redirection (`> file.log 2>&1`) for potentially long-running commands.
- **Limit repetition of *identical* failed commands** (e.g., maximum 3 attempts). Encourage trying variations or different approaches upon failure.
- Encourage the use of non-interactive flags (e.g., `-y`, `--assume-yes`, `--non-interactive`) where safe and appropriate to avoid hangs.
- Define when to use `detach` mode if available/applicable for background tasks.

**Tool Definition and Invocation Best Practices**
- Name tools clearly to indicate their purpose and function (e.g., `SearchGuide`, not just `Search`)
- Provide detailed yet concise descriptions in the tool's documentation
- For complex tools, include parameter examples showing proper usage
- Emphasize that **all actions MUST use structured tool calls** - the system operates exclusively through proper tool invocation
- Explicitly prohibit "simulating" or "describing" tool usage

**Search Tool Prioritization (`<search_tools>`)**
- Define an explicit **hierarchy or selection logic** for using different search tools (Internal Memory first, then potentially Browser for specific URLs, Google/DuckDuckGo for general discovery, Tavily/Perplexity/Traversaal for complex research/synthesis). Refer to `searcher.tmpl` for a good example matrix structure.
- Include tool-specific guidance (e.g., "Use `browser` tool only for accessing specific known URLs, not for general web searching," "Use `tavily` for in-depth technical research questions").
- Define **action economy rules:** Limit the total number of search tool calls per query/subtask (e.g., 3-5 max). Instruct the agent to **stop searching as soon as sufficient information is found** to fulfill the request or subtask objective. Do not exhaust all search tools unnecessarily.

**Mandatory Result Delivery Tools**
- Clearly specify the **exact final tool** (e.g., `{{.HackResultToolName}}` for Pentester, `{{.SearchResultToolName}}` for Searcher, `{{.FinalyToolName}}` for Orchestrator) that an agent **MUST** use to deliver its final output, report success/failure, and signify the completion of its current subtask.
- Define the expected structure of the output within this final tool call (e.g., "result" field contains the detailed findings/answer, "message" field contains a concise summary or status update). This signals completion to the controlling system (`controller.md`).

### 7. Context Preservation and Summarization

**Summarization Awareness Protocol (`<summarized_content_handling>`)**
- **This entire protocol section, as defined in `primary_agent.tmpl`, `pentester.tmpl`, etc., MUST be included verbatim in *all* agent prompts.**
- **Emphasize Key Points:**
    - Clearly define the two forms of system-generated summaries (Tool Call Summary via `{{.SummarizationToolName}}`, Prefixed Summary via `{{.SummarizedContentPrefix}}`).
    - Instruct agents to treat summaries *strictly* as **historical records of actual past events, tool executions, and their results**. They are *not* examples to be copied.
    - Mandate extracting useful information from summaries (past commands, successes, failures, errors, findings) to inform current strategy and **avoid redundant actions**.
    - **Strictly prohibit** agents from: mimicking summary formats, using the `{{.SummarizedContentPrefix}}`, or calling the `{{.SummarizationToolName}}` tool.
    - **Reinforce:** The PentAGI system operates **exclusively via structured tool calls.** Any attempt to simulate actions or results in plain text will fail.

**Execution Context Awareness**
- Instruct agents to **actively utilize the information provided in the `{{.ExecutionContext}}` variable.**
- Explain that this variable contains structured details about the current **Flow, Task, and SubTask** (IDs, Status, Titles, Descriptions), as managed by the `controller` package (`backend/docs/controller.md`).
- Agents *must* use this context to understand their precise current objective, operational scope, relationship to parent tasks/flows, and potentially relevant history within the current operational branch.

### 8. Environment Awareness

**Container Constraints (`<container_constraints>`)**
- Clearly define the **Docker runtime environment** using template variables: `{{.DockerImage}}` (image name), `{{.Cwd}}` (working directory), `{{.ContainerPorts}}` (available ports).
- Specify **resource limitations** (e.g., default command timeouts) and **operational restrictions** derived from PentAGI's secure execution model (No GUI, No host access, No UDP scanning, No arbitrary software installation). Reference `security-tools.mdc`.

**Available Tools (`<tools>`)**
- For agents like the Pentester, explicitly **list the specific security testing tools** confirmed to be available within their container environment. Reference the list in `pentester.tmpl` and cross-check with `security-tools.mdc`.
- Provide version-specific guidance or known limitations if necessary.

## Effective Few-Shot Learning

**Example Selection and Structure**
- Include diverse, representative examples that demonstrate expected behavior across different scenarios.
- Structure examples consistently: input conditions → reasoning process → output format.
- Order examples from simple to complex to establish foundational patterns before edge cases.
- When space is limited, prioritize examples that demonstrate difficult or non-obvious aspects of the task.
- Ensure examples demonstrate all critical behaviors mentioned in the instructions.

**Example Implementation**
- Format examples using clear delimiters like XML tags, markdown blocks, or consistent headings.
- For each example, explicitly show both the process (reasoning, planning) and the outcome.
- Include examples of both successful operations and appropriate error handling.
- If possible, annotate examples with brief explanations of why specific approaches were taken.
- Ensure examples reflect the exact output format requirements.

## Handling Ambiguity and Uncertainty

**Ambiguity Resolution Strategies**
- Establish clear protocols for handling incomplete or ambiguous information.
- Define a hierarchy of information sources to consult when clarification is needed.
- Include explicit instructions for requesting additional information when necessary.
- Specify how to present multiple interpretations when a definitive answer isn't possible.
- Mandate expression of confidence levels for conclusions based on uncertain data.

**Conflict Resolution**
- Define a clear hierarchy of priorities for resolving conflicting requirements.
- Establish explicit rules for handling contradictory information from different sources.
- Include a protocol for identifying and surfacing contradictions rather than making assumptions.
- Specify when to defer to specific authorities (documentation, security policies) in case of conflicts.
- Provide a framework for transparently documenting resolution decisions when conflicts are encountered.

## Language Model Optimization

**Structured Tool Invocation is Mandatory**
- **Reiterate:** *All* actions, queries, commands, memory operations, delegations, and final result reporting **MUST** be performed via **structured tool calls** using the correct tool name variable (e.g., `{{.ToolName}}`).
- **Explicitly state:** Plain text descriptions or simulations of actions (e.g., writing "Running command `nmap -sV target.com`") **will not be executed** by the system.
- Use consistent template variables for tool names (see list below).
- Ensure prompts clearly specify expected parameters for critical tool calls.

**Completion Requirements Section**
- Always end prompts with a clearly marked section (e.g., `## COMPLETION REQUIREMENTS`) containing a **numbered list** of final instructions.
- Include a reminder about language: Respond/report in the user's/manager's preferred language (`{{.Lang}}`).
- Specify the required **final output format** and the **mandatory final tool** to use for delivery (e.g., `MUST use "{{.HackResultToolName}}" to deliver the final report`).
- **Crucially, place the `{{.ToolPlaceholder}}` variable at the very end of the prompt.** This allows the system backend to correctly inject tool definitions for the LLM.

### LLM Instruction Following Characteristics

**Modern LLM Instruction Following**
- Understand that newer LLMs (like those used in PentAGI) follow instructions **more literally and precisely** than previous generations. Make instructions explicit and unambiguous, avoiding indirect or implied guidance.
- Use **directive language** rather than suggestions: "DO X" instead of "You might want to do X" when the action is truly required.
- For critical behaviors, use **clear, unequivocal instructions** rather than lengthy explanations. A single direct statement is often more effective than paragraphs of background.
- When creating prompts, remember that if agent behavior deviates from expectations, a single clear corrective instruction is usually sufficient to guide it back on track.

**Literal Adherence vs. Intent Inference**
- Design prompts with the understanding that PentAGI agents will **follow the letter of instructions** rather than attempting to infer unstated intent.
- Make all critical behaviors explicit rather than relying on the agent to infer them from context or examples.
- If you need the agent to reason through problems rather than following a rigid process, explicitly instruct it to "think step-by-step" or "consider alternatives before deciding."

### Prompt Template Variables

**Essential Context Variables**
- Ensure prompts utilize essential context variables provided by the PentAGI backend:
    - `{{.ExecutionContext}}`: **Critical.** Provides structured details (IDs, status, titles, descriptions) about the current `Flow`, `Task`, and `SubTask`. Essential for scope and objective understanding.
    - `{{.Lang}}`: Specifies the preferred language for agent responses and reports.
    - `{{.CurrentTime}}`: Provides the execution timestamp for context.
    - `{{.DockerImage}}`: Name of the Docker image the agent operates within.
    - `{{.Cwd}}`: Default working directory inside the Docker container.
    - `{{.ContainerPorts}}`: Available/mapped ports within the container environment.

**Standardized Tool Name Variables**
- Use the consistent naming pattern for all tool invocation variables:
    - *Specialist Invocation:*
        - `{{.SearchToolName}}`
        - `{{.PentesterToolName}}`
        - `{{.CoderToolName}}`
        - `{{.AdviceToolName}}`
        - `{{.MemoristToolName}}`
        - `{{.MaintenanceToolName}}`
    - *Memory Operations:*
        - `{{.SearchGuideToolName}}` (Retrieve Guide)
        - `{{.StoreGuideToolName}}` (Store Guide)
        - `{{.SearchAnswerToolName}}` (Retrieve Answer/General)
        - `{{.StoreAnswerToolName}}` (Store Answer/General)
        - `{{.SearchCodeToolName}}` (*Likely needed*) (Retrieve Code Snippet)
        - `{{.StoreCodeToolName}}` (*Likely needed*) (Store Code Snippet)
    - *Result Delivery:*
        - `{{.HackResultToolName}}` (Pentester Final Report)
        - `{{.SearchResultToolName}}` (Searcher Final Report)
        - `{{.FinalyToolName}}` (Orchestrator Subtask Completion Report)
    - *System & Environment Tools:*
        - `{{.SummarizationToolName}}` (**System Use Only** - Marker for historical summaries)
        - `{{.TerminalToolName}}` (*Assumed name for terminal function*)
        - `{{.FileToolName}}` (*Assumed name for file operations function*)
        - `{{.BrowserToolName}}` (*Assumed name for browser/scraping function*)
        - *Ensure this list is kept synchronized with the actual tool names defined and passed by the backend.*

## Prompt Patterns and Anti-Patterns

**Effective Patterns**
- **Progressive Disclosure**: Introduce concepts in layers of increasing complexity.
- **Explicit Ordering**: Number steps or use clear sequence markers for sequential operations.
- **Task Decomposition**: Break complex tasks into clearly defined subtasks with their own guidelines.
- **Parameter Validation**: Include instructions for validating inputs before proceeding with operations.
- **Fallback Chains**: Define explicit alternatives when primary approaches fail.

**Common Anti-Patterns**
- **Overspecification**: Providing too many constraints that paralyze decision-making.
- **Conflicting Priorities**: Giving contradictory guidance without clear hierarchy.
- **Vague Success Criteria**: Failing to define when a task is considered complete.
- **Implicit Assumptions**: Relying on unstated knowledge or context.
- **Tool Ambiguity**: Unclear guidance on which tools to use for specific situations.

## Iterative Prompt Improvement

**Systematic Diagnosis**
- When prompts underperform, systematically isolate the issue: is it in task definition, reasoning guidance, tool usage, or output formatting?
- Document specific patterns of failure to address in revisions.
- Use controlled testing with identical inputs to validate improvements.
- Maintain version history with clear annotations about changes and their effects.
- Focus on targeted, minimal changes rather than wholesale rewrites when refining.

**Improvement Metrics**
- Define objective success criteria for prompt performance before making changes.
- Measure improvements across specific dimensions: accuracy, completeness, efficiency, robustness.
- Test prompts against edge cases and unusual inputs to ensure generalizability.
- Compare performance across different LLM providers to ensure consistency.
- Document both successful and unsuccessful prompt modifications to build institutional knowledge.

## Multimodal Integration

**Text-Visual Integration**
- When referencing visual elements, use precise descriptive language and spatial relationships.
- Define protocols for describing and referencing images, diagrams, or visualizations.
- For security-relevant visual information, instruct agents to extract and document specific details systematically.
- Establish clear formats for describing visual evidence in reports and documentation.
- Include guidance on when to request visual confirmation versus relying on textual descriptions.

## Agent-Specific Guidelines Summary

### Primary Agent (Orchestrator)
- **Focus**: Task decomposition, delegation orchestration, context management across subtasks, final subtask result aggregation.
- **Key Sections**: `TEAM CAPABILITIES`, `OPERATIONAL PROTOCOLS` (esp. Task Analysis, Boundaries, Delegation Efficiency), `DELEGATION PROTOCOL`, `SUMMARIZATION AWARENESS PROTOCOL`, `COMPLETION REQUIREMENTS` (using `{{.FinalyToolName}}`).
- **Critical Instructions**: Gather context *before* delegating, strictly enforce current subtask scope, provide *full* context upon delegation, manage execution attempts/failures, report subtask completion status and comprehensive results using `{{.FinalyToolName}}`.

### Pentester Agent
- **Focus**: Hands-on security testing, execution of tools (`nmap`, `sqlmap`, etc.), vulnerability exploitation, evidence collection and documentation.
- **Key Sections**: `KNOWLEDGE MANAGEMENT` (Memory Protocol), `OPERATIONAL ENVIRONMENT` (Container Constraints), `COMMAND EXECUTION RULES` (Terminal Protocol), `PENETRATION TESTING TOOLS` (list available), `TEAM COLLABORATION`, `DELEGATION PROTOCOL`, `SUMMARIZATION AWARENESS PROTOCOL`, `COMPLETION REQUIREMENTS` (using `{{.HackResultToolName}}`).
- **Critical Instructions**: Check memory first, strictly adhere to terminal rules & container constraints, use only listed available tools, delegate appropriately (e.g., exploit development to Coder), provide detailed, evidence-backed exploitation reports using `{{.HackResultToolName}}`.

### Searcher Agent
- **Focus**: Highly efficient information retrieval (internal memory & external sources), source evaluation and prioritization, synthesis of findings.
- **Key Sections**: `CORE CAPABILITIES` (Action Economy, Search Optimization), `SEARCH TOOL DEPLOYMENT MATRIX`, `OPERATIONAL PROTOCOLS` (Search Efficiency, Query Engineering), `SUMMARIZATION AWARENESS PROTOCOL`, `SEARCH RESULT DELIVERY` (using `{{.SearchResultToolName}}`).
- **Critical Instructions**: **Always prioritize memory search** (`{{.SearchAnswerToolName}}`), strictly limit the number of search actions, use the right tool for the query complexity (Matrix), **stop searching once sufficient information is gathered**, deliver concise yet comprehensive synthesized results via `{{.SearchResultToolName}}`.

*(Guidelines for Developer, Adviser, Memorist, Installer agents should be developed following this structure, focusing on their unique roles, tools, and interactions based on their specific implementations and prompt templates).*

## Prompt Maintenance and Evolution

### Version Control and Documentation
- Store all prompt templates consistently within the `backend/pkg/templates/prompts/` directory.
- Use a clear and consistent naming pattern: `<agent_role>[_optional_specifier].tmpl`.
- Include version information or brief changelog comments within the templates themselves or in associated documentation.
- Document the purpose, expected template variables (`{{.Variable}}`), and the general input/output behavior for each prompt template. Ensure this documentation stays synchronized with the backend code that populates the variables.

### Testing and Refinement
- Utilize the `ctester` utility (`backend/cmd/ctester/`) for validating LLM provider compatibility and basic prompt adherence (e.g., JSON formatting, function calling capabilities) for different agent types. Reference `development-workflow.mdc` / `README.md`.
- Employ the `ftester` utility (`backend/cmd/ftester/`) for **in-depth testing** of specific agent functions and prompt behaviors within realistic contexts (Flow/Task/SubTask). This is crucial for debugging complex interactions and prompt logic.
- Actively analyze agent performance, errors, and interaction traces using observability tools like **Langfuse**. Identify patterns where prompts are misunderstood, lead to inefficient actions, or violate protocols.
- Refine prompts iteratively based on `ctester`, `ftester`, and Langfuse analysis. Test changes thoroughly before deployment.
- Verify prompt changes across different supported LLM providers to ensure consistent behavior.
- Regularly validate that XML structures are well-formed and consistently applied across prompts.

### Prompt Evolution Workflow
- Document successful vs. unsuccessful prompt patterns to build institutional knowledge
- Identify areas where agents commonly misunderstand instructions or violate protocols
- Focus refinement efforts on critical sections with highest impact on performance
- Test prompt changes systematically with controlled variables
- When adding new agent types or specializations, adapt existing templates rather than creating entirely new structures

### Prompt Debugging Guide
- When agents act incorrectly, first check: Are instructions contradictory? Are priorities clear? Is context sufficient?
- For reasoning failures, examine if the problem has been properly decomposed and if verification steps exist.
- For tool usage errors, verify tool descriptions and examples are clear and parameters well-defined.
- When memory usage is suboptimal, check memory protocol clarity and retrieval/storage guidance.
- Document common failure modes to address in future prompt revisions.

## Implementation Examples

*(Refer to the actual, up-to-date files in `backend/pkg/templates/prompts/` such as `primary_agent.tmpl`, `pentester.tmpl`, and `searcher.tmpl` for concrete implementation patterns that follow these guidelines.)*
