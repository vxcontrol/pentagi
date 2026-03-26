# Analytics API

GraphQL API for system analytics and statistics monitoring. This document covers three main use cases for developers building dashboards and analytics tools.

## Important Data Notes

### Duration Storage and Calculation

**Database columns (as of migration 20260129_120000):**
- `toolcalls.duration_seconds`: Pre-calculated duration for completed toolcalls
  - Automatically set during migration: `EXTRACT(EPOCH FROM (updated_at - created_at))`
  - Set to 0.0 for incomplete toolcalls (`received`, `running`)
  - Column is `NOT NULL DEFAULT 0.0`
  - **Incremental updates**: Backend calculates delta time and passes it to SQL
  - Formula: `duration_seconds = duration_seconds + duration_delta` (parameter from Go code)
  - Updated by SQL queries: `UpdateToolcallStatus`, `UpdateToolcallFinishedResult`, `UpdateToolcallFailedResult`
  
- `msgchains.duration_seconds`: Pre-calculated duration for message chains
  - Automatically set during migration: `EXTRACT(EPOCH FROM (updated_at - created_at))`
  - Set to 0.0 for incomplete msgchains
  - Column is `NOT NULL DEFAULT 0.0`
  - **Incremental updates**: Backend calculates delta time and passes it to SQL
  - Formula: `duration_seconds = duration_seconds + duration_delta` (parameter from Go code)
  - Updated by SQL queries: `UpdateMsgChain`, `UpdateMsgChainUsage`

**Benefits of pre-calculated durations:**
- Faster query performance (no real-time `EXTRACT(EPOCH FROM ...)` calculations)
- Consistent duration values (accumulated over multiple updates)
- Simpler SQL queries for analytics (just `tc.duration_seconds` instead of complex expressions)
- Reduced database load for dashboard queries
- Incremental tracking: accurately captures time even when records are updated multiple times
- All analytics SQL queries updated to use `duration_seconds` column directly

**Incremental update logic:**
- Backend code tracks execution time and calculates delta in seconds
- Delta is passed as parameter to SQL update queries
- SQL adds delta to existing `duration_seconds` value
- Prevents overwrites and ensures accurate cumulative duration
- Works correctly for long-running operations with multiple status updates
- Backend has full control over time measurement (can use high-precision timers)

### Execution Time Metrics (`totalDurationSeconds`)

**What the data shows:**
- **Subtask duration**: Linear wall-clock time from subtask start to finish (created_at → updated_at)
  - Excludes subtasks in `created` or `waiting` status (not yet started)
  - Includes only `running`, `finished`, `failed` subtasks
  - **NOT** the sum of toolcalls (avoids double-counting nested agent calls)
  
- **Task duration**: Total execution time including:
  - Generator Agent execution (runs before subtasks)
  - All subtasks execution (with overlap compensation for batch-created subtasks)
  - Refiner Agent execution(s) (runs between subtasks)
  
- **Flow duration**: Complete flow execution time including:
  - All tasks duration (generator + subtasks + refiner)
  - Flow-level toolcalls (e.g., Assistant toolcalls without task/subtask binding)
  - **EXCLUDES** Assistant msgchain lifetime (only active toolcall time counted)

**Why not sum of toolcalls:**
- Primary Agent calls Coder Agent → Coder calls Terminal
- Summing toolcalls would count Terminal execution twice (in Coder time AND separately)
- Linear time gives accurate wall-clock duration

**Overlap compensation:**
- Generator creates multiple subtasks simultaneously (same created_at)
- Subtasks execute sequentially, not in parallel
- System compensates for this overlap to show real execution time

### Toolcalls Count Metrics (`totalCount`, `totalToolcallsCount`)

**What the data shows:**
- Only **completed** toolcalls (status = `finished` or `failed`)
- Excludes `received` (not started) and `running` (in progress) toolcalls
- Represents actual completed operations, not attempted or pending

**Use this for:**
- Counting successful/failed operations
- Calculating success rate: `finished_count / (finished_count + failed_count)`
- Understanding actual work performed

### Agent Type Breakdown

