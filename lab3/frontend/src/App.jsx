import React, { useCallback, useEffect, useState } from 'react';
import { LogOut, Pencil, Plus, Trash2 } from 'lucide-react';
import { API } from './constants.js';
import { useApi } from './hooks/useApi.js';
import { useMainTabs } from './app/useMainTabs.js';
import { useTableColumns } from './app/useTableColumns.jsx';
import { coerceForApi, EMPTY_ENTITY, stripForApi, TITLE_FOR_TAB } from './app/entityApi.js';
import { Modal } from './components/Modal.jsx';
import { DataTable } from './components/DataTable.jsx';
import { CrudModalFields } from './components/CrudModalFields.jsx';

export function App() {
  const [token, setToken] = useState(localStorage.getItem('token') || '');
  const [me, setMe] = useState(() => {
    try {
      const s = localStorage.getItem('me');
      return s ? JSON.parse(s) : null;
    } catch {
      return null;
    }
  });
  const [login, setLogin] = useState('admin');
  const [password, setPassword] = useState('admin123');
  const [tab, setTab] = useState('cards');
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: '', text: '' });
  const [pay, setPay] = useState({
    card_number: '',
    amount: '',
    terminal_serial: '',
  });
  const [modal, setModal] = useState(null);
  const [keysForSelect, setKeysForSelect] = useState([]);

  const api = useApi(token);
  const isAdmin = !!me?.is_admin;
  const tabs = useMainTabs(isAdmin);
  const columns = useTableColumns(tab, isAdmin);

  const showMsg = (type, text) => {
    setMsg({ type, text });
    if (text) setTimeout(() => setMsg((m) => (m.text === text ? { type: '', text: '' } : m)), 5000);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('me');
    setToken('');
    setMe(null);
    setRows([]);
    setTab('cards');
  };

  const doLogin = async (e) => {
    e.preventDefault();
    try {
      const j = await fetch(`${API}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ login, password }),
      }).then(async (r) => {
        const t = await r.text();
        const j = t ? JSON.parse(t) : {};
        if (!r.ok) throw new Error(j.error || 'Ошибка входа');
        return j;
      });
      if (!j.token) {
        showMsg('err', 'Нет токена в ответе');
        return;
      }
      localStorage.setItem('token', j.token);
      localStorage.setItem('me', JSON.stringify(j.user));
      setToken(j.token);
      setMe(j.user);
      showMsg('ok', `Добро пожаловать, ${j.user?.login || ''}`);
    } catch (err) {
      showMsg('err', err.message);
    }
  };

  const loadList = useCallback(
    async (t) => {
      if (!token) return;
      if (t === 'profile') {
        setRows(me ? [me] : []);
        return;
      }
      setLoading(true);
      try {
        const data = await api('GET', `/${t}`);
        setRows(Array.isArray(data) ? data : []);
      } catch (err) {
        setRows([]);
        showMsg('err', err.status === 403 ? 'Нет прав на этот раздел' : err.message);
      } finally {
        setLoading(false);
      }
    },
    [api, token, me]
  );

  useEffect(() => {
    if (token && me) loadList(tab);
  }, [token, me, tab, loadList]);

  useEffect(() => {
    if (!isAdmin) return;
    if (modal?.entity !== 'cards') return;
    (async () => {
      try {
        const k = await api('GET', '/keys');
        setKeysForSelect(Array.isArray(k) ? k : []);
      } catch {
        setKeysForSelect([]);
      }
    })();
  }, [isAdmin, modal, api]);

  const openCreate = (entity) => {
    setModal({ mode: 'create', entity, item: { ...(EMPTY_ENTITY[entity] || {}) } });
  };

  const openEdit = (entity, item) => {
    const copy = JSON.parse(JSON.stringify(item));
    if (entity === 'users') {
      copy.password = '';
      delete copy.password_hash;
    }
    const numKeys = ['balance', 'key_id', 'amount', 'card_id', 'terminal_id', 'user_id'];
    for (const k of numKeys) {
      if (copy[k] !== undefined && copy[k] !== null && copy[k] !== '') copy[k] = String(copy[k]);
    }
    setModal({ mode: 'edit', entity, item: copy });
  };

  const saveModal = async () => {
    if (!modal) return;
    const { mode, entity, item } = modal;
    const pathBase = `/${entity}`;
    try {
      if (mode === 'create') {
        let body = stripForApi(entity, item);
        if (entity === 'users' && !body.password) {
          showMsg('err', 'Укажите пароль');
          return;
        }
        if (entity === 'cards' || entity === 'transactions') {
          body = coerceForApi(entity, body, isAdmin);
        }
        await api('POST', pathBase, body);
        showMsg('ok', 'Создано');
      } else {
        const id = item.id;
        let body = stripForApi(entity, item);
        if (entity === 'cards' || entity === 'transactions') {
          body = coerceForApi(entity, body, isAdmin);
        }
        const saved = await api('PUT', `${pathBase}/${id}`, body);
        showMsg('ok', 'Сохранено');
        if (entity === 'users' && me?.id === id && saved) {
          const u = { ...saved };
          delete u.password;
          setMe(u);
          localStorage.setItem('me', JSON.stringify(u));
        }
      }
      setModal(null);
      loadList(tab === 'profile' ? 'profile' : tab);
    } catch (err) {
      showMsg('err', err.message);
    }
  };

  const removeRow = async (entity, id) => {
    if (!window.confirm('Удалить запись?')) return;
    try {
      await api('DELETE', `/${entity}/${id}`);
      showMsg('ok', 'Удалено');
      loadList(tab);
    } catch (err) {
      showMsg('err', err.message);
    }
  };

  const authorize = async () => {
    const amt = parseInt(String(pay.amount).trim(), 10);
    if (!pay.card_number?.trim() || !pay.terminal_serial?.trim() || !Number.isFinite(amt) || amt <= 0) {
      showMsg('err', 'Укажите номер карты, сумму (> 0) и серийный номер терминала');
      return;
    }
    try {
      const r = await fetch(`${API}/terminal/payments/authorize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          card_number: pay.card_number.trim(),
          terminal_serial: pay.terminal_serial.trim(),
          amount: amt,
        }),
      });
      const j = await r.json();
      if (!r.ok) {
        showMsg('err', j.message || j.error || 'Ошибка');
        return;
      }
      showMsg(
        'ok',
        j.authorized
          ? `Одобрено. Баланс карты: ${j.balance}. ${j.message || ''}`
          : `Отклонено: ${j.message || ''}`
      );
      if (token && tab === 'transactions') loadList('transactions');
    } catch (e) {
      showMsg('err', e.message);
    }
  };

  useEffect(() => {
    const allowed = tabs.map((t) => t.id);
    if (!allowed.includes(tab)) setTab(allowed[0] || 'cards');
  }, [tabs, tab]);

  const canCrud = tab !== 'profile';
  const userCanMutateTerminals = isAdmin;
  const showAddButton =
    canCrud &&
    (tab !== 'users' || isAdmin) &&
    !(tab === 'terminals' && !userCanMutateTerminals);
  const showRowActions =
    canCrud && !(tab === 'users' && !isAdmin) && !(tab === 'terminals' && !userCanMutateTerminals);

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">
          <h1>Transport Auth</h1>
          <p className="muted small">Карты · терминалы · платежи</p>
        </div>
        {!token ? null : (
          <>
            <div className="user-chip">
              <span className="dot" data-admin={isAdmin} />
              <div>
                <strong>{me?.login}</strong>
                <div className="muted small">{isAdmin ? 'Администратор' : 'Пользователь'}</div>
              </div>
            </div>
            <nav className="nav">
              {tabs.map((t) => {
                const Icon = t.icon;
                return (
                  <button
                    key={t.id}
                    type="button"
                    className={`nav-item ${tab === t.id ? 'active' : ''}`}
                    onClick={() => setTab(t.id)}
                  >
                    <Icon size={18} />
                    {t.label}
                  </button>
                );
              })}
            </nav>
            <button type="button" className="btn ghost nav-logout" onClick={logout}>
              <LogOut size={18} /> Выйти
            </button>
          </>
        )}
      </aside>

      <main className="content">
        {!token ? (
          <section className="card login-card">
            <h2>Вход</h2>
            <form onSubmit={doLogin} className="form-grid">
              <label>Логин</label>
              <input value={login} onChange={(e) => setLogin(e.target.value)} autoComplete="username" />
              <label>Пароль</label>
              <input
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                type="password"
                autoComplete="current-password"
              />
              <div className="form-actions">
                <button type="submit" className="btn primary">
                  Войти
                </button>
              </div>
            </form>
          </section>
        ) : (
          <>
            {msg.text ? (
              <div className={`banner ${msg.type === 'ok' ? 'banner-ok' : 'banner-err'}`}>{msg.text}</div>
            ) : null}

            <section className="card">
              {tab === 'transactions' ? (
                <>
                  <h2 className="card-title-top">Транзакции</h2>
                  <p className="muted small">
                    Чтобы провести оплату, укажите номер карты, сумму и серийный номер терминала. При одобрении с
                    карты спишется сумма и в журнале ниже появится транзакция.
                  </p>
                  <div className="inline-fields pay-row">
                    <input
                      value={pay.card_number}
                      onChange={(e) => setPay({ ...pay, card_number: e.target.value })}
                      placeholder="Номер карты"
                    />
                    <input
                      type="text"
                      inputMode="numeric"
                      autoComplete="off"
                      value={pay.amount}
                      onChange={(e) => setPay({ ...pay, amount: e.target.value })}
                      placeholder="Сумма списания"
                    />
                    <input
                      value={pay.terminal_serial}
                      onChange={(e) => setPay({ ...pay, terminal_serial: e.target.value })}
                      placeholder="Серийный № терминала"
                    />
                    <button type="button" className="btn primary" onClick={authorize}>
                      Провести оплату
                    </button>
                  </div>
                  <h3 className="subheading">Журнал</h3>
                </>
              ) : null}

              <div className={`section-head ${tab === 'transactions' ? 'section-head-compact' : ''}`}>
                {tab !== 'transactions' ? <h2>{TITLE_FOR_TAB[tab]}</h2> : null}
                {showAddButton ? (
                  tab === 'users' ? (
                    <button type="button" className="btn primary" onClick={() => openCreate('users')}>
                      <Plus size={18} /> Пользователь
                    </button>
                  ) : (
                    <button type="button" className="btn primary" onClick={() => openCreate(tab)}>
                      <Plus size={18} /> Добавить
                    </button>
                  )
                ) : null}
              </div>

              {tab === 'profile' ? (
                <div className="profile-actions">
                  <p className="muted">Редактирование своего профиля (логин, имя, пароль).</p>
                  <button type="button" className="btn secondary" onClick={() => openEdit('users', me)}>
                    <Pencil size={18} /> Изменить профиль
                  </button>
                </div>
              ) : null}

              {loading ? <p className="muted">Загрузка…</p> : null}

              {tab === 'terminals' && !isAdmin ? (
                <p className="muted small">Терминалы можно только просматривать. Создание и изменение — у администратора.</p>
              ) : null}

              {!loading ? (
                <DataTable
                  columns={columns}
                  rows={rows}
                  rowKey={(r) => r.id}
                  actions={
                    showRowActions
                      ? (r) => (
                          <div className="row-actions">
                            <button type="button" className="btn icon" title="Изменить" onClick={() => openEdit(tab, r)}>
                              <Pencil size={16} />
                            </button>
                            <button
                              type="button"
                              className="btn icon danger"
                              title="Удалить"
                              onClick={() => removeRow(tab, r.id)}
                            >
                              <Trash2 size={16} />
                            </button>
                          </div>
                        )
                      : null
                  }
                />
              ) : null}
            </section>
          </>
        )}
      </main>

      {modal ? (
        <Modal title={modal.mode === 'create' ? 'Создание' : 'Редактирование'} onClose={() => setModal(null)}>
          <div className="form-stack">
            <CrudModalFields modal={modal} isAdmin={isAdmin} keysForSelect={keysForSelect} setModal={setModal} />
          </div>
          <div className="modal-actions">
            <button type="button" className="btn ghost" onClick={() => setModal(null)}>
              Отмена
            </button>
            <button type="button" className="btn primary" onClick={saveModal}>
              Сохранить
            </button>
          </div>
        </Modal>
      ) : null}
    </div>
  );
}
