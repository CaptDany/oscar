import type { APIRoute } from 'astro';

export const prerender = false;

export const GET: APIRoute = async ({ request, cookies }) => {
  const url = new URL(request.url);
  const objectKey = url.searchParams.get('key');

  if (!objectKey) {
    return new Response(JSON.stringify({
      error: { message: 'object_key is required' }
    }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const token = cookies.get('oscar_token')?.value;

  try {
    const backendRes = await fetch(`http://localhost:8080/api/v1/upload/branding/asset?key=${encodeURIComponent(objectKey)}`, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Accept': 'image/*',
      },
      credentials: 'include',
    });

    if (!backendRes.ok) {
      return new Response(JSON.stringify({
        error: { message: 'Asset not found' }
      }), {
        status: backendRes.status,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const data = await backendRes.arrayBuffer();
    const contentType = backendRes.headers.get('Content-Type') || 'image/svg+xml';

    return new Response(data, {
      status: 200,
      headers: { 'Content-Type': contentType },
    });
  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'Failed to fetch asset' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};