import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from './auth.service';

// Guard скрывает страницы в интерфейсе. Доступ к данным всегда проверяет Go.
export const authGuard: CanActivateFn = async () => {
  const router = inject(Router);
  const user = await inject(AuthService).restoreSession();
  return user ? true : router.createUrlTree(['/login']);
};

export const adminGuard: CanActivateFn = async () => {
  const router = inject(Router);
  const user = await inject(AuthService).restoreSession();
  if (!user) return router.createUrlTree(['/login']);
  return user.role === 'admin' ? true : router.createUrlTree(['/requests']);
};

export const guestGuard: CanActivateFn = async () => {
  const router = inject(Router);
  const user = await inject(AuthService).restoreSession();
  return user ? router.createUrlTree([user.role === 'admin' ? '/admin' : '/requests']) : true;
};
