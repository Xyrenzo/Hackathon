'use strict';

const API = 'http://localhost:8080/api';

/* ─── helpers ─── */
const $ = id => document.getElementById(id);

function setMsg(id, text, type = 'error') {
  const el = $(id);
  el.textContent = text;
  el.className = 'msg ' + type;
}

function clearMsg(id) {
  const el = $(id);
  el.textContent = '';
  el.className = 'msg';
}

function setLoading(btn, loading) {
  btn.classList.toggle('loading', loading);
  btn.disabled = loading;
}

async function apiFetch(path, options = {}) {
  const res = await fetch(API + path, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  });

  const data = await res.json().catch(() => ({}));

  return {
    ok: res.ok,
    status: res.status,
    data
  };
}

/* ─── tabs ─── */
function switchTab(tab) {
  const switcher = document.querySelector('.tab-switcher');

  switcher.dataset.active = tab;

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.remove('active');
  });

  document.querySelectorAll('.form-panel').forEach(panel => {
    panel.classList.add('hidden');
  });

  $('tab-' + tab).classList.add('active');
  $('panel-' + tab).classList.remove('hidden');
}

/* ════════════════════════════════
   LOGIN
════════════════════════════════ */

let loginName = '';

async function requestCode() {
  // FIX: читаем поле login-name (имя), а не login-phone
  const name = $('login-name').value.trim();

  clearMsg('login-step1-msg');

  if (!name) {
    setMsg('login-step1-msg', 'Введите имя');
    return;
  }

  const btn = $('btn-request-code');
  setLoading(btn, true);

  try {
    const { ok, data } = await apiFetch('/auth/login/request', {
      method: 'POST',
      body: JSON.stringify({ name })
    });

    if (ok) {
      loginName = name;
      $('login-code-hint').textContent = 'Код отправлен в Telegram';
      showStep('login', 1, 2);
    } else {
      setMsg('login-step1-msg', data.message || 'Ошибка отправки кода');
    }
  } catch {
    setMsg('login-step1-msg', 'Нет соединения с сервером');
  } finally {
    setLoading(btn, false);
  }
}

async function loginWithCode() {
  const code = $('login-code').value.trim();

  clearMsg('login-step2-msg');

  if (!code) {
    setMsg('login-step2-msg', 'Введите код');
    return;
  }

  const btn = document.querySelector('#login-step-2 .btn--primary');
  setLoading(btn, true);

  try {
    const { ok, data } = await apiFetch('/auth/login/verify', {
      method: 'POST',
      body: JSON.stringify({
        name: loginName,
        code
      })
    });

    if (ok) {
      const token = data.token;

      localStorage.setItem('jt_token', token);

      await loadProfile(token);

      showStep('login', 2, 3);
    } else {
      setMsg('login-step2-msg', data.message || 'Неверный код');
    }
  } catch {
    setMsg('login-step2-msg', 'Нет соединения с сервером');
  } finally {
    setLoading(btn, false);
  }
}

async function loadProfile(token) {
  try {
    const { ok, data } = await apiFetch('/profile', {
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    const el = $('profile-data');

    if (ok) {
      el.innerHTML = `
        <div><strong>Имя:</strong> ${esc(data.name || '—')}</div>
        <div><strong>Телефон:</strong> ${esc(data.phone || '—')}</div>
        <div><strong>Город:</strong> ${esc(data.city || '—')}</div>
      `;
    } else {
      el.innerHTML = '<div>Профиль недоступен</div>';
    }
  } catch {
    $('profile-data').innerHTML = '<div>Ошибка загрузки профиля</div>';
  }
}

function logout() {
  localStorage.removeItem('jt_token');

  loginName = '';

  // FIX: очищаем правильное поле login-name
  $('login-name').value = '';
  $('login-code').value = '';

  $('profile-data').innerHTML = '';

  showStep('login', 3, 1);
}

function goBack(panel) {
  if (panel === 'login') {
    showStep('login', 2, 1);
  }
}

/* ════════════════════════════════
   REGISTER
════════════════════════════════ */

let registerName = '';

async function register() {
  const name = $('reg-name').value.trim();
  const phone = $('reg-phone').value.trim();
  const city = $('reg-city').value.trim();

  clearMsg('reg-step1-msg');

  if (!name) {
    setMsg('reg-step1-msg', 'Введите имя');
    return;
  }

  if (!phone) {
    setMsg('reg-step1-msg', 'Введите номер');
    return;
  }

  if (!city) {
    setMsg('reg-step1-msg', 'Введите город');
    return;
  }

  const btn = document.querySelector('#reg-step-1 .btn--primary');
  setLoading(btn, true);

  try {
    const { ok, data } = await apiFetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        name,
        phone,
        city
      })
    });

    if (ok) {
      registerName = name;

      $('reg-tg-link').href = data.verificationLink;
      $('reg-tg-info').innerHTML = `
        <div>Открой Telegram и активируй аккаунт</div>
      `;

      showStep('register', 1, 2);

      startActivationPolling();
    } else {
      setMsg('reg-step1-msg', data.message || 'Ошибка регистрации');
    }
  } catch {
    setMsg('reg-step1-msg', 'Нет соединения с сервером');
  } finally {
    setLoading(btn, false);
  }
}

function startActivationPolling() {
  const interval = setInterval(async () => {
    try {
      const { ok, data } = await apiFetch('/auth/status?name=' + encodeURIComponent(registerName));

      if (ok && data.activated) {
        clearInterval(interval);

        $('reg-tg-info').innerHTML = `
          <div>Telegram успешно активирован</div>
        `;
      }
    } catch {}
  }, 3000);
}

/* ─── steps ─── */

function showStep(panel, from, to) {
  const prefix = panel === 'login' ? 'login-step-' : 'reg-step-';

  const fromEl = $(prefix + from);
  const toEl = $(prefix + to);

  if (fromEl) fromEl.classList.add('hidden');
  if (toEl) toEl.classList.remove('hidden');
}

/* ─── escape ─── */

function esc(str) {
  return String(str).replace(/[&<>"']/g, s => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  })[s]);
}

/* ─── init ─── */

(async function init() {
  const token = localStorage.getItem('jt_token');

  if (token) {
    await loadProfile(token);
    showStep('login', 1, 3);
  }
})();
