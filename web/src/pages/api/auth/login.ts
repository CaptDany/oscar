import type { APIRoute } from 'astro';

const ACCESS_TOKEN_MAX_AGE = 60 * 15;
const REFRESH_TOKEN_MAX_AGE = 60 * 60 * 24 * 7;

export const POST: APIRoute = async ({ request, cookies }) => {
  try {
    const body = await request.json();
    const { email, password } = body;

    const backendRes = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    const data = await backendRes.json();

    if (!backendRes.ok) {
      return new Response(JSON.stringify({
        error: data.error || { message: 'Login failed' }
      }), {
        status: backendRes.status,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    cookies.set('oscar_token', data.access_token, {
      path: '/',
      httpOnly: false,
      sameSite: 'lax',
      maxAge: ACCESS_TOKEN_MAX_AGE,
    });

    if (data.refresh_token) {
      cookies.set('oscar_refresh_token', data.refresh_token, {
        path: '/',
        httpOnly: false,
        sameSite: 'lax',
        maxAge: REFRESH_TOKEN_MAX_AGE,
      });
    }

    cookies.set('oscar_user', JSON.stringify(data.user), {
      path: '/',
      httpOnly: false,
      sameSite: 'lax',
      maxAge: REFRESH_TOKEN_MAX_AGE,
    });

    return new Response(JSON.stringify({
      success: true,
      token: data.access_token,
      user: data.user,
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'An error occurred during login' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
