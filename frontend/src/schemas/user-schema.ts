import * as z from 'zod';

export const userFormSchema = z.object({
    name: z.string().min(1),
    email: z.string().email(),
});
