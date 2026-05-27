import React, { useMemo } from 'react';
import { fmtDate } from '../utils/format.js';

export function useTableColumns(tab, isAdmin) {
  return useMemo(() => {
    const cardCols = [
      { key: 'id', label: 'ID' },
      { key: 'number', label: 'Номер карты' },
      { key: 'balance', label: 'Баланс' },
      {
        key: 'blocked',
        label: 'Блок',
        render: (r) => (r.blocked ? 'да' : 'нет'),
      },
      { key: 'owner_name', label: 'Владелец' },
    ];
    if (isAdmin) {
      cardCols.push({ key: 'user_id', label: 'Владелец (ID)' });
    }
    cardCols.push(
      { key: 'key_id', label: 'ID ключа' },
      {
        key: 'key',
        label: 'Ключ',
        render: (r) => r.key?.name || '—',
      }
    );

    const map = {
      cards: cardCols,
      terminals: [
        { key: 'id', label: 'ID' },
        { key: 'serial_number', label: 'Серийный №' },
        { key: 'name', label: 'Название' },
        { key: 'address', label: 'Адрес' },
        { key: 'description', label: 'Описание' },
      ],
      transactions: [
        { key: 'id', label: 'ID' },
        { key: 'amount', label: 'Сумма' },
        {
          key: 'card',
          label: 'Карта',
          render: (r) => r.card?.number ?? r.card_id,
        },
        {
          key: 'terminal',
          label: 'Терминал',
          render: (r) => r.terminal?.serial_number ?? r.terminal_id,
        },
        { key: 'status', label: 'Статус' },
        { key: 'message', label: 'Сообщение' },
        {
          key: 'created_at',
          label: 'Время',
          render: (r) => fmtDate(r.created_at),
        },
      ],
      keys: [
        { key: 'id', label: 'ID' },
        { key: 'name', label: 'Название' },
        {
          key: 'key_value',
          label: 'Значение',
          render: (r) => (
            <code className="mono">{r.key_value?.length > 32 ? `${r.key_value.slice(0, 32)}…` : r.key_value}</code>
          ),
        },
        { key: 'description', label: 'Описание' },
      ],
      users: [
        { key: 'id', label: 'ID' },
        { key: 'login', label: 'Логин' },
        { key: 'name', label: 'Имя' },
        {
          key: 'is_admin',
          label: 'Админ',
          render: (r) => (r.is_admin ? 'да' : 'нет'),
        },
        {
          key: 'created_at',
          label: 'Создан',
          render: (r) => fmtDate(r.created_at),
        },
      ],
      profile: [
        { key: 'id', label: 'ID' },
        { key: 'login', label: 'Логин' },
        { key: 'name', label: 'Имя' },
        {
          key: 'is_admin',
          label: 'Роль',
          render: () => 'Пользователь',
        },
        {
          key: 'created_at',
          label: 'Создан',
          render: (r) => fmtDate(r.created_at),
        },
      ],
    };
    return map[tab] || [];
  }, [tab, isAdmin]);
}
