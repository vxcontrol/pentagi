import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { Badge } from '@/components/ui/badge';
import { StatusType } from '@/graphql/types';

const STATUS_LABELS: Record<StatusType, string> = {
    [StatusType.Created]: 'Created',
    [StatusType.Failed]: 'Failed',
    [StatusType.Finished]: 'Finished',
    [StatusType.Running]: 'Running',
    [StatusType.Waiting]: 'Waiting',
};

export const FlowStatusBadge = ({ className, status }: { className?: string; status: StatusType }) => (
    <Badge
        className={className}
        variant="outline"
    >
        <FlowStatusIcon
            className="size-3"
            status={status}
        />
        {STATUS_LABELS[status]}
    </Badge>
);
