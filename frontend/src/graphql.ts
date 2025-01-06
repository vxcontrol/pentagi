import { devtoolsExchange } from "@urql/devtools";
import { Data, cacheExchange } from "@urql/exchange-graphcache";
import { gql } from "graphql-tag";
import { createClient as createWSClient, Client } from "graphql-ws";
import { createClient, fetchExchange, subscriptionExchange } from "urql";

import schema from "../generated/graphql.schema.json";

const baseURL = "/api/v1";

export const cache = cacheExchange({
  schema: schema,
  updates: {
    Mutation: {
      createFlow: (result, _args, cache) => {
        const flow = result.createFlow as Data;
        const flows = cache.resolve("Query", "flows") as string[];

        if (flows) {
          const existingFlow = flows.find(f => cache.resolve(f as string, "id") === flow.id);
          if (!existingFlow) {
            cache.link("Query", "flows", [flow, ...flows]);
          }
        } else {
          cache.link("Query", "flows", [flow]);
        }
      },
    },
    Subscription: {
      flowCreated: (result, _args, cache) => {
        const flow = result.flowCreated as Data;
        const flows = cache.resolve("Query", "flows") as string[];

        if (flows) {
          const existingFlow = flows.find(f => cache.resolve(f as string, "id") === flow.id);
          if (!existingFlow) {
            cache.link("Query", "flows", [flow, ...flows]);
          }
        } else {
          cache.link("Query", "flows", [flow]);
        }
      },

      flowDeleted: (result, _args, cache) => {
        const deletedFlow = result.flowDeleted as Data;
        const flows = cache.resolve("Query", "flows") as string[];

        if (flows) {
          cache.link(
            "Query",
            "flows",
            flows.filter((f) => cache.resolve(f as string, "id") !== deletedFlow.id)
          );
        }
      },

      flowUpdated: (result, _args, cache) => {
        const flow = result.flowUpdated as Data;

        cache.writeFragment(
          gql`
            fragment FlowUpdate on Flow {
              id
              status
              terminals
              updatedAt
            }
          `,
          flow
        );
      },

      taskCreated: (result, _args, cache) => {
        const task = result.taskCreated as Data;
        const flowId = task.flowId as string;

        cache.writeFragment(
          gql`
            fragment TaskCreate on Task {
              id
              title
              status
              input
              result
              flowId
              subtasks
              createdAt
              updatedAt
            }
          `,
          task
        );

        const tasksQuery = `tasks({"flowId":"${flowId}"})`;
        const tasksResult = cache.resolve("Query", tasksQuery) as string[];
        if (tasksResult) {
          const existingTask = tasksResult.find(t => cache.resolve(t, "id") === task.id);
          if (!existingTask) {
            cache.link("Query", tasksQuery, [task, ...tasksResult]);
          }
        } else {
          cache.link("Query", tasksQuery, [task]);
        }
      },

      taskUpdated: (result, _args, cache) => {
        const task = result.taskUpdated as Data;

        cache.writeFragment(
          gql`
            fragment TaskUpdate on Task {
              id
              status
              result
              subtasks
              updatedAt
            }
          `,
          task
        );
      },

      messageLogAdded: (result, _args, cache) => {
        const messageLog = result.messageLogAdded as Data;
        const flowId = messageLog.flowId as string;

        cache.writeFragment(
          gql`
            fragment MessageLogAdd on MessageLog {
              id
              type
              message
              result
              flowId
              taskId
              subtaskId
              createdAt
            }
          `,
          messageLog
        );

        const msgLogsQuery = `messageLogs({"flowId":"${flowId}"})`;
        const msgLogsResult = cache.resolve("Query", msgLogsQuery) as string[];
        if (msgLogsResult) {
          const existingMsgLog = msgLogsResult.find(m => cache.resolve(m, "id") === messageLog.id);
          if (!existingMsgLog) {
            cache.link("Query", msgLogsQuery, [...msgLogsResult, messageLog]);
          }
        } else {
          cache.link("Query", msgLogsQuery, [messageLog]);
        }
      },

      messageLogUpdated: (result, _args, cache) => {
        const messageLog = result.messageLogUpdated as Data;

        cache.writeFragment(
          gql`
            fragment MessageLogUpdate on MessageLog {
              id
              type
              message
              result
              resultFormat
              flowId
              taskId
              subtaskId
              createdAt
            }
          `,
          messageLog
        );
      },

      terminalLogAdded: (result, _args, cache) => {
        const terminalLog = result.terminalLogAdded as Data;
        const flowId = terminalLog.flowId as string;

        cache.writeFragment(
          gql`
            fragment TerminalLogAdd on TerminalLog {
              id
              flowId
              type
              text
              terminal
              createdAt
            }
          `,
          terminalLog
        );

        const termLogsQuery = `terminalLogs({"flowId":"${flowId}"})`;
        const termLogsResult = cache.resolve("Query", termLogsQuery) as string[];
        if (termLogsResult) {
          const existingTermLog = termLogsResult.find(t => cache.resolve(t, "id") === terminalLog.id);
          if (!existingTermLog) {
            cache.link("Query", termLogsQuery, [...termLogsResult, terminalLog]);
          }
        } else {
          cache.link("Query", termLogsQuery, [terminalLog]);
        }
      },

      screenshotAdded: (result, _args, cache) => {
        const screenshot = result.screenshotAdded as Data;
        const flowId = screenshot.flowId as string;
        
        cache.writeFragment(
          gql`
            fragment ScreenshotAdd on Screenshot {
              id
              flowId
              name
              url
              createdAt
            }
          `,
          screenshot
        );

        const screenshotsQuery = `screenshots({"flowId":"${flowId}"})`;
        const screenshotsResult = cache.resolve("Query", screenshotsQuery) as string[];
        if (screenshotsResult) {
          const existingScreenshot = screenshotsResult.find(s => cache.resolve(s, "id") === screenshot.id);
          if (!existingScreenshot) {
            cache.link("Query", screenshotsQuery, [...screenshotsResult, screenshot]);
          }
        } else {
          cache.link("Query", screenshotsQuery, [screenshot]);
        }
      },

      agentLogAdded: (result, _args, cache) => {
        const agentLog = result.agentLogAdded as Data;
        const flowId = agentLog.flowId as string;
        
        cache.writeFragment(
          gql`
            fragment AgentLogAdd on AgentLog {
              id
              flowId
              initiator
              executor
              task
              result
              taskId
              subtaskId
              createdAt
            }
          `,
          agentLog
        );

        const agentLogsQuery = `agentLogs({"flowId":"${flowId}"})`;
        const agentLogsResult = cache.resolve("Query", agentLogsQuery) as string[];
        if (agentLogsResult) {
          const existingAgentLog = agentLogsResult.find(a => cache.resolve(a, "id") === agentLog.id);
          if (!existingAgentLog) {
            cache.link("Query", agentLogsQuery, [...agentLogsResult, agentLog]);
          }
        } else {
          cache.link("Query", agentLogsQuery, [agentLog]);
        }
      },

      searchLogAdded: (result, _args, cache) => {
        const searchLog = result.searchLogAdded as Data;
        const flowId = searchLog.flowId as string;
        
        cache.writeFragment(
          gql`
            fragment SearchLogAdd on SearchLog {
              id
              flowId
              initiator
              executor
              engine
              query
              result
              taskId
              subtaskId
              createdAt
            }
          `,
          searchLog
        );

        const searchLogsQuery = `searchLogs({"flowId":"${flowId}"})`;
        const searchLogsResult = cache.resolve("Query", searchLogsQuery) as string[];
        if (searchLogsResult) {
          const existingSearchLog = searchLogsResult.find(s => cache.resolve(s, "id") === searchLog.id);
          if (!existingSearchLog) {
            cache.link("Query", searchLogsQuery, [...searchLogsResult, searchLog]);
          }
        } else {
          cache.link("Query", searchLogsQuery, [searchLog]);
        }
      },

      vectorStoreLogAdded: (result, _args, cache) => {
        const vectorStoreLog = result.vectorStoreLogAdded as Data;
        const flowId = vectorStoreLog.flowId as string;
        
        cache.writeFragment(
          gql`
            fragment VectorStoreLogAdd on VectorStoreLog {
              id
              flowId
              initiator
              executor
              filter
              query
              action
              result
              taskId
              subtaskId
              createdAt
            }
          `,
          vectorStoreLog
        );

        const vectorStoreLogsQuery = `vectorStoreLogs({"flowId":"${flowId}"})`;
        const vectorStoreLogsResult = cache.resolve("Query", vectorStoreLogsQuery) as string[];
        if (vectorStoreLogsResult) {
          const existingVectorStoreLog = vectorStoreLogsResult.find(v => cache.resolve(v, "id") === vectorStoreLog.id);
          if (!existingVectorStoreLog) {
            cache.link("Query", vectorStoreLogsQuery, [...vectorStoreLogsResult, vectorStoreLog]);
          }
        } else {
          cache.link("Query", vectorStoreLogsQuery, [vectorStoreLog]);
        }
      },
    },
  },
  resolvers: {
    Query: {
      flow: (_parent, args) => {
        const flowId = args.flowId as string;
        return `Flow:${flowId}`;
      },
    },
  },
  keys: {
    Flow: (data) => data.id as string,
    Task: (data) => data.id as string,
    MessageLog: (data) => data.id as string,
    TerminalLog: (data) => data.id as string,
    Screenshot: (data) => data.id as string,
  },
});

let wsClient: Client | null = null;

const getWSClient = () => {
  if (!wsClient) {
    const url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${baseURL}/graphql`;
    wsClient = createWSClient({
      url: url,
    });
  }
  return wsClient;
};

export const graphqlClient = createClient({
  url: `${window.location.origin}${baseURL}/graphql`,
  fetchOptions: {},
  exchanges: [
    devtoolsExchange,
    cache,
    fetchExchange,
    subscriptionExchange({
      forwardSubscription(request) {
        const input = { ...request, query: request.query || "" };
        const wsClient = getWSClient();
        return {
          subscribe(sink) {
            const unsubscribe = wsClient.subscribe(input, sink);

            return { unsubscribe };
          },
        };
      },
    }),
  ],
});
