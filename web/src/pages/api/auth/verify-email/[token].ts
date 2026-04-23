import type { APIRoute } from 'astro';

export const prerender = false;

export const GET: APIRoute = async ({ params }) => {
  const token = params.token;

  if (!token) {
    return new Response(JSON.stringify({
      error: { message: 'Verification token is required' }
    }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  try {
    const backendRes = await fetch(`http://localhost:8080/api/v1/auth/verify-email/${token}`, {
      method: 'GET',
    });

    const data = await backendRes.json();

    if (!backendRes.ok) {
      return new Response(JSON.stringify({
        error: data.error || { message: 'Verification failed' }
      }), {
        status: backendRes.status,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response(JSON.stringify(data), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'Unable to connect to the server' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
