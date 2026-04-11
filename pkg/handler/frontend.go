package handler

import (
	"fmt"
	"net/http"
)

func handlerFrontend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, frontendHTML)
}

const frontendHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>MGT Service - Admin Panel</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f172a;color:#e2e8f0;min-height:100vh}
.login-page{display:flex;align-items:center;justify-content:center;min-height:100vh}
.login-box{background:#1e293b;padding:2.5rem;border-radius:12px;width:380px;box-shadow:0 4px 20px rgba(0,0,0,.4)}
.login-box h1{text-align:center;margin-bottom:1.5rem;color:#60a5fa;font-size:1.5rem}
.app{display:none}
.sidebar{position:fixed;left:0;top:0;bottom:0;width:220px;background:#1e293b;padding:1rem 0;border-right:1px solid #334155}
.sidebar h2{padding:0 1rem;margin-bottom:1.5rem;color:#60a5fa;font-size:1.1rem}
.sidebar a{display:block;padding:.6rem 1rem;color:#94a3b8;text-decoration:none;font-size:.9rem;cursor:pointer;transition:all .15s}
.sidebar a:hover,.sidebar a.active{background:#334155;color:#f1f5f9}
.main{margin-left:220px;padding:1.5rem 2rem}
.topbar{display:flex;justify-content:space-between;align-items:center;margin-bottom:1.5rem;padding-bottom:1rem;border-bottom:1px solid #334155}
.topbar h1{font-size:1.3rem;color:#f1f5f9}
.btn{padding:.5rem 1rem;border:none;border-radius:6px;cursor:pointer;font-size:.85rem;font-weight:500;transition:all .15s}
.btn-primary{background:#3b82f6;color:#fff}.btn-primary:hover{background:#2563eb}
.btn-danger{background:#ef4444;color:#fff}.btn-danger:hover{background:#dc2626}
.btn-sm{padding:.35rem .7rem;font-size:.8rem}
.btn-outline{background:transparent;border:1px solid #475569;color:#94a3b8}.btn-outline:hover{border-color:#60a5fa;color:#60a5fa}
.card{background:#1e293b;border-radius:10px;padding:1.5rem;margin-bottom:1rem;border:1px solid #334155}
.form-group{margin-bottom:1rem}
.form-group label{display:block;margin-bottom:.4rem;font-size:.85rem;color:#94a3b8}
.form-group input,.form-group select{width:100%;padding:.55rem .75rem;border:1px solid #475569;border-radius:6px;background:#0f172a;color:#e2e8f0;font-size:.9rem;outline:none}
.form-group input:focus,.form-group select:focus{border-color:#3b82f6}
.form-row{display:flex;gap:1rem}.form-row .form-group{flex:1}
table{width:100%;border-collapse:collapse;font-size:.88rem}
th{text-align:left;padding:.6rem .75rem;background:#334155;color:#94a3b8;font-weight:500;font-size:.8rem;text-transform:uppercase;letter-spacing:.05em}
td{padding:.6rem .75rem;border-bottom:1px solid #1e293b}
tr:hover td{background:#1e293b80}
.badge{display:inline-block;padding:.15rem .5rem;border-radius:4px;font-size:.75rem;font-weight:500}
.badge-blue{background:#1e3a5f;color:#60a5fa}
.badge-green{background:#14532d;color:#4ade80}
.badge-red{background:#7f1d1d;color:#fca5a5}
.toast{position:fixed;top:1rem;right:1rem;padding:.75rem 1.25rem;border-radius:8px;font-size:.85rem;z-index:9999;animation:slideIn .3s ease}
.toast-success{background:#065f46;color:#6ee7b7;border:1px solid #059669}
.toast-error{background:#7f1d1d;color:#fca5a5;border:1px solid #dc2626}
@keyframes slideIn{from{transform:translateX(100%);opacity:0}to{transform:translateX(0);opacity:1}}
.section{display:none}.section.active{display:block}
.empty{text-align:center;padding:2rem;color:#64748b}
.actions{display:flex;gap:.5rem;flex-wrap:wrap}
.stat-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:1rem;margin-bottom:1.5rem}
.stat-card{background:#1e293b;border:1px solid #334155;border-radius:10px;padding:1.2rem}
.stat-card .num{font-size:1.8rem;font-weight:700;color:#60a5fa}
.stat-card .label{font-size:.8rem;color:#64748b;margin-top:.25rem}
.loading{text-align:center;padding:2rem;color:#64748b}
/* Autocomplete */
.ac-wrap{position:relative}
.ac-wrap input{width:100%;padding:.55rem .75rem;border:1px solid #475569;border-radius:6px;background:#0f172a;color:#e2e8f0;font-size:.9rem;outline:none}
.ac-wrap input:focus{border-color:#3b82f6}
.ac-list{position:absolute;top:100%;left:0;right:0;max-height:200px;overflow-y:auto;background:#1e293b;border:1px solid #475569;border-top:none;border-radius:0 0 6px 6px;z-index:100;display:none}
.ac-list.show{display:block}
.ac-item{padding:.45rem .75rem;cursor:pointer;font-size:.85rem;color:#e2e8f0}
.ac-item:hover,.ac-item.active{background:#334155;color:#60a5fa}
.ac-item .ac-sub{font-size:.75rem;color:#64748b;margin-left:.5rem}
</style>
</head>
<body>

<!-- Login -->
<div class="login-page" id="loginPage">
<div class="login-box">
  <h1>MGT Service</h1>
  <div class="form-group">
    <label>Username</label>
    <input type="text" id="loginUser" placeholder="admin" autofocus>
  </div>
  <div class="form-group">
    <label>Password</label>
    <input type="password" id="loginPass" placeholder="password">
  </div>
  <button class="btn btn-primary" style="width:100%;margin-top:.5rem" onclick="doLogin()">Login</button>
</div>
</div>

<!-- App -->
<div class="app" id="app">

<div class="sidebar">
  <h2>MGT Service</h2>
  <a onclick="showSection('dashboard')" class="active" data-section="dashboard">Dashboard</a>
  <a onclick="showSection('users')" data-section="users">Users</a>
  <a onclick="showSection('permissions')" data-section="permissions">Permissions</a>
  <a onclick="showSection('nes')" data-section="nes">Network Elements</a>
  <a onclick="showSection('ne-mapping')" data-section="ne-mapping">NE Mapping</a>
  <a onclick="showSection('role-mapping')" data-section="role-mapping">Role Mapping</a>
  <a onclick="showSection('history')" data-section="history">History</a>
  <a onclick="showSection('import')" data-section="import">Import</a>
</div>

<div class="main">
<div class="topbar">
  <h1 id="pageTitle">Dashboard</h1>
  <div>
    <span style="color:#64748b;font-size:.85rem;margin-right:1rem" id="currentUser"></span>
    <button class="btn btn-outline btn-sm" onclick="doLogout()">Logout</button>
  </div>
</div>

<!-- Dashboard -->
<div class="section active" id="sec-dashboard">
  <div class="stat-grid" id="statsGrid"></div>
</div>

<!-- Users -->
<div class="section" id="sec-users">
  <div class="card">
    <h3 style="margin-bottom:1rem">Create User</h3>
    <div class="form-row">
      <div class="form-group"><label>Username</label><input id="newUsername" placeholder="username"></div>
      <div class="form-group"><label>Password</label><input id="newPassword" type="password" placeholder="password"></div>
      <div class="form-group" style="flex:0;display:flex;align-items:flex-end"><button class="btn btn-primary" onclick="createUser()">Create</button></div>
    </div>
  </div>
  <div class="card">
    <h3 style="margin-bottom:1rem">All Users</h3>
    <div id="usersTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- Permissions -->
<div class="section" id="sec-permissions">
  <div class="card">
    <h3 style="margin-bottom:1rem">Create Permission</h3>
    <div class="form-row">
      <div class="form-group"><label>Permission</label><input id="newPermission" placeholder="admin"></div>
      <div class="form-group"><label>Scope</label><input id="newScope" placeholder="ext-config"></div>
      <div class="form-group"><label>NE Type</label><input id="newNeType" placeholder="5GC"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Include Type</label><input id="newIncludeType" placeholder="include"></div>
      <div class="form-group"><label>Path</label><input id="newPath" placeholder="/"></div>
      <div class="form-group" style="flex:0;display:flex;align-items:flex-end"><button class="btn btn-primary" onclick="createPermission()">Create</button></div>
    </div>
  </div>
  <div class="card">
    <h3 style="margin-bottom:1rem">All Permissions</h3>
    <div id="permissionsTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- NEs -->
<div class="section" id="sec-nes">
  <div class="card">
    <h3 style="margin-bottom:1rem">Create Network Element</h3>
    <div class="form-row">
      <div class="form-group"><label>Name</label><input id="neCreateName" placeholder="HTSMF03"></div>
      <div class="form-group"><label>Site</label><input id="neCreateSite" placeholder="HCM"></div>
      <div class="form-group"><label>IP Address</label><input id="neCreateIP" placeholder="10.10.1.3"></div>
    </div>
    <div class="form-row">
      <div class="form-group"><label>Port</label><input id="neCreatePort" placeholder="22" type="number"></div>
      <div class="form-group"><label>Namespace</label><input id="neCreateNS" placeholder="hcm-5gc"></div>
      <div class="form-group"><label>Description</label><input id="neCreateDesc" placeholder="HCM SMF Node 03"></div>
      <div class="form-group" style="flex:0;display:flex;align-items:flex-end"><button class="btn btn-primary" onclick="createNe()">Create</button></div>
    </div>
  </div>
  <div class="card">
    <h3 style="margin-bottom:1rem">All Network Elements</h3>
    <div id="nesTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- NE Mapping -->
<div class="section" id="sec-ne-mapping">
  <div class="card">
    <h3 style="margin-bottom:1rem">Assign NE to User</h3>
    <div class="form-row">
      <div class="form-group"><label>Username</label><div class="ac-wrap" id="acNeMapUser"></div></div>
      <div class="form-group"><label>Network Element</label><div class="ac-wrap" id="acNeMapNe"></div></div>
      <div class="form-group" style="flex:0;display:flex;align-items:flex-end"><button class="btn btn-primary" onclick="assignNe()">Assign</button></div>
    </div>
  </div>
  <div class="card">
    <h3 style="margin-bottom:1rem">User - NE Assignments</h3>
    <div id="neMappingTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- Role Mapping -->
<div class="section" id="sec-role-mapping">
  <div class="card">
    <h3 style="margin-bottom:1rem">Assign Role to User</h3>
    <div class="form-row">
      <div class="form-group"><label>Username</label><div class="ac-wrap" id="acRoleMapUser"></div></div>
      <div class="form-group"><label>Permission</label><div class="ac-wrap" id="acRoleMapPerm"></div></div>
      <div class="form-group" style="flex:0;display:flex;align-items:flex-end"><button class="btn btn-primary" onclick="assignRole()">Assign</button></div>
    </div>
  </div>
  <div class="card">
    <h3 style="margin-bottom:1rem">User - Role Assignments</h3>
    <div id="roleMappingTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- History -->
<div class="section" id="sec-history">
  <div class="card">
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:1rem">
      <h3>Operation History</h3>
      <div style="display:flex;gap:.5rem;align-items:center">
        <select id="historyLimit" style="padding:.4rem .6rem;border:1px solid #475569;border-radius:6px;background:#0f172a;color:#e2e8f0;font-size:.85rem">
          <option value="50">Last 50</option>
          <option value="100" selected>Last 100</option>
          <option value="200">Last 200</option>
          <option value="500">Last 500</option>
        </select>
        <button class="btn btn-outline btn-sm" onclick="loadHistory()">Refresh</button>
      </div>
    </div>
    <div id="historyTable"><div class="loading">Loading...</div></div>
  </div>
</div>

<!-- Import -->
<div class="section" id="sec-import">
  <div class="card">
    <h3 style="margin-bottom:1rem">Import Data</h3>
    <p style="color:#94a3b8;font-size:.85rem;margin-bottom:1rem">Upload a file or paste import data below. Format: CSV sections separated by [section_name] headers.</p>
    <div style="margin-bottom:1rem">
      <input type="file" id="importFile" accept=".txt,.csv" style="display:none" onchange="handleImportFile(event)">
      <button class="btn btn-outline btn-sm" onclick="document.getElementById('importFile').click()">Upload File</button>
      <button class="btn btn-outline btn-sm" onclick="fillSampleImport()">Load Sample</button>
    </div>
    <div class="form-group">
      <textarea id="importText" rows="20" style="width:100%;padding:.75rem;border:1px solid #475569;border-radius:6px;background:#0f172a;color:#e2e8f0;font-size:.82rem;font-family:monospace;resize:vertical;outline:none" placeholder="[users]
username,password
admin,admin123

[nes]
name,site_name,ip_address,port,namespace,description
HTSMF01,HCM,10.10.1.1,22,hcm-5gc,HCM SMF Node 01

[roles]
permission,scope,ne_type,include_type,path
admin,ext-config,5GC,include,/

[user_roles]
username,permission
admin,admin

[user_nes]
username,ne_name
admin,HTSMF01"></textarea>
    </div>
    <button class="btn btn-primary" onclick="runImport()">Import</button>
  </div>
  <div class="card" id="importResultsCard" style="display:none">
    <h3 style="margin-bottom:1rem">Import Results</h3>
    <div id="importResults"></div>
  </div>
</div>

</div><!-- main -->
</div><!-- app -->

<script>
let TOKEN = localStorage.getItem('mgt_token') || '';
let USERNAME = localStorage.getItem('mgt_user') || '';
const API = '/aa';

// cached data for autocomplete
let cachedUsers = [];
let cachedNes = [];
let cachedPerms = [];

// ── Helpers ──
async function api(method, path, body) {
  const opts = {method, headers: {'Content-Type':'application/json'}};
  if (TOKEN) opts.headers['Authorization'] = TOKEN;
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(API + path, opts);
  let data;
  try { data = await res.json(); } catch { data = null; }
  return {status: res.status, data};
}

function toast(msg, type='success') {
  const el = document.createElement('div');
  el.className = 'toast toast-' + type;
  el.textContent = msg;
  document.body.appendChild(el);
  setTimeout(() => el.remove(), 3000);
}

function esc(s) { const d=document.createElement('div'); d.textContent=s; return d.innerHTML; }

// ── Autocomplete ──
function initAC(containerId, placeholder, getItems) {
  const wrap = document.getElementById(containerId);
  const input = document.createElement('input');
  input.placeholder = placeholder;
  input.autocomplete = 'off';
  const list = document.createElement('div');
  list.className = 'ac-list';
  wrap.appendChild(input);
  wrap.appendChild(list);

  let selectedValue = '';
  let activeIdx = -1;

  function render(query) {
    const items = getItems();
    const q = query.toLowerCase();
    const filtered = q ? items.filter(i => i.label.toLowerCase().includes(q) || (i.sub||'').toLowerCase().includes(q)) : items;
    activeIdx = -1;
    if (filtered.length === 0) { list.innerHTML = '<div class="ac-item" style="color:#64748b">No matches</div>'; list.classList.add('show'); return; }
    list.innerHTML = filtered.map((item, idx) =>
      '<div class="ac-item" data-val="' + esc(item.value) + '" data-label="' + esc(item.label) + '" data-idx="' + idx + '">'
      + esc(item.label) + (item.sub ? '<span class="ac-sub">' + esc(item.sub) + '</span>' : '')
      + '</div>'
    ).join('');
    list.classList.add('show');
  }

  input.addEventListener('focus', () => render(input.value));
  input.addEventListener('input', () => { selectedValue = ''; render(input.value); });
  input.addEventListener('keydown', e => {
    const items = list.querySelectorAll('.ac-item[data-val]');
    if (e.key === 'ArrowDown') { e.preventDefault(); activeIdx = Math.min(activeIdx+1, items.length-1); items.forEach((el,i) => el.classList.toggle('active', i===activeIdx)); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); activeIdx = Math.max(activeIdx-1, 0); items.forEach((el,i) => el.classList.toggle('active', i===activeIdx)); }
    else if (e.key === 'Enter') { e.preventDefault(); if (activeIdx >= 0 && items[activeIdx]) items[activeIdx].click(); }
    else if (e.key === 'Escape') { list.classList.remove('show'); }
  });

  list.addEventListener('click', e => {
    const item = e.target.closest('.ac-item[data-val]');
    if (!item) return;
    selectedValue = item.dataset.val;
    input.value = item.dataset.label;
    list.classList.remove('show');
  });

  document.addEventListener('click', e => { if (!wrap.contains(e.target)) list.classList.remove('show'); });

  return { getValue: () => selectedValue || input.value, clear: () => { input.value = ''; selectedValue = ''; } };
}

// ── Section navigation ──
function showSection(name) {
  document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
  document.getElementById('sec-' + name).classList.add('active');
  document.querySelectorAll('.sidebar a').forEach(a => a.classList.toggle('active', a.dataset.section === name));
  const titles = {dashboard:'Dashboard',users:'Users',permissions:'Permissions',nes:'Network Elements','ne-mapping':'NE Mapping','role-mapping':'Role Mapping',history:'History','import':'Import'};
  document.getElementById('pageTitle').textContent = titles[name] || name;
  if (name === 'dashboard') loadDashboard();
  if (name === 'users') loadUsers();
  if (name === 'permissions') loadPermissions();
  if (name === 'nes') loadNes();
  if (name === 'ne-mapping') loadNeMapping();
  if (name === 'role-mapping') loadRoleMapping();
  if (name === 'history') loadHistory();
}

// ── Auth ──
async function doLogin() {
  const username = document.getElementById('loginUser').value;
  const password = document.getElementById('loginPass').value;
  const {status, data} = await api('POST', '/authenticate', {username, password});
  if (status === 200 && data && data.response_data) {
    TOKEN = data.response_data;
    USERNAME = username;
    localStorage.setItem('mgt_token', TOKEN);
    localStorage.setItem('mgt_user', USERNAME);
    enterApp();
  } else {
    toast('Login failed', 'error');
  }
}

function doLogout() {
  TOKEN = ''; USERNAME = '';
  localStorage.removeItem('mgt_token');
  localStorage.removeItem('mgt_user');
  document.getElementById('app').style.display = 'none';
  document.getElementById('loginPage').style.display = 'flex';
}

async function refreshCaches() {
  const [u, n, p] = await Promise.all([
    api('GET', '/authenticate/user/show'),
    api('GET', '/authorize/ne/show'),
    api('GET', '/authorize/permission/show')
  ]);
  cachedUsers = Array.isArray(u.data) ? u.data : [];
  cachedNes = Array.isArray(n.data) ? n.data : [];
  cachedPerms = Array.isArray(p.data) ? p.data : [];
}

// autocomplete instances
let acNeMapUser, acNeMapNe, acRoleMapUser, acRoleMapPerm;

function enterApp() {
  document.getElementById('loginPage').style.display = 'none';
  document.getElementById('app').style.display = 'block';
  document.getElementById('currentUser').textContent = USERNAME;

  // init autocompletes
  acNeMapUser = initAC('acNeMapUser', 'Search user...', () => cachedUsers.map(u => ({label: u.username, value: u.username, sub: u.role||''})));
  acNeMapNe = initAC('acNeMapNe', 'Search NE...', () => cachedNes.map(n => ({label: n.name, value: String(n.id), sub: n.site_name + ' - ' + n.ip_address})));
  acRoleMapUser = initAC('acRoleMapUser', 'Search user...', () => cachedUsers.map(u => ({label: u.username, value: u.username, sub: u.role||''})));
  acRoleMapPerm = initAC('acRoleMapPerm', 'Search permission...', () => {
    const unique = [...new Set(cachedPerms.map(p => p.permission))];
    return unique.map(p => ({label: p, value: p}));
  });

  refreshCaches().then(() => showSection('dashboard'));
}

document.getElementById('loginPass').addEventListener('keydown', e => { if (e.key === 'Enter') doLogin(); });
document.getElementById('loginUser').addEventListener('keydown', e => { if (e.key === 'Enter') document.getElementById('loginPass').focus(); });

// ── Dashboard ──
async function loadDashboard() {
  document.getElementById('statsGrid').innerHTML =
    '<div class="stat-card"><div class="num">' + cachedUsers.length + '</div><div class="label">Users</div></div>' +
    '<div class="stat-card"><div class="num">' + cachedPerms.length + '</div><div class="label">Permissions</div></div>' +
    '<div class="stat-card"><div class="num">' + cachedNes.length + '</div><div class="label">Network Elements</div></div>';
}

// ── Users ──
async function loadUsers() {
  const data = cachedUsers;
  if (data.length === 0) { document.getElementById('usersTable').innerHTML = '<div class="empty">No users found</div>'; return; }
  let html = '<table><thead><tr><th>Username</th><th>Role</th><th>NEs</th><th>Actions</th></tr></thead><tbody>';
  data.forEach(u => {
    const nes = (u.tblNes || []).map(n => '<span class="badge badge-blue">' + esc(n.ne) + '</span>').join(' ') || '-';
    const role = u.role ? '<span class="badge badge-green">' + esc(u.role) + '</span>' : '-';
    html += '<tr><td>' + esc(u.username) + '</td><td>' + role + '</td><td>' + nes + '</td>';
    html += '<td class="actions"><button class="btn btn-danger btn-sm" onclick="deleteUser(\'' + esc(u.username) + '\')">Disable</button></td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('usersTable').innerHTML = html;
}

async function createUser() {
  const username = document.getElementById('newUsername').value;
  const password = document.getElementById('newPassword').value;
  if (!username || !password) { toast('Fill in all fields', 'error'); return; }
  const {status} = await api('POST', '/authenticate/user/set', {account_name: username, password: password});
  if (status === 201) { toast('User created'); document.getElementById('newUsername').value=''; document.getElementById('newPassword').value=''; await refreshCaches(); loadUsers(); }
  else toast('Failed to create user', 'error');
}

async function deleteUser(username) {
  if (!confirm('Disable user ' + username + '?')) return;
  const {status} = await api('POST', '/authenticate/user/delete', {account_name: username});
  if (status === 200) { toast('User disabled'); await refreshCaches(); loadUsers(); }
  else toast('Failed to disable user', 'error');
}

// ── Permissions ──
async function loadPermissions() {
  const data = cachedPerms;
  if (data.length === 0) { document.getElementById('permissionsTable').innerHTML = '<div class="empty">No permissions found</div>'; return; }
  let html = '<table><thead><tr><th>ID</th><th>Permission</th><th>Scope</th><th>NE Type</th><th>Include Type</th><th>Path</th><th>Actions</th></tr></thead><tbody>';
  data.forEach(p => {
    html += '<tr><td>' + (p.role_id||p.id||p.ID||'-') + '</td><td><span class="badge badge-green">' + esc(p.permission||'-') + '</span></td>';
    html += '<td>' + esc(p.scope||'-') + '</td><td>' + esc(p.ne_type||'-') + '</td>';
    html += '<td>' + esc(p.include_type||'-') + '</td><td>' + esc(p.path||'-') + '</td>';
    html += '<td class="actions"><button class="btn btn-danger btn-sm" onclick="deletePermission(\'' + esc(p.permission||'') + '\',\'' + esc(p.scope||'') + '\',\'' + esc(p.ne_type||'') + '\',\'' + esc(p.include_type||'') + '\',\'' + esc(p.path||'') + '\')">Delete</button></td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('permissionsTable').innerHTML = html;
}

async function createPermission() {
  const permission = document.getElementById('newPermission').value;
  const scope = document.getElementById('newScope').value;
  const ne_type = document.getElementById('newNeType').value;
  const include_type = document.getElementById('newIncludeType').value;
  const path = document.getElementById('newPath').value;
  if (!permission) { toast('Permission name is required', 'error'); return; }
  const {status} = await api('POST', '/authorize/permission/set', {permission, scope, ne_type, include_type, path});
  if (status === 201) { toast('Permission created'); ['newPermission','newScope','newNeType','newIncludeType','newPath'].forEach(id => document.getElementById(id).value=''); await refreshCaches(); loadPermissions(); }
  else toast('Failed to create permission', 'error');
}

async function deletePermission(permission, scope, ne_type, include_type, path) {
  if (!confirm('Delete permission ' + permission + '?')) return;
  const {status} = await api('POST', '/authorize/permission/delete', {permission, scope, ne_type, include_type, path});
  if (status === 200) { toast('Permission deleted'); await refreshCaches(); loadPermissions(); }
  else toast('Failed to delete permission', 'error');
}

// ── NEs ──
async function loadNes() {
  const data = cachedNes;
  if (data.length === 0) { document.getElementById('nesTable').innerHTML = '<div class="empty">No NEs found</div>'; return; }
  let html = '<table><thead><tr><th>ID</th><th>Name</th><th>Site</th><th>IP Address</th><th>Port</th><th>Namespace</th><th>Description</th><th>Actions</th></tr></thead><tbody>';
  data.forEach(n => {
    html += '<tr><td>' + n.id + '</td><td><span class="badge badge-blue">' + esc(n.name) + '</span></td>';
    html += '<td>' + esc(n.site_name||'-') + '</td><td>' + esc(n.ip_address||'-') + '</td>';
    html += '<td>' + (n.port||'-') + '</td><td>' + esc(n.namespace||'-') + '</td><td>' + esc(n.description||'-') + '</td>';
    html += '<td class="actions"><button class="btn btn-danger btn-sm" onclick="deleteNe(' + n.id + ',\'' + esc(n.name) + '\')">Delete</button></td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('nesTable').innerHTML = html;
}

async function deleteNe(id, name) {
  if (!confirm('Delete NE ' + name + ' (ID ' + id + ')?')) return;
  const {status} = await api('POST', '/authorize/ne/remove', {id});
  if (status === 200) { toast('NE deleted'); await refreshCaches(); loadNes(); }
  else toast('Failed to delete NE', 'error');
}

async function createNe() {
  const name = document.getElementById('neCreateName').value;
  const site_name = document.getElementById('neCreateSite').value;
  const ip_address = document.getElementById('neCreateIP').value;
  const port = parseInt(document.getElementById('neCreatePort').value) || 22;
  const namespace = document.getElementById('neCreateNS').value;
  const description = document.getElementById('neCreateDesc').value;
  if (!name) { toast('Name is required', 'error'); return; }
  const {status} = await api('POST', '/authorize/ne/create', {name, site_name, ip_address, port, namespace, description, system_type: '5GC'});
  if (status === 201) {
    toast('NE created');
    ['neCreateName','neCreateSite','neCreateIP','neCreatePort','neCreateNS','neCreateDesc'].forEach(id => document.getElementById(id).value='');
    await refreshCaches(); loadNes();
  } else toast('Failed to create NE', 'error');
}

// ── NE Mapping ──
async function loadNeMapping() {
  const data = cachedUsers;
  const neIdMap = {};
  cachedNes.forEach(n => { neIdMap[n.name] = n.id; });

  if (data.length === 0) { document.getElementById('neMappingTable').innerHTML = '<div class="empty">No data</div>'; return; }
  let html = '<table><thead><tr><th>Username</th><th>Assigned NEs</th><th>Actions</th></tr></thead><tbody>';
  data.forEach(u => {
    const nes = (u.tblNes || []);
    const nesBadges = nes.map(n => '<span class="badge badge-blue">' + esc(n.ne) + ' (' + esc(n.site) + ')</span>').join(' ') || '<span style="color:#64748b">None</span>';
    const removeButtons = nes.map(n => {
      const neid = neIdMap[n.ne] || '';
      return '<button class="btn btn-danger btn-sm" onclick="removeNe(\'' + esc(u.username) + '\',' + neid + ')">' + esc(n.ne) + ' &times;</button>';
    }).join(' ');
    html += '<tr><td>' + esc(u.username) + '</td><td>' + nesBadges + '</td><td class="actions">' + (removeButtons || '-') + '</td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('neMappingTable').innerHTML = html;
}

async function removeNe(username, neid) {
  if (!confirm('Remove NE ' + neid + ' from ' + username + '?')) return;
  const {status} = await api('POST', '/authorize/ne/delete', {username, neid: String(neid)});
  if (status === 200) { toast('NE removed'); await refreshCaches(); loadNeMapping(); }
  else toast('Failed to remove NE', 'error');
}

async function assignNe() {
  const username = acNeMapUser.getValue();
  const neid = acNeMapNe.getValue();
  if (!username || !neid) { toast('Select user and NE', 'error'); return; }
  const {status} = await api('POST', '/authorize/ne/set', {username, neid});
  if (status === 200) { toast('NE assigned'); acNeMapUser.clear(); acNeMapNe.clear(); await refreshCaches(); loadNeMapping(); }
  else toast('Failed to assign NE', 'error');
}

// ── Role Mapping ──
async function loadRoleMapping() {
  const {data} = await api('GET', '/authorize/user/show');
  if (!Array.isArray(data) || data.length === 0) {
    document.getElementById('roleMappingTable').innerHTML = '<div class="empty">No data</div>';
    return;
  }
  let html = '<table><thead><tr><th>Username</th><th>Permissions</th><th>Actions</th></tr></thead><tbody>';
  data.forEach(u => {
    const permList = u.permissions ? u.permissions.split(' ').filter(Boolean) : [];
    const permsBadges = permList.length > 0
      ? permList.map(p => '<span class="badge badge-green">' + esc(p) + '</span>').join(' ')
      : '<span style="color:#64748b">None</span>';
    const removeButtons = permList.map(p =>
      '<button class="btn btn-danger btn-sm" onclick="removeRole(\'' + esc(u.username) + '\',\'' + esc(p) + '\')">' + esc(p) + ' &times;</button>'
    ).join(' ');
    html += '<tr><td>' + esc(u.username) + '</td><td>' + permsBadges + '</td>';
    html += '<td class="actions">' + (removeButtons || '-') + '</td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('roleMappingTable').innerHTML = html;
}

async function assignRole() {
  const username = acRoleMapUser.getValue();
  const permission = acRoleMapPerm.getValue();
  if (!username || !permission) { toast('Select user and permission', 'error'); return; }
  const {status} = await api('POST', '/authorize/user/set', {username, permission});
  if (status === 200) { toast('Role assigned'); acRoleMapUser.clear(); acRoleMapPerm.clear(); await refreshCaches(); loadRoleMapping(); }
  else toast('Failed to assign role', 'error');
}

async function removeRole(username, permission) {
  if (!confirm('Remove permission ' + permission + ' from ' + username + '?')) return;
  const {status} = await api('POST', '/authorize/user/delete', {username, permission});
  if (status === 200) { toast('Role removed'); loadRoleMapping(); }
  else toast('Failed to remove role', 'error');
}

// ── History ──
async function loadHistory() {
  const limit = document.getElementById('historyLimit').value;
  const {status, data} = await api('GET', '/history/list?limit=' + limit);
  if (status !== 200 || !Array.isArray(data) || data.length === 0) {
    document.getElementById('historyTable').innerHTML = '<div class="empty">No history records</div>';
    return;
  }
  let html = '<table><thead><tr><th>Time</th><th>User</th><th>Command</th><th>NE</th><th>Result</th><th>Scope</th></tr></thead><tbody>';
  data.forEach(h => {
    const time = h.executed_time ? new Date(h.executed_time).toLocaleString() : '-';
    const resultBadge = h.result === 'success'
      ? '<span class="badge badge-green">success</span>'
      : h.result === 'failure'
        ? '<span class="badge badge-red">failure</span>'
        : '<span class="badge badge-blue">' + esc(h.result||'-') + '</span>';
    html += '<tr><td style="white-space:nowrap;font-size:.8rem">' + time + '</td>';
    html += '<td>' + esc(h.account||'-') + '</td>';
    html += '<td style="font-family:monospace;font-size:.8rem">' + esc(h.cmd_name||'-') + '</td>';
    html += '<td>' + esc(h.ne_name||'-') + '</td>';
    html += '<td>' + resultBadge + '</td>';
    html += '<td>' + esc(h.scope||'-') + '</td></tr>';
  });
  html += '</tbody></table>';
  document.getElementById('historyTable').innerHTML = html;
}

// ── Import ──
function handleImportFile(event) {
  const file = event.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = e => { document.getElementById('importText').value = e.target.result; };
  reader.readAsText(file);
}

function fillSampleImport() {
  document.getElementById('importText').value = '[users]\nusername,password\nadmin,admin123\noperator1,Pass@123\n\n[nes]\nname,site_name,ip_address,port,namespace,description\nHTSMF01,HCM,10.10.1.1,22,hcm-5gc,HCM SMF Node 01\nHTAMF01,HCM,10.10.2.1,22,hcm-5gc,HCM AMF Node 01\n\n[roles]\npermission,scope,ne_type,include_type,path\nadmin,ext-config,5GC,include,/\noperator,ext-config,5GC,include,/\n\n[user_roles]\nusername,permission\nadmin,admin\noperator1,operator\n\n[user_nes]\nusername,ne_name\nadmin,HTSMF01\noperator1,HTAMF01';
}

async function runImport() {
  const text = document.getElementById('importText').value;
  if (!text.trim()) { toast('Paste or upload import data first', 'error'); return; }

  const res = await fetch(API + '/import/', {
    method: 'POST',
    headers: {'Content-Type': 'text/plain', 'Authorization': TOKEN},
    body: text
  });
  let data;
  try { data = await res.json(); } catch { data = null; }

  if (!Array.isArray(data) || data.length === 0) {
    toast('No results returned', 'error');
    return;
  }

  const card = document.getElementById('importResultsCard');
  card.style.display = 'block';
  let html = '<table><thead><tr><th>Type</th><th>Name</th><th>Status</th><th>Detail</th></tr></thead><tbody>';
  let okCount = 0, errCount = 0, skipCount = 0;
  data.forEach(r => {
    const badge = r.status === 'ok' ? '<span class="badge badge-green">ok</span>'
      : r.status === 'skip' ? '<span class="badge badge-blue">skip</span>'
      : '<span class="badge badge-red">error</span>';
    if (r.status === 'ok') okCount++;
    else if (r.status === 'skip') skipCount++;
    else errCount++;
    html += '<tr><td>' + esc(r.type) + '</td><td>' + esc(r.name) + '</td><td>' + badge + '</td><td style="font-size:.8rem">' + esc(r.detail) + '</td></tr>';
  });
  html += '</tbody></table>';
  html += '<p style="margin-top:.75rem;font-size:.85rem;color:#94a3b8">'
    + '<span class="badge badge-green">' + okCount + ' ok</span> '
    + '<span class="badge badge-blue">' + skipCount + ' skipped</span> '
    + '<span class="badge badge-red">' + errCount + ' errors</span></p>';
  document.getElementById('importResults').innerHTML = html;

  await refreshCaches();
  toast('Import complete: ' + okCount + ' ok, ' + skipCount + ' skipped, ' + errCount + ' errors');
}

// ── Init ──
async function checkToken() {
  if (!TOKEN) return;
  const res = await fetch(API + '/validate-token', {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({token: TOKEN})
  });
  if (res.status === 200) { enterApp(); }
  else { doLogout(); }
}
checkToken();
</script>
</body>
</html>`
