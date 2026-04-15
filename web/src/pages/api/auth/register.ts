import type { APIRoute } from 'astro';

const ACCESS_TOKEN_MAX_AGE = 60 * 15;
const REFRESH_TOKEN_MAX_AGE = 60 * 60 * 24 * 7;

async function setAuthCookies(cookies: any, data: any) {
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
}

async function loginAndSetCookies(cookies: any, email: string, password: string) {
  const loginRes = await fetch('http://localhost:8080/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });

  const loginData = await loginRes.json();

  if (!loginRes.ok) {
    return { success: false, error: loginData.error };
  }

  await setAuthCookies(cookies, loginData);
  return { success: true, data: loginData };
}

export const POST: APIRoute = async ({ request, cookies }) => {
  try {
    const body = await request.json();
    const { first_name, last_name, email, password, tenant_name } = body;

    const loginResult = await loginAndSetCookies(cookies, email, password);
    if (loginResult.success) {
      return new Response(JSON.stringify({
        success: true,
        token: loginResult.data.access_token,
        user: loginResult.data.user,
        already_registered: true,
      }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const registerRes = await fetch('http://localhost:8080/api/v1/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        first_name,
        last_name,
        email,
        password,
        tenant_name,
        tenant_slug: tenant_name.toLowerCase().replace(/\s+/g, '-'),
      }),
    });

    if (!registerRes.ok) {
      const data = await registerRes.json();
      return new Response(JSON.stringify({
        error: data.error || { message: 'Registration failed' }
      }), {
        status: registerRes.status,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const newLoginResult = await loginAndSetCookies(cookies, email, password);
    if (!newLoginResult.success) {
      return new Response(JSON.stringify({
        error: newLoginResult.error || { message: 'Registration succeeded but login failed' }
      }), {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response(JSON.stringify({
      success: true,
      token: newLoginResult.data.access_token,
      user: newLoginResult.data.user,
      already_registered: false,
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

  } catch (error) {
    return new Response(JSON.stringify({
      error: { message: 'An error occurred during registration' }
    }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