**Important distinction:**
- `primary_agent`: Main orchestrator for subtasks
- `generator`: Creates subtask plans (runs once per task, before subtasks)
- `refiner`: Updates subtask plans (runs between subtasks, can run multiple times)
- `reporter`: Generates final task reports
- `assistant`: Interactive mode (can have long idle periods)
- Specialist agents: `coder`, `pentester`, `installer`, `searcher`, `memorist`, `adviser`

**Usage interpretation:**
- High `generator` usage = complex task decomposition
- High `refiner` usage = adaptive planning (many plan adjustments)
- High specialist usage = delegated work (Primary Agent using team members)

### Agent vs Non-Agent Tools (`isAgent` field)

The `isAgent` field in `FunctionToolcallsStats` categorizes tools into two types:

**Agent Tools (`isAgent: true`):**
- Delegation to AI agents (e.g., `coder`, `pentester`, `searcher`)
- Store results from agent execution (e.g., `coder_result`, `hack_result`)
- Execution time represents delegation overhead + agent decision time
- Does NOT include nested tool calls made by the agent
- Examples: `coder`, `pentester`, `installer`, `searcher`, `memorist`, `adviser`

**Non-Agent Tools (`isAgent: false`):**
- Direct tool execution (e.g., `terminal`, `browser`, `file`)
- Search engines (e.g., `google`, `duckduckgo`, `tavily`, `sploitus`, `searxng`)
- Vector database operations (e.g., `search_in_memory`, `store_guide`)
- Environment operations (e.g., `terminal`, `file`)
- Execution time represents actual operation duration
- Examples: `terminal`, `browser`, `file`, `google`, `search_in_memory`

**Use this distinction for:**
- Analyzing delegation overhead vs direct execution time
- Identifying which agents are most frequently used
- Optimizing agent selection strategies
- Understanding the balance between AI-driven and direct operations
- Dashboard visualizations: color-code or separate agent vs non-agent tools

## Common Fragments

Define these fragments once and reuse across queries:

```graphql
fragment UsageStats on UsageStats {
  totalUsageIn
  totalUsageOut
  totalUsageCacheIn
  totalUsageCacheOut
  totalUsageCostIn
  totalUsageCostOut
}

fragment ToolcallsStats on ToolcallsStats {
  totalCount
  totalDurationSeconds
}

fragment FlowsStats on FlowsStats {
  totalFlowsCount
  totalTasksCount
  totalSubtasksCount
  totalAssistantsCount
}

fragment FlowStats on FlowStats {
  totalTasksCount
  totalSubtasksCount
  totalAssistantsCount
}

fragment FunctionToolcallsStats on FunctionToolcallsStats {
  functionName
  isAgent
  totalCount
  totalDurationSeconds
  avgDurationSeconds
}

fragment SubtaskExecutionStats on SubtaskExecutionStats {
  subtaskId
  subtaskTitle
  totalDurationSeconds
  totalToolcallsCount
}

fragment TaskExecutionStats on TaskExecutionStats {
  taskId
  taskTitle
  totalDurationSeconds
  totalToolcallsCount
  subtasks {
    ...SubtaskExecutionStats
  }
}

fragment FlowExecutionStats on FlowExecutionStats {
  flowId
  flowTitle
  totalDurationSeconds
  totalToolcallsCount
  totalAssistantsCount
  tasks {
    ...TaskExecutionStats
  }
}
```

---

## Use Case 1: Time-Filtered Dashboard

**Purpose:** Display system activity over time periods (week/month/quarter) with day-by-day breakdowns and execution metrics.

### Query

```graphql
query TimeDashboard($period: UsageStatsPeriod!) {
  # LLM token usage over time
  usageStatsByPeriod(period: $period) {
    date
    stats { ...UsageStats }
  }
  
  # Toolcalls activity over time
  toolcallsStatsByPeriod(period: $period) {
    date
    stats { ...ToolcallsStats }
  }
  
  # Flows/tasks/subtasks created over time
  flowsStatsByPeriod(period: $period) {
    date
    stats { ...FlowsStats }
  }
  
  # Flow execution times with full hierarchy
  flowsExecutionStatsByPeriod(period: $period) {
    ...FlowExecutionStats
  }
}
```

