import * as z from 'zod';

export const userFormSchema = z.object({
    email: z.string().email(),
    name: z.string().min(1),
});
