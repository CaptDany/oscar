import type { APIRoute } from 'astro';

export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await request.json();
    const { email } = body;

    if (!email) {
      return new Response(JSON.stringify({
        error: { message: 'Email is required' }
      }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const backendRes = await fetch('http://localhost:8080/api/v1/auth/resend-verification', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    });

    const data = await backendRes.json();

    return new Response(JSON.stringify(data), {
      status: backendRes.status,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'An error occurred' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