**Variables:**
```json
{ "period": "week" }  // or "month", "quarter"
```

### Data Interpretation

**Daily Usage Trends:**
- `usageStatsByPeriod`: Track token consumption patterns (input/output, cache hits, costs)
- Chart: Line graph showing daily token usage and costs
- Insights: Identify peak usage days, cost optimization opportunities

**Toolcalls Performance:**
- `toolcallsStatsByPeriod`: Monitor tool execution frequency and duration
- **Count:** Only completed operations (finished/failed), excludes pending/running
- **Duration:** Sum of individual toolcall execution times (created_at → updated_at)
- Chart: Dual-axis chart (count bars + duration line)
- Insights: Detect performance degradation, identify bottlenecks
- **Note:** Duration here IS sum of toolcalls (unlike execution stats which use wall-clock)

**Flow Activity:**
- `flowsStatsByPeriod`: Track system load (flows/tasks/subtasks/assistants created per day)
- Chart: Stacked bar chart showing hierarchy depth
- Insights: Understand workload distribution, capacity planning, assistant usage patterns

**Execution Breakdown:**
- `flowsExecutionStatsByPeriod`: Hierarchical view of execution times
- Chart: Treemap or sunburst showing time distribution across flows → tasks → subtasks
- **Important:** Duration is **wall-clock time**, not sum of toolcalls
  - Subtask: Linear time from start to finish (excludes created/waiting subtasks)
  - Task: Subtasks + Generator + Refiner agents
  - Flow: Tasks + Flow-level toolcalls (Assistant active time, NOT lifetime)
- **Count:** Only completed toolcalls (finished/failed status)
- Insights: Identify slow flows/tasks, optimize critical paths
- **Note:** Batch-created subtasks have overlap compensation applied

**Cross-Correlations:**
- Compare token usage vs execution time (efficiency metric)
- Correlate toolcall count with flow complexity
- Cost per flow: `totalUsageCost / flowsCount`
- Average toolcalls per task: `totalToolcallsCount / totalTasksCount`

---

## Use Case 2: Overall System Statistics

**Purpose:** Get comprehensive system-wide metrics without time filters.

### Query

```graphql
query SystemOverview {
  # Total LLM usage
  usageStatsTotal { ...UsageStats }
  
  # Usage breakdown by provider
  usageStatsByProvider {
    provider
    stats { ...UsageStats }
  }
  
  # Usage breakdown by model
  usageStatsByModel {
    model
    provider
    stats { ...UsageStats }
  }
  
  # Usage breakdown by agent type
  usageStatsByAgentType {
    agentType
    stats { ...UsageStats }
  }
  
  # Total toolcalls stats
  toolcallsStatsTotal { ...ToolcallsStats }
  
  # Toolcalls breakdown by function
  toolcallsStatsByFunction {
    ...FunctionToolcallsStats
  }
  
  # Total flows/tasks/subtasks
  flowsStatsTotal { ...FlowsStats }
}
```

### Data Interpretation

**Token Economics:**
- `usageStatsTotal`: Overall system cost and token consumption
- Metrics: Total spend, cache efficiency (`cacheIn / (cacheIn + usageIn)`)
- Dashboard KPI: Display as headline metrics

**Provider Distribution:**
- `usageStatsByProvider`: Cost and usage per LLM provider
- Chart: Pie chart showing provider share by cost
- Insights: Identify most/least cost-effective providers

**Model Efficiency:**
- `usageStatsByModel`: Granular per-model breakdown
- Chart: Table sorted by cost with usage metrics
- Metrics:
  - Cost per token: `(costIn + costOut) / (usageIn + usageOut)`
  - Cache hit rate per model
- Insights: Choose optimal models for different tasks

**Agent Performance:**
- `usageStatsByAgentType`: Resource consumption by agent role
- Chart: Horizontal bar chart showing usage by agent
- Insights: Understand which agents consume most resources

