import { useTasksQuery, useTaskCreatedSubscription, useTaskUpdatedSubscription } from "@/generated/graphql";
import { TaskItem } from "./TaskItem/TaskItem";
import { taskTreeStyles } from "./TaskTree.css";

type TaskTreeProps = {
  flowId: string;
  visibility?: boolean;
};

export const TaskTree = ({ flowId, visibility = true }: TaskTreeProps) => {
  const [{ data }] = useTasksQuery({
    variables: { flowId },
    requestPolicy: "cache-and-network",
  });

  useTaskCreatedSubscription({
    variables: { flowId },
  });

  useTaskUpdatedSubscription({
    variables: { flowId },
  });

  const tasks = data?.tasks ?? [];

  return (
    <div className={taskTreeStyles[visibility ? "expanded" : "collapsed"]}>
      {tasks.sort((a, b) => Number(a.id) - Number(b.id)).map((task) => (
        <TaskItem key={task.id} task={task} />
      ))}
    </div>
  );
}; 