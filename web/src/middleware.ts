import { defineMiddleware } from 'astro:middleware';

export const onRequest = defineMiddleware(async (context, next) => {
  const token = context.cookies.get('oscar_token')?.value;
  const userJson = context.cookies.get('oscar_user')?.value;

  if (token && userJson) {
    try {
      context.locals.user = JSON.parse(userJson);
      context.locals.token = token;
    } catch {
      context.locals.user = null;
    }
  } else {
    // Also check localStorage via a client-side check
    // For SSR, we rely on cookies set by the login endpoint
    context.locals.user = null;
  }

  return next();
});
