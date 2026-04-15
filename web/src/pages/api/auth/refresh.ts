import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request, cookies }) => {
  try {
    const refreshToken = cookies.get('oscar_refresh_token')?.value;

    if (!refreshToken) {
      return new Response(JSON.stringify({
        error: { message: 'No refresh token' }
      }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const backendRes = await fetch('http://localhost:8080/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    const data = await backendRes.json();

    if (!backendRes.ok) {
      cookies.delete('oscar_token', { path: '/' });
      cookies.delete('oscar_refresh_token', { path: '/' });
      cookies.delete('oscar_user', { path: '/' });
      
      return new Response(JSON.stringify({
        error: data.error || { message: 'Token refresh failed' }
      }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    cookies.set('oscar_token', data.access_token, {
      path: '/',
      httpOnly: false,
      sameSite: 'lax',
      maxAge: 60 * 15,
    });

    return new Response(JSON.stringify({
      success: true,
      token: data.access_token,
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'An error occurred during token refresh' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
