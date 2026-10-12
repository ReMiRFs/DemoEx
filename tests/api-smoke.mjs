// Run against a TEST database: creates two accounts and removes its bookings.
// API_URL=http://127.0.0.1:8080 ADMIN_LOGIN=... ADMIN_PASSWORD=... node tests/api-smoke.mjs
import assert from 'node:assert/strict';

const base = (process.env.API_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
const suffix = Date.now().toString(36);
const password = 'DemoSmoke123!';

function client() {
  let cookie = '';
  return async (path, method = 'GET', data) => {
    const response = await fetch(base + path, {
      method,
      headers: { ...(cookie ? { Cookie: cookie } : {}), ...(data ? { 'Content-Type': 'application/json' } : {}) },
      body: data ? JSON.stringify(data) : undefined,
      redirect: 'manual',
    });
    const setCookie = response.headers.get('set-cookie');
    if (setCookie) cookie = setCookie.split(';')[0];
    const text = await response.text();
    let body = null;
    if (text) {
      assert.match(response.headers.get('content-type') || '', /application\/json/, `${path}: JSON expected, got ${text.slice(0, 160)}`);
      body = JSON.parse(text);
    }
    return { status: response.status, body, setCookie };
  };
}

const anonymous = client();
const first = client();
const second = client();
const admin = client();
const account = (index) => ({
  login: `smoke${suffix}${index}`, password,
  fullName: 'Иван Иванов', phone: '8(999)123-45-67', email: `smoke${suffix}${index}@example.com`,
});
const booking = { room: 'Коворкинг', startsAt: '2027-10-20', payment: 'cash' };

assert.equal((await anonymous('/api/health')).status, 200);
assert.equal((await anonymous('/api/requests')).status, 401);
assert.equal((await anonymous('/api/missing')).status, 404);
const bad = await anonymous('/api/auth/register', 'POST', { ...account('bad'), login: 'x' });
assert.equal(bad.status, 400);
assert.equal(typeof bad.body.error, 'string');
for (const [api, index] of [[first, 'a'], [second, 'b']]) {
  const registered = await api('/api/auth/register', 'POST', account(index));
  assert.equal(registered.status, 201);
  assert.equal(registered.body.role, 'user');
  assert.ok(!('password' in registered.body) && !('passwordHash' in registered.body));
  assert.match(registered.setCookie, /HttpOnly/i);
  assert.equal((await api('/api/auth/me')).body.login, account(index).login);
}
assert.equal((await anonymous('/api/auth/register', 'POST', account('a'))).status, 409);
assert.equal((await anonymous('/api/auth/login', 'POST', { login: account('a').login, password: 'wrong' })).status, 401);
assert.deepEqual((await first('/api/requests')).body, []);
assert.equal((await first('/api/requests', 'POST', { ...booking, startsAt: '2027-02-30' })).status, 400);
assert.equal((await first('/api/requests', 'POST', { ...booking, room: ' ' })).status, 400);
const created = await first('/api/requests', 'POST', booking);
assert.equal(created.status, 201);
const id = created.body.id;
assert.equal(typeof id, 'number');
assert.equal(created.body.status, 'new');
assert.equal(created.body.startsAt, booking.startsAt);
assert.deepEqual((await second('/api/requests')).body, []);
for (const method of ['PUT', 'DELETE']) {
  const forbidden = await second(`/api/requests/${id}`, method, method === 'PUT' ? booking : undefined);
  assert.ok([403, 404].includes(forbidden.status), `${method}: ownership not enforced`);
}
assert.ok([400, 403, 409].includes((await first(`/api/requests/${id}/review`, 'POST', { review: 'Отлично' })).status));
assert.equal((await first('/api/admin/requests')).status, 403);
const updated = await first(`/api/requests/${id}`, 'PUT', { ...booking, room: 'Аудитория', payment: 'sbp' });
assert.equal(updated.status, 200);
assert.equal(updated.body.room, 'Аудитория');
assert.equal(updated.body.payment, 'sbp');

if (process.env.ADMIN_LOGIN && process.env.ADMIN_PASSWORD) {
  assert.equal((await admin('/api/auth/login', 'POST', { login: process.env.ADMIN_LOGIN, password: process.env.ADMIN_PASSWORD })).body.role, 'admin');
  assert.equal((await admin(`/api/admin/requests/${id}/status`, 'PATCH', { status: 'done' })).status, 200);
  assert.ok((await admin('/api/admin/requests?status=done')).body.some(row => row.id === id));
  assert.equal((await first(`/api/requests/${id}/review`, 'POST', { review: 'Отлично' })).body.review, 'Отлично');
  assert.equal((await admin(`/api/admin/requests/${id}/status`, 'PATCH', { status: 'unknown' })).status, 400);
} else {
  console.log('SKIP admin lifecycle: set ADMIN_LOGIN and ADMIN_PASSWORD to include it.');
}
assert.equal((await first(`/api/requests/${id}`, 'DELETE')).status, 204);
assert.deepEqual((await first('/api/requests')).body, []);
assert.equal((await first('/api/auth/logout', 'POST')).status, 204);
assert.equal((await first('/api/auth/me')).status, 401);
console.log('PASS: registration, session, validation, ownership, CRUD, JSON errors and logout.');
