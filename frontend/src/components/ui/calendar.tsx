import { ChevronLeft, ChevronRight } from 'lucide-react';
import { DayPicker, type DayPickerProps } from 'react-day-picker';

import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export type CalendarProps = DayPickerProps;

function Calendar({ className, classNames, showOutsideDays = true, ...props }: CalendarProps) {
    return (
        <DayPicker
            className={cn('p-3', className)}
            classNames={{
                button_next: cn(
                    buttonVariants({ variant: 'outline' }),
                    'absolute top-3 right-3 z-10 h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100',
                ),
                button_previous: cn(
                    buttonVariants({ variant: 'outline' }),
                    'absolute top-3 left-3 z-10 h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100',
                ),
                caption: 'flex justify-center pt-1 pb-2 relative items-center select-none',
                caption_label: 'text-sm font-medium select-none',
                day: cn(
                    buttonVariants({ variant: 'ghost' }),
                    'hover:bg-accent hover:text-accent-foreground h-9 w-9 p-0 font-normal aria-selected:opacity-100',
                ),
                day_button: 'h-9 w-9 p-0 font-normal',
                disabled: 'text-muted-foreground opacity-50',
                hidden: 'invisible',
                month: 'space-y-4',
                month_caption: 'flex justify-center pt-1 pb-2 relative items-center',
                month_grid: 'w-full border-collapse mt-4',
                months: 'flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0',
                nav: 'flex items-center m-0',
                outside:
                    'day-outside text-muted-foreground aria-selected:bg-accent/50 aria-selected:text-muted-foreground aria-selected:opacity-30',
                range_end: 'day-range-end',
                range_middle: 'aria-selected:bg-accent aria-selected:text-accent-foreground',
                selected:
                    'bg-primary text-primary-foreground hover:bg-primary hover:text-primary-foreground focus:bg-primary focus:text-primary-foreground',
                today: 'bg-accent text-accent-foreground',
                week: 'flex w-full mt-2 justify-center',
                weekday: 'text-muted-foreground rounded-md w-9 font-normal text-[0.8rem]',
                weekdays: 'flex justify-center',
                ...classNames,
            }}
            components={{
                Chevron: ({ orientation }) =>
                    orientation === 'left' ? <ChevronLeft className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />,
            }}
            showOutsideDays={showOutsideDays}
            {...props}
        />
    );
}

Calendar.displayName = 'Calendar';

export { Calendar };
