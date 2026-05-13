function platform() {
  var ua = navigator.userAgent;
  if (/Android/i.test(ua)) return 'android';
  if (/iPhone|iPad|iPod/i.test(ua)) return 'ios';
  return 'pc';
}

function getQuery(name) {
  var m = location.search.match(new RegExp('[?&]' + name + '=([^&]*)'));
  return m ? decodeURIComponent(m[1]) : '';
}

function showToast(msg, duration) {
  duration = duration || 2000;
  var el = document.querySelector('.toast');
  if (!el) {
    el = document.createElement('div');
    el.className = 'toast';
    document.body.appendChild(el);
  }
  el.textContent = msg;
  el.classList.add('show');
  clearTimeout(el._timer);
  el._timer = setTimeout(function () { el.classList.remove('show'); }, duration);
}

function getToken() {
  return localStorage.getItem('x-token');
}

function saveToken(data) {
  if (data.access_token) localStorage.setItem('x-token', data.access_token);
  if (data.sn) localStorage.setItem('x-sn', data.sn);
  if (data.refresh_token) localStorage.setItem('x-refresh-token', data.refresh_token);
}

function apiUrl(path) {
  var base = localStorage.getItem('x-api-base') || 'http://localhost:4462';
  return base + path;
}

function postJSON(url, data) {
  return fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  }).then(function (r) { return r.json(); });
}

function fetchState() {
  var apiBase = localStorage.getItem('x-api-auth') || 'http://localhost:4463';
  return postJSON(apiBase + '/auth/state', {}).then(function (res) {
    return res.data ? res.data.state : '';
  });
}

function setLoading(btn, loading) {
  if (loading) {
    btn.disabled = true;
    btn.classList.add('btn-loading');
  } else {
    btn.disabled = false;
    btn.classList.remove('btn-loading');
  }
}
