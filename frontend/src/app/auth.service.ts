import { inject, Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError, firstValueFrom, of, tap } from 'rxjs';
import { Registration, User } from './models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  readonly user = signal<User | null>(null);
  private loaded = false;

  // Cookie сессии отправляется браузером автоматически на тот же адрес /api.
  async restoreSession(): Promise<User | null> {
    if (this.loaded) return this.user();
    return firstValueFrom(this.http.get<User>('/api/auth/me').pipe(
      catchError(() => of(null)),
      tap(user => {
        this.user.set(user);
        this.loaded = true;
      }),
    ));
  }

  login(login: string, password: string) {
    return this.http.post<User>('/api/auth/login', { login, password }).pipe(
      tap(user => { this.user.set(user); this.loaded = true; }),
    );
  }

  register(data: Registration) {
    return this.http.post<User>('/api/auth/register', data).pipe(
      tap(user => { this.user.set(user); this.loaded = true; }),
    );
  }

  logout() {
    return this.http.post<void>('/api/auth/logout', {}).pipe(
      tap(() => this.user.set(null)),
    );
  }
}
