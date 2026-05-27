import { useMemo } from 'react';
import {
  ArrowLeftRight,
  CreditCard,
  KeyRound,
  Smartphone,
  UserCircle,
  Users,
} from 'lucide-react';

export function useMainTabs(isAdmin) {
  return useMemo(() => {
    const base = [
      { id: 'cards', label: 'Карты', icon: CreditCard, adminOnly: false },
      { id: 'terminals', label: 'Терминалы', icon: Smartphone, adminOnly: false },
      { id: 'transactions', label: 'Транзакции', icon: ArrowLeftRight, adminOnly: false },
      { id: 'keys', label: 'Ключи MIFARE', icon: KeyRound, adminOnly: true },
      { id: 'users', label: 'Пользователи', icon: Users, adminOnly: true },
    ];
    const visible = base.filter((t) => !t.adminOnly || isAdmin);
    if (!isAdmin) {
      visible.push({ id: 'profile', label: 'Профиль', icon: UserCircle, adminOnly: false });
    }
    return visible;
  }, [isAdmin]);
}
