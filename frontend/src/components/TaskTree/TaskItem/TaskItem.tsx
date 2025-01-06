import { useState } from "react";
import { Task, Subtask, StatusType } from "@/generated/graphql";
import { taskItemStyles, taskTitleStyles, subtaskListStyles, preWrapper } from "./TaskItem.css";
import { Icon } from "../../Icon/Icon";

type TaskItemProps = {
  task: Task;
};

const getStatusIcon = (status: StatusType) => {
  switch (status) {
    case StatusType.Starting:
      return Icon.BlankCircle;
    case StatusType.Running:
      return Icon.PlayCircle;
    case StatusType.Finished:
      return Icon.CheckCircle;
    case StatusType.Failed:
      return Icon.CloseCircle;
    case StatusType.Waiting:
      return Icon.HelpCircle;
    default:
      return Icon.BlankCircle;
  }
};

export const TaskItem = ({ task }: TaskItemProps) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const toggleExpanded = () => {
    setIsExpanded(!isExpanded);
  };

  const StatusIcon = getStatusIcon(task.status);

  return (
    <div className={taskItemStyles.wrapper}>
      <div className={taskTitleStyles} onClick={toggleExpanded}>
        <StatusIcon className={taskItemStyles.status[task.status]} />
        <span>{task.title}</span>
      </div>
      {isExpanded && task.result && (
        <div className={taskItemStyles.description}>
          <pre className={preWrapper}>
            {task.result}
          </pre>
        </div>
      )}
      {task.subtasks && task.subtasks.length > 0 && isExpanded && (
        <div className={subtaskListStyles}>
          {task.subtasks.sort((a, b) => Number(a.id) - Number(b.id)).map((subtask) => (
            <SubtaskItem key={subtask.id} subtask={subtask} />
          ))}
        </div>
      )}
    </div>
  );
};

type SubtaskItemProps = {
    subtask: Subtask;
};
  
export const SubtaskItem = ({ subtask }: SubtaskItemProps) => {
    const [isDescriptionVisible, setIsDescriptionVisible] = useState(false);

    const toggleDescription = () => {
        setIsDescriptionVisible(!isDescriptionVisible);
    };

    const StatusIcon = getStatusIcon(subtask.status);

    return (
      <div className={taskItemStyles.wrapper}>
      <div className={taskTitleStyles} onClick={toggleDescription}>
        <StatusIcon className={taskItemStyles.status[subtask.status]} />
        <span>{subtask.title}</span>
      </div>
      {isDescriptionVisible && subtask.description && (
        <div className={taskItemStyles.description}>
          <pre className={preWrapper}>
            {subtask.description}
          </pre>
        </div>
      )}
      {isDescriptionVisible && subtask.result && (
        <div className={taskItemStyles.description}>
          <pre className={preWrapper}>
            {subtask.result}
          </pre>
        </div>
      )}
      </div>
    );
}; 