**Tool Usage Patterns:**
- `toolcallsStatsByFunction`: Top tool functions by usage and duration
- **Agent Classification:** The `isAgent` field indicates if the function is an agent delegation tool
  - Agent tools (e.g., `coder`, `pentester`) show their own execution time
  - Does NOT include time of nested calls they make
  - Example: `coder` toolcall = time for Coder Agent to decide + delegate, not terminal commands
  - Non-agent tools (e.g., `terminal`, `browser`) show direct execution time
- Chart: Table with sortable columns (count, duration, average, agent type)
- Metrics:
  - Slowest tools: Sort by `avgDurationSeconds`
  - Most used tools: Sort by `totalCount` (only completed)
  - Time sinks: Sort by `totalDurationSeconds`
  - Agent vs non-agent breakdown: Filter by `isAgent`
- Insights: Optimize frequently-used slow tools, identify unused tools, distinguish delegation overhead from direct execution

**System Scale:**
- `flowsStatsTotal`: Total entities in system
- Metrics:
  - Tasks per flow: `totalTasksCount / totalFlowsCount`
  - Subtasks per task: `totalSubtasksCount / totalTasksCount`
  - Assistants per flow: `totalAssistantsCount / totalFlowsCount`
- Insights: Understand average flow complexity and assistant usage

---

## Use Case 3: Flow-Specific Dashboard

**Purpose:** Deep dive into a specific flow's metrics.

### Query

```graphql
query FlowAnalytics($flowId: ID!) {
  # Basic flow info
  flow(flowId: $flowId) {
    id
    title
    status
    createdAt
    updatedAt
  }
  
  # LLM usage for this flow
  usageStatsByFlow(flowId: $flowId) { ...UsageStats }
  
  # Agent usage breakdown
  usageStatsByAgentTypeForFlow(flowId: $flowId) {
    agentType
    stats { ...UsageStats }
  }
  
  # Toolcalls stats
  toolcallsStatsByFlow(flowId: $flowId) { ...ToolcallsStats }
  
  # Tool function breakdown
  toolcallsStatsByFunctionForFlow(flowId: $flowId) {
    ...FunctionToolcallsStats
  }
  
  # Example: Separate agent and non-agent tools
  # Filter client-side: stats.filter(s => s.isAgent) for agent tools
  # Filter client-side: stats.filter(s => !s.isAgent) for direct execution tools
  
  # Flow structure stats
  flowStatsByFlow(flowId: $flowId) { ...FlowStats }
}
```

**Variables:**
```json
{ "flowId": "123" }
```

### Data Interpretation

**Flow Performance Summary:**
- `usageStatsByFlow`: Total LLM costs for this flow (all msgchains)
- `toolcallsStatsByFlow`: Execution metrics (duration, toolcall count)
  - **Count**: Only completed toolcalls (finished/failed)
  - **Duration**: Sum of individual toolcall times
- KPIs:
  - Cost per toolcall: `totalUsageCost / totalToolcallsCount`
  - Average toolcall duration: `totalDurationSeconds / totalCount`
  - Cost efficiency: tokens per second
- **Note:** For actual flow wall-clock time, use `flowsExecutionStatsByPeriod`

**Agent Activity:**
- `usageStatsByAgentTypeForFlow`: Which agents were most active
- Chart: Donut chart showing token distribution by agent
- Insights: Understand which agents drive flow execution

**Tool Usage Analysis:**
- `toolcallsStatsByFunctionForFlow`: Detailed breakdown per tool
- Chart: Bubble chart (x=count, y=avgDuration, size=totalDuration, color=isAgent)
- Metrics:
  - Identify bottleneck tools (high avgDuration)
  - Find frequently-used tools (high totalCount)
  - Calculate tool efficiency scores
  - Separate agent delegation overhead from direct execution time using `isAgent` field

**Flow Complexity:**
- `flowStatsByFlow`: Structural metrics
- Metrics:
  - Subtasks per task: `totalSubtasksCount / totalTasksCount`
  - Toolcalls per task: `toolcallsCount / tasksCount`
  - Assistants count: `totalAssistantsCount`
- Insights: Compare against average complexity to identify outliers, track assistant usage per flow

**Cross-Flow Comparisons:**
Fetch multiple flows and compare:
- Cost efficiency (cost per task)
- Execution speed (duration per task)
- Tool utilization patterns
- Agent composition differences

