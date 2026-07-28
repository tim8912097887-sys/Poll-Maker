import z from 'zod';

export const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'test', 'production']),
  PORT: z.coerce.number().default(3333),
  DATABASE_URL: z.string(),
});

export default () => {
  const result = envSchema.safeParse(process.env);

  if (!result.success) {
    throw new Error(result.error.issues[0].message);
  }

  return result.data;
};
