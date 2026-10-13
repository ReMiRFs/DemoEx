import { Routes } from '@angular/router';
import { authGuard, adminGuard, guestGuard } from './guards';
import { RegisterPage } from './pages/register';
import { LoginPage } from './pages/login';
import { RequestsPage } from './pages/requests';
import { RequestFormPage } from './pages/request-form';
import { AdminPage } from './pages/admin';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'requests' },
  { path: 'register', component: RegisterPage, canActivate: [guestGuard] },
  { path: 'login', component: LoginPage, canActivate: [guestGuard] },
  { path: 'requests', component: RequestsPage, canActivate: [authGuard] },
  { path: 'requests/new', component: RequestFormPage, canActivate: [authGuard] },
  { path: 'requests/:id/edit', component: RequestFormPage, canActivate: [authGuard] },
  { path: 'admin', component: AdminPage, canActivate: [adminGuard] },
  { path: '**', redirectTo: 'requests' },
];