---

## Understanding Metric Differences

### Execution Stats vs Toolcalls Stats

These two metric types measure different aspects of system performance:

**Execution Stats (`flowsExecutionStatsByPeriod`):**
- **Purpose:** Measure real wall-clock time for flows/tasks/subtasks
- **Duration calculation:** Linear time (start → end timestamp)
- **What it shows:** How long a flow/task/subtask actually ran
- **Use for:** Performance analysis, SLA monitoring, user-facing progress
- **Example:** Subtask ran for 100 seconds (even if it made 10 toolcalls)

**Toolcalls Stats (`toolcallsStatsByPeriod`, `toolcallsStatsByFunction`):**
- **Purpose:** Measure individual tool execution metrics
- **Duration calculation:** Sum of toolcall durations (each toolcall's created_at → updated_at)
- **What it shows:** Aggregate time spent in specific tools
- **Use for:** Tool optimization, identifying slow functions, resource attribution
- **Example:** 50 terminal toolcalls totaling 300 seconds

**Key difference:**
```
Flow execution time = 100 seconds (wall-clock)
Toolcalls total time = 150 seconds (sum of all toolcalls)

Why different? 
- Flow time is LINEAR (real time elapsed)
- Toolcalls time INCLUDES OVERLAPS (nested agent calls counted in parent time)
```

**When to use which:**
- User wants to know "how long did my pentest take?" → Use **Execution Stats**
- Developer wants to optimize slow tools → Use **Toolcalls Stats**
- Manager wants to see system utilization → Use **Toolcalls Stats**
- SLA monitoring → Use **Execution Stats**

### Subtask Status and Inclusion

**Included in metrics (counted):**
- `running`: Currently executing (duration = created_at → now)
- `finished`: Completed successfully (duration = created_at → updated_at)
- `failed`: Terminated with error (duration = created_at → updated_at)

**Excluded from metrics (NOT counted):**
- `created`: Generated but not started yet (duration = 0)
- `waiting`: Paused for user input (duration = 0)

**Why this matters:**
- Generator creates 10 subtasks at once, only 1 starts executing
- You'll see 1 subtask in stats, not 10
- As more subtasks execute, they appear in metrics
- Final stats include only executed subtasks

### Assistant Time Accounting

**Assistant msgchains:**
- Can exist for days/weeks (created once, used intermittently)
- Their **lifetime is NOT counted** in flow duration
- Only their **active toolcalls** are counted

**Example:**
```
Assistant created: Monday 9 AM
User asks question: Monday 10 AM (toolcall 1: 5 seconds)
User asks question: Tuesday 3 PM (toolcall 2: 3 seconds)

Flow duration contribution: 8 seconds (5 + 3)
NOT: 30+ hours (Monday to Tuesday)
```

### Generator and Refiner Inclusion

**Task execution includes:**
1. **Generator Agent** (runs once at task start):
   - Creates initial subtask plan
   - Has msgchain with task_id, NO subtask_id
   - Time is added to task duration
   
2. **Subtasks** (execute sequentially):
   - Each has Primary Agent msgchain
   - Individual durations with overlap compensation
   
3. **Refiner Agent** (runs between subtasks):
   - Updates subtask plan based on results
   - Can run multiple times per task
   - Each run has msgchain with task_id, NO subtask_id
   - Total refiner time added to task duration

**Example task timeline:**
```
Generator (5s) → Subtask 1 (10s) → Refiner (3s) → Subtask 2 (15s) → Refiner (2s) → Subtask 3 (8s)

Task duration = 5 + 10 + 3 + 15 + 2 + 8 = 43 seconds
```

---

## Advanced Analytics Patterns

### 1. Cost Optimization Dashboard

Combine queries to identify cost reduction opportunities:

```graphql
query CostOptimization {
  usageStatsByModel {
    model
    provider
    stats { ...UsageStats }
  }
  toolcallsStatsByFunction {
    functionName
    isAgent
    totalCount
    avgDurationSeconds
  }
}
```

**Analysis:**
- Expensive models with high usage → candidates for cheaper alternatives
- Slow tools called frequently → optimization targets
- Calculate ROI per model: `performance_gain / cost_increase`

### 2. Performance Monitoring

Track system responsiveness:

```graphql
query PerformanceMetrics($period: UsageStatsPeriod!) {
  toolcallsStatsByPeriod(period: $period) {
    date
    stats { ...ToolcallsStats }
  }
  flowsExecutionStatsByPeriod(period: $period) {
    flowTitle
    totalDurationSeconds
    totalToolcallsCount
  }
}
```

**Metrics:**
- Average execution time per flow: `totalDurationSeconds / flowsCount`
- Toolcalls per day trend (detect performance degradation)
- P95/P99 flow durations (for SLA monitoring)

**Important for performance analysis:**
- Use `flowsExecutionStatsByPeriod` for wall-clock time (what users experience)
- Compare with `toolcallsStatsByPeriod` to detect overhead (high ratio = optimization needed)
- Ratio > 2.0 suggests significant nested agent call overhead

### 3. Resource Attribution

Understand resource consumption patterns:

```graphql
query ResourceAttribution {
  usageStatsByAgentType {
    agentType
    stats { ...UsageStats }
  }
  toolcallsStatsByFunction {
    functionName
    isAgent
    totalDurationSeconds
  }
}
```

**Analysis:**
- Which agents consume most resources
- Tool time distribution (execution time budget)
- Cost attribution by capability (pentesting vs coding vs searching)

---

## Implementation Tips

**Caching Strategy:**
- Cache `*StatsTotal` queries (update every 5-10 minutes)
- Cache `*StatsByPeriod` per period (update hourly)
- Real-time for flow-specific queries

**Visualization Libraries:**
- Time series: Recharts, Chart.js (daily trends)
- Hierarchical: D3 treemap/sunburst (execution breakdown)
- Tables: TanStack Table with sorting/filtering (function stats)

**Performance Optimization:**
- Use query batching for multiple flows
- Implement pagination for large datasets
- Add loading states for slow queries (execution stats can be heavy)

**Data Refresh:**
- Overall stats: Manual refresh or 10min polling
- Time-filtered: Auto-refresh on period change
- Flow-specific: Subscribe to flow updates for real-time metrics

---

## Practical Examples

### Example 1: Understanding a Slow Flow

**Scenario:** Flow took 300 seconds but only has 50 toolcalls

**Investigation:**
```graphql
query InvestigateSlowFlow($flowId: ID!) {
  flowsExecutionStatsByPeriod(period: week) {
    flowId
    flowTitle
    totalDurationSeconds
    totalToolcallsCount
    tasks {
      taskTitle
      totalDurationSeconds
      totalToolcallsCount
      subtasks {
        subtaskTitle
        totalDurationSeconds
        totalToolcallsCount
      }
    }
  }
}
```

**Analysis:**
1. Check task breakdown: Which task took longest?
2. Check subtask breakdown: Which subtasks in slow task took longest?
3. Check toolcall count: High duration + low count = slow individual operations
4. Check toolcalls by function: Which tools are slow?

**Common causes:**
- Long-running terminal commands (compilation, scanning)
- Slow search engine responses (tavily, perplexity)
- Large file operations
- Network latency for browser tool

### Example 2: Cost Attribution

**Scenario:** Need to understand cost per capability

**Query:**
```graphql
query CostAttribution {
  usageStatsByAgentType {
    agentType
    stats {
      totalUsageCostIn
      totalUsageCostOut
    }
  }
  
  toolcallsStatsByFunction {
    functionName
    totalCount
    avgDurationSeconds
  }
}
```

**Interpretation:**
- `primary_agent`: Orchestration overhead
- `generator`: Planning cost (usually low, runs once per task)
- `refiner`: Replanning cost (high value = many adjustments)
- `pentester`: Security testing operations
- `coder`: Development work
- `searcher`: Research and information gathering

**Cost optimization:**
- High generator cost → simplify task descriptions
- High refiner cost → improve initial planning (fewer adjustments needed)
- High searcher cost → use memory tools more (cheaper than web search)

### Example 3: Detecting Inefficient Flows

**Red flags:**
```
Flow A: 100 toolcalls, 500 seconds → 5s per toolcall (GOOD)
Flow B: 100 toolcalls, 5000 seconds → 50s per toolcall (INVESTIGATE)
```

**Check:**
```graphql
query DetectInefficiency {
  toolcallsStatsByFunction {
    functionName
    isAgent
    totalCount
    avgDurationSeconds
  }
}
```

**Common issues:**
- High `terminal` avg → long commands, consider timeout tuning
- High `browser` avg → slow websites, consider caching
- High `tavily`/`perplexity` avg → deep research, optimize queries
- High agent tools (`coder`, `pentester`) avg → complex delegated work

### Example 4: Understanding Task Complexity

**Scenario:** Why do some tasks take so long?

**Metrics to check:**
```
Task complexity indicators:
- High subtask count → decomposition into many steps
- High refiner calls → adaptive planning (many plan changes)
- High generator time → complex initial planning
- Low subtask count + high duration → individual subtasks are slow
```

**Query:**
```graphql
query TaskComplexity($flowId: ID!) {
  flowsExecutionStatsByPeriod(period: week) {
    flowId
    tasks {
      taskTitle
      totalDurationSeconds
      totalToolcallsCount
      subtasks {
        subtaskTitle
        totalDurationSeconds
      }
    }
  }
  
  usageStatsByAgentTypeForFlow(flowId: $flowId) {
    agentType
    stats { totalUsageIn totalUsageOut }
  }
}
```

**Analysis patterns:**
- Many subtasks + low generator usage → simple decomposition
- Many subtasks + high generator usage → complex planning
- High refiner usage → dynamic adaptation (plan changed during execution)
- Few subtasks + high duration → intensive work per subtask

---

## Data Quality Guarantees

### Accuracy

**Time measurements:**
- ✅ No double-counting of nested agent calls
- ✅ Overlap compensation for batch-created subtasks
- ✅ Excludes non-started subtasks (created/waiting)
- ✅ Includes all execution phases (generator, subtasks, refiner)

**Count measurements:**
- ✅ Only completed operations (finished/failed)
- ✅ Excludes pending (received) and in-progress (running)
- ✅ Consistent across all queries

**Cost measurements:**
- ✅ Aggregated from msgchains (source of truth for LLM calls)
- ✅ Includes cache hits and misses
- ✅ Separate input/output costs

### Known Limitations

**Current limitations:**
1. **Historical data:** Subtasks created before this update may have:
   - Missing primary_agent subtask_id (known bug, now fixed)
   - Use linear time fallback (still accurate)

2. **Running entities:** Duration calculated as `created_at → now`
   - Updates as entity continues execution
   - Final duration set when status changes to finished/failed

3. **Assistant lifetime:** Long-lived assistants
   - Only active toolcall time counted (correct behavior)
   - Msgchain lifetime NOT included in flow duration

**Edge cases handled:**
- ✅ Batch-created subtasks (overlap compensation)
- ✅ Missing primary_agent msgchain (fallback to linear time)
- ✅ Subtasks in waiting status (excluded from duration)
- ✅ Flow-level toolcalls without task binding (counted separately)
- ✅ Generator/Refiner without subtask binding (counted in task duration)

---

## Troubleshooting

### "My flow shows 0 duration but has toolcalls"

**Possible causes:**
1. All subtasks are in `created` or `waiting` status
2. Flow just started (no completed subtasks yet)

**Check:**
```graphql
query CheckFlowStatus($flowId: ID!) {
  flow(flowId: $flowId) {
    status
  }
  tasks(flowId: $flowId) {
    status
    subtasks {
      status
    }
  }
}
```

### "Task duration seems low compared to subtasks"

**This is normal if:**
- Subtasks were created in batch (overlap compensation applied)
- Example: 3 subtasks created at 10:00:00, finished at 10:00:10, 10:00:20, 10:00:30
- Naive sum: 10 + 20 + 30 = 60 seconds
- Actual time: 30 seconds (overlap compensated)

### "Toolcalls duration > execution duration"

**This is expected:**
- **Toolcalls duration:** Sum of all toolcall times (includes nested calls)
- **Execution duration:** Wall-clock time (linear)
- Nested agent calls cause toolcalls > execution
- Example: Primary Agent (100s) calls Coder (30s), toolcalls = 130s, execution = 100s

### "Count doesn't match my expectations"

**Remember:**
- Only **completed** toolcalls counted (finished/failed)
- Received (pending) and running (in-progress) excluded
- Check toolcall status distribution if counts seem low

---

## Best Practices

### Dashboard Design

**Real-time monitoring:**
- Use execution stats for user-facing progress
- Show flow/task/subtask hierarchy with durations
- Update as status changes (subscribe to updates)

**Historical analysis:**
- Use toolcalls stats for tool performance
- Use usage stats for cost tracking
- Group by period for trend analysis

**Cost optimization:**
- Compare cost per agent type
- Identify expensive models with low value
- Track cache hit rates for efficiency

### Query Optimization

**For large datasets:**
```graphql
# Don't fetch full hierarchy if not needed
query LightweightStats {
  toolcallsStatsTotal { totalCount totalDurationSeconds }
  usageStatsTotal { totalUsageCostIn totalUsageCostOut }
}

# Instead of:
query HeavyStats($period: UsageStatsPeriod!) {
  flowsExecutionStatsByPeriod(period: $period) {
    # Full hierarchy - expensive for many flows
    tasks { subtasks { ... } }
  }
}
```

**Batch requests:**
```graphql
query BatchedAnalytics {
  # Fetch multiple metrics in one request
  usageStatsTotal { ... }
  toolcallsStatsTotal { ... }
  flowsStatsTotal { ... }
}
```

---

## Data Refresh Strategy

### Real-time (WebSocket subscriptions)
- Flow status changes
- Task/subtask creation
- Toolcall completion
- **Use for:** Live flow monitoring

### Polling (every 1-5 minutes)
- Execution stats for running flows
- Toolcalls stats for active periods
- **Use for:** Dashboard auto-refresh

### Cached (refresh every 10-30 minutes)
- Historical period stats
- Total stats (system-wide)
- Provider/model breakdowns
- **Use for:** Reports and analytics

### On-demand (user action)
- Flow-specific deep dives
- Custom period queries
- Export operations
- **Use for:** Detailed investigation

---

## Migration Notes

### Duration Calculation Changes (20260129_120000)

**Previous behavior:**
- Durations calculated on-the-fly in SQL queries: `EXTRACT(EPOCH FROM (updated_at - created_at))`
- Slower query performance due to real-time calculations
- Required complex SQL expressions in every analytics query

**New behavior (improved performance):**
- Pre-calculated `duration_seconds` columns in `toolcalls` and `msgchains` tables
- Duration calculated once during migration for existing records
- Analytics queries use simple column references: `tc.duration_seconds`
- Significant performance improvement for dashboard queries (simpler execution plans)

**Migration steps:**
1. Add `duration_seconds DOUBLE PRECISION NULL` column
2. Calculate duration for existing records: `EXTRACT(EPOCH FROM (updated_at - created_at))`
3. Set remaining NULL values to 0.0
4. Alter column to `NOT NULL`
5. Set default value to 0.0 for future records

**For developers:**
- All SQL queries in `backend/sqlc/models/toolcalls.sql` updated to use `duration_seconds`
- Update queries accept `duration_delta` parameter from Go code
- Updated queries with new signatures:
  - `UpdateToolcallStatus(status, duration_delta, id)`: adds delta to duration when status changes
  - `UpdateToolcallFinishedResult(result, duration_delta, id)`: adds final delta when toolcall finishes
  - `UpdateToolcallFailedResult(result, duration_delta, id)`: adds final delta when toolcall fails
  - `UpdateMsgChain(chain, duration_delta, id)`: adds delta when chain is updated
  - `UpdateMsgChainUsage(usage_in, usage_out, usage_cache_in, usage_cache_out, usage_cost_in, usage_cost_out, duration_delta, id)`: adds delta when usage updated
- Backend code must calculate `duration_delta` in seconds before calling update
- Use `time.Since(startTime).Seconds()` or similar for accurate measurements
- SQL simply adds the delta: `duration_seconds = duration_seconds + $duration_delta`
