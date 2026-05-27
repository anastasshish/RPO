export function stripForApi(entity, item) {
  const o = { ...item };
  delete o.key;
  delete o.card;
  delete o.terminal;
  delete o.cards;
  if (entity === 'users') delete o.password_hash;
  return o;
}

export function coerceForApi(entity, raw, isAdmin) {
  const o = { ...raw };
  const reqInt = (label, v, min = 0) => {
    if (v === '' || v === undefined || v === null) return null;
    const n = parseInt(String(v).trim(), 10);
    if (!Number.isFinite(n) || n < min) throw new Error(`Некорректное число: ${label}`);
    return n;
  };
  const optInt = (v) => {
    if (v === '' || v === undefined || v === null) return 0;
    const n = parseInt(String(v).trim(), 10);
    if (!Number.isFinite(n) || n < 0) throw new Error('Некорректное целое число');
    return n;
  };
  const optBalance = (v) => {
    if (v === '' || v === undefined || v === null) return 0;
    const n = parseInt(String(v).trim().replace(/\s/g, ''), 10);
    if (!Number.isFinite(n)) throw new Error('Некорректный баланс');
    return n;
  };

  if (entity === 'cards') {
    o.balance = optBalance(o.balance);
    o.key_id = optInt(o.key_id);
    if (isAdmin) {
      o.user_id = o.user_id === '' || o.user_id === undefined || o.user_id === null ? 0 : optInt(o.user_id);
    } else {
      delete o.user_id;
    }
  }
  if (entity === 'transactions') {
    const amt = reqInt('сумма', o.amount, 1);
    const cid = reqInt('ID карты', o.card_id, 1);
    const tid = reqInt('ID терминала', o.terminal_id, 1);
    if (amt === null) throw new Error('Укажите сумму');
    if (cid === null) throw new Error('Укажите ID карты');
    if (tid === null) throw new Error('Укажите ID терминала');
    o.amount = amt;
    o.card_id = cid;
    o.terminal_id = tid;
  }
  return o;
}

export const EMPTY_ENTITY = {
  cards: { number: '', balance: '', blocked: false, owner_name: '', key_id: '', user_id: '' },
  terminals: { serial_number: '', name: '', address: '', description: '' },
  transactions: { amount: '', card_id: '', terminal_id: '', status: 'pending', message: '' },
  keys: { name: '', key_value: '', description: '' },
  users: { login: '', name: '', password: '', is_admin: false },
};

export const TITLE_FOR_TAB = {
  cards: 'Карты',
  terminals: 'Терминалы',
  transactions: 'Транзакции',
  keys: 'Ключи MIFARE',
  users: 'Пользователи',
  profile: 'Мой профиль',
};
