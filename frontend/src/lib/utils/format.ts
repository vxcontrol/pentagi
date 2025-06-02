import { format, isThisYear, isToday } from 'date-fns';
import { enUS } from 'date-fns/locale';

export const formatName = (name?: string): string =>
    (name || '')
        .split('_')
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');

export const formatDate = (date: Date) => {
    if (isToday(date)) {
        return format(date, 'HH:mm:ss');
    }

    if (isThisYear(date)) {
        return format(date, 'HH:mm, d MMM', { locale: enUS });
    }

    return format(date, 'HH:mm, d MMM yyyy', { locale: enUS });
};
