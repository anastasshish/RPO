import React from 'react';

export function CrudModalFields({ modal, isAdmin, keysForSelect, setModal }) {
  if (!modal) return null;
  const { entity, item } = modal;
  const set = (patch) => setModal((m) => (m ? { ...m, item: { ...m.item, ...patch } } : m));

  if (entity === 'cards') {
    return (
      <>
        <label>Номер карты</label>
        <input value={item.number} onChange={(e) => set({ number: e.target.value })} disabled={modal.mode === 'edit'} />
        <label>Баланс</label>
        <input
          type="text"
          inputMode="numeric"
          autoComplete="off"
          value={item.balance ?? ''}
          onChange={(e) => set({ balance: e.target.value })}
        />
        <label className="row">
          <input type="checkbox" checked={!!item.blocked} onChange={(e) => set({ blocked: e.target.checked })} />
          Заблокирована
        </label>
        <label>Владелец (имя)</label>
        <input value={item.owner_name || ''} onChange={(e) => set({ owner_name: e.target.value })} />
        {isAdmin ? (
          <>
            <label>Владелец (ID пользователя)</label>
            <input
              type="text"
              inputMode="numeric"
              autoComplete="off"
              placeholder="пусто при редактировании — не менять"
              value={item.user_id ?? ''}
              onChange={(e) => set({ user_id: e.target.value })}
            />
          </>
        ) : null}
        <label>Ключ (ID)</label>
        {isAdmin && keysForSelect.length ? (
          <select
            value={item.key_id === '' || item.key_id === 0 ? '' : String(item.key_id)}
            onChange={(e) => set({ key_id: e.target.value })}
          >
            <option value="">— выберите ключ</option>
            {keysForSelect.map((k) => (
              <option key={k.id} value={String(k.id)}>
                {k.id} — {k.name}
              </option>
            ))}
          </select>
        ) : (
          <input
            type="text"
            inputMode="numeric"
            autoComplete="off"
            placeholder="ID ключа"
            value={item.key_id ?? ''}
            onChange={(e) => set({ key_id: e.target.value })}
          />
        )}
      </>
    );
  }
  if (entity === 'terminals') {
    return (
      <>
        <label>Серийный номер</label>
        <input value={item.serial_number} onChange={(e) => set({ serial_number: e.target.value })} />
        <label>Название</label>
        <input value={item.name || ''} onChange={(e) => set({ name: e.target.value })} />
        <label>Адрес</label>
        <input value={item.address || ''} onChange={(e) => set({ address: e.target.value })} />
        <label>Описание</label>
        <input value={item.description || ''} onChange={(e) => set({ description: e.target.value })} />
      </>
    );
  }
  if (entity === 'transactions') {
    return (
      <>
        <label>Сумма</label>
        <input
          type="text"
          inputMode="numeric"
          autoComplete="off"
          value={item.amount ?? ''}
          onChange={(e) => set({ amount: e.target.value })}
        />
        <label>ID карты</label>
        <input
          type="text"
          inputMode="numeric"
          autoComplete="off"
          value={item.card_id ?? ''}
          onChange={(e) => set({ card_id: e.target.value })}
        />
        <label>ID терминала</label>
        <input
          type="text"
          inputMode="numeric"
          autoComplete="off"
          value={item.terminal_id ?? ''}
          onChange={(e) => set({ terminal_id: e.target.value })}
        />
        <label>Статус</label>
        <input value={item.status || ''} onChange={(e) => set({ status: e.target.value })} />
        <label>Сообщение</label>
        <input value={item.message || ''} onChange={(e) => set({ message: e.target.value })} />
      </>
    );
  }
  if (entity === 'keys') {
    return (
      <>
        <label>Название</label>
        <input value={item.name} onChange={(e) => set({ name: e.target.value })} />
        <label>Значение ключа</label>
        <input value={item.key_value} onChange={(e) => set({ key_value: e.target.value })} />
        <label>Описание</label>
        <input value={item.description || ''} onChange={(e) => set({ description: e.target.value })} />
      </>
    );
  }
  if (entity === 'users') {
    return (
      <>
        <label>Логин</label>
        <input value={item.login} onChange={(e) => set({ login: e.target.value })} />
        <label>Имя</label>
        <input value={item.name || ''} onChange={(e) => set({ name: e.target.value })} />
        <label>Пароль {modal.mode === 'edit' ? '(оставьте пустым, чтобы не менять)' : ''}</label>
        <input type="password" value={item.password || ''} onChange={(e) => set({ password: e.target.value })} />
        {isAdmin ? (
          <label className="row">
            <input type="checkbox" checked={!!item.is_admin} onChange={(e) => set({ is_admin: e.target.checked })} />
            Администратор
          </label>
        ) : null}
      </>
    );
  }
  return null;
}